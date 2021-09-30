const core = require("@actions/core");
const github = require("@actions/github");
const fetch = require("node-fetch");

const APP_NAME = "GitHub";
const APP_SOURCE = "jira-remote-link-action";

const JIRA_URL = core.getInput("jira-url", { required: true });
const JIRA_EMAIL = core.getInput("jira-email", { required: true });
const JIRA_TOKEN = core.getInput("jira-token", { required: true });
const JIRA_AUTH = Buffer.from(`${JIRA_EMAIL}:${JIRA_TOKEN}`).toString("base64");

const ICON_OPENED =
  "https://raw.githubusercontent.com/carlsberg/jira-remote-link-action/main/assets/opened.png";

const ICON_CLOSED =
  "https://raw.githubusercontent.com/carlsberg/jira-remote-link-action/main/assets/closed.png";

const ICON_REOPENED =
  "https://raw.githubusercontent.com/carlsberg/jira-remote-link-action/main/assets/reopened.png";

main();

async function main() {
  try {
    for (const jiraKey of findJiraKeys()) {
      await createRemoteLink(jiraKey);
    }
  } catch (error) {
    core.debug(JSON.stringify(error, null, 2));
    core.setFailed(error.message);
  }
}

function createRemoteLink(jiraKey) {
  const {
    issue: { html_url: url },
  } = github.context.payload;

  const status = getStatus();
  const title = buildTitle();
  const globalId = buildGlobalId();

  return fetch(`https://${JIRA_URL}/rest/api/3/issue/${jiraKey}/remotelink`, {
    method: "POST",
    headers: {
      Authorization: `Basic ${JIRA_AUTH}`,
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      globalId,
      application: { name: APP_NAME },
      object: {
        url,
        title,
        icon: {
          title: status.title,
          url16x16: status.icon,
        },
        status: {
          icon: {
            title: status.title,
            url16x16: status.icon,
          },
          resolved: status.resolved,
        },
      },
      relationship: "links to",
    }),
  });
}

function getStatus() {
  const { state } = github.context.payload.issue;

  const status =
    github.context.payload.action === "reopened"
      ? "reopened"
      : state === "open"
      ? "opened"
      : state === "closed"
      ? "closed"
      : false;

  switch (status) {
    case "opened":
      return { title: "Opened", resolved: false, icon: ICON_OPENED };

    case "closed":
      return { title: "Closed", resolved: true, icon: ICON_CLOSED };

    case "reopened":
      return { title: "Reopened", resolved: false, icon: ICON_REOPENED };

    default:
      core.setFailed("Couldn't detect status from event");
  }
}

function buildGlobalId() {
  const {
    repository: { full_name: repo },
    issue: { number: issueNumber },
  } = github.context.payload;

  return `source=${APP_NAME}-${APP_SOURCE}&repo=${repo}&issue=${issueNumber}`;
}

function buildTitle() {
  const {
    repository: { full_name: repo },
    issue: { number: issueNumber, title },
  } = github.context.payload;

  return `${title} (${repo}#${issueNumber})`;
}

function findJiraKeys() {
  const keys = [];
  const pattern = new RegExp(/([A-Z]+-[0-9]+)/g);

  const titleMatches = github.context.payload.issue.title.match(pattern);

  if (titleMatches) {
    titleMatches.map((k) => keys.push(k));
  }

  const bodyMatches = github.context.payload.issue.body.match(pattern);

  if (bodyMatches) {
    bodyMatches.map((k) => keys.push(k));
  }

  if (github.context.payload.comment) {
    const commentMatches = github.context.payload.comment.body.match(pattern);

    if (commentMatches) {
      commentMatches.push((k) => keys.push(k));
    }
  }

  return keys.filter((el, i, arr) => arr.indexOf(el) === i);
}
