package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/syohex/cmds"
)

var dryRun = false
var ticketRegex = regexp.MustCompile(`[A-Z]+-([0-9])+`)

func init() {
	flag.BoolVar(&dryRun, "n", false, "dry run")
	flag.Parse()
}

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

	matches := ticketRegex.FindAllString(branch, -1)
	if len(matches) == 0 {
		return "", fmt.Errorf("ticket name is not found in branch name(%s)", branch)
	}

	return matches[0], nil
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
	if dryRun {
		fmt.Println(url)
		return 0
	}

	if err := cmds.OpenBrowser(url); err != nil {
		fmt.Fprintf(os.Stderr, "failed to '%s' browser: %v\n", url, err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(_main())
}
