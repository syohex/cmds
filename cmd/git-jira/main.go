package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/syohex/cmds"
)

func getJiraUrl() (string, error) {
	url := os.Getenv("JIRA_URL")
	if url == "" {
		return "", fmt.Errorf("environment variable 'JIRA_URL' is not defined")
	}

	return url, nil
}

func getTicketName() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", nil
	}

	branch := strings.TrimSpace(string(out))
	pos := strings.LastIndex(branch, "/")
	return branch[pos+1:], nil
}

func _main() int {
	jiraBase, err := getJiraUrl()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get JIRA base URL: %v\n", err)
		return 1
	}

	ticket, err := getTicketName()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get ticket name from branch: %v\n", err)
		return 1
	}

	url := fmt.Sprintf("%s/browse/%s", jiraBase, ticket)
	if err := cmds.OpenBrowser(url); err != nil {
		fmt.Fprintf(os.Stderr, "failed to '%s' browser: %v\n", url, err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(_main())
}
