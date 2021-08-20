# Jira Remote Link Action

This action creates [remote links](https://developer.atlassian.com/server/jira/platform/creating-remote-issue-links/) in Jira from the issue keys found in GitHub Issues.

## Usage

```yaml
- uses: carlsberg/jira-remote-link-action@v1
  with:
    # URL for Jira instance (i.e. `example.atlassian.net`)
    jira-url: ${{ secrets.JIRA_URL }}

    # Email for Jira account (i.e. `john.doe@example.com`)
    jira-email: ${{ secrets.JIRA_EMAIL }}
    
    # Token for Jira account
    jira-token: ${{ secrets.JIRA_TOKEN }}
```

## License

This project is released under the [MIT License](LICENSE).