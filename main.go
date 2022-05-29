package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/crqra/go-action/pkg/action"
	"github.com/google/go-github/v43/github"
)

const appName = "GitHub"
const appSource = "jira-remote-link-action"

const iconOpened = "https://raw.githubusercontent.com/carlsberg/jira-remote-link-action/main/assets/opened.png?raw=true"
const iconClosed = "https://raw.githubusercontent.com/carlsberg/jira-remote-link-action/main/assets/closed.png?raw=true"
const iconReopened = "https://raw.githubusercontent.com/carlsberg/jira-remote-link-action/main/assets/reopened.png?raw=true"

var pat = regexp.MustCompile(`[A-Z]{1,5}-[0-9]+`)

type JiraRemoteLinkAction struct {
	JiraUrl   string `action:"jira-url"`
	JiraEmail string `action:"jira-email"`
	JiraToken string `action:"jira-token"`
}

func (a *JiraRemoteLinkAction) Run() error {

	evt, err := action.GetEvent()
	if err != nil {
		return err
	}

	switch evt := evt.(type) {
	case *github.IssuesEvent:
		if evt.GetAction() == "created" || evt.GetAction() == "edited" {

			keys := append(findJiraKeys(evt.Issue.GetBody()), findJiraKeys(evt.Issue.GetTitle())...)

			for _, key := range keys {

				aEnc := getJiraAuth(a.JiraEmail, a.JiraToken)

				issueNumber := evt.Issue.GetNumber()

				repo := evt.Repo.GetFullName()
				title := evt.Issue.GetTitle()

				globalId := fmt.Sprintf("source=%s-%s&repo=%s&issue=%d", appName, appSource, repo, issueNumber)
				linkTitle := fmt.Sprintf("%s (%s#%d)", title, repo, issueNumber)

				status, err := getStatus(*evt)
				if err != nil {
					action.SetFailed(err, map[string]string{})
				}

				body := JiraRequestBody{
					GlobalId:    globalId,
					Application: Application{Name: appName},
					Object: Object{
						Url:   evt.Issue.GetHTMLURL(),
						Title: linkTitle,
						Icon: Icon{
							Title: status.title,
							Url:   status.icon,
						},
						Status: Status{
							Icon: Icon{
								Title: status.title,
								Url:   status.icon,
							},
							Resolved: status.resolved,
						},
					},
					Relationship: "links to",
				}

				buf := bytes.Buffer{}
				jsonEnc := json.NewEncoder(&buf)
				err = jsonEnc.Encode(body)
				if err != nil {
					action.SetFailed(err, map[string]string{})
				}

				c := http.Client{Timeout: time.Duration(3) * time.Second}
				url := fmt.Sprintf("https://%s/rest/api/3/issue/%s/remotelink", a.JiraUrl, key)

				req, err := http.NewRequest(http.MethodPost, url, &buf)
				if err != nil {
					action.SetFailed(err, map[string]string{})
				}

				req.Header.Add("authorization", fmt.Sprintf("Basic %s", aEnc))
				req.Header.Add("Accept", "application/json")
				req.Header.Add("Content-Type", "application/json")

				_, err = c.Do(req)
				if err != nil {
					action.SetFailed(err, map[string]string{})
				}
			}

		} else if evt.GetAction() == "deleted" {
			//remove link
		}

		return nil

	default:
		action.Notice(
			"jira-remote-link-action skipped: only runs for pull_request and push events",
			map[string]string{},
		)
	}

	return nil
}

func getJiraAuth(jiraEmail, jiraToken string) string {
	jiraAuth := fmt.Sprintf("%s:%s", jiraEmail, jiraToken)
	aEnc := base64.StdEncoding.EncodeToString([]byte(jiraAuth))
	return aEnc
}

type JiraRequestBody struct {
	GlobalId     string      `json:"globalId"`
	Application  Application `json:"application"`
	Object       Object      `json:"jira-token"`
	Relationship string      `json:"relationship"`
}

type Application struct {
	Name string `json:"name"`
}

type Object struct {
	Url    string `json:"url"`
	Title  string `json:"title"`
	Icon   Icon   `json:"icon"`
	Status Status `json:"status"`
}

type Icon struct {
	Title string `json:"title"`
	Url   string `json:"url16x16"`
}
type Status struct {
	Icon     Icon `json:"icon"`
	Resolved bool `json:"resolved"`
}

type TempStatus struct {
	title    string
	resolved bool
	icon     string
}

func getStatus(i github.IssuesEvent) (TempStatus, error) {
	if i.GetAction() == "reopened" {
		return TempStatus{"Reopened", false, iconReopened}, nil
	}

	if i.Issue.GetState() == "open" {
		return TempStatus{"Opened", false, iconOpened}, nil
	}

	if i.Issue.GetState() == "closed" {
		return TempStatus{"Closed", true, iconClosed}, nil
	}
	return TempStatus{}, fmt.Errorf("get status: couldn't detect status from event")
}

func findJiraKeys(text string) []string {
	return pat.FindAllString(text, -1)
}

func main() {
	if err := action.Execute(&JiraRemoteLinkAction{}); err != nil {
		action.SetFailed(err, map[string]string{})
	}
}
