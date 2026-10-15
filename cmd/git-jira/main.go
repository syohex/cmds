package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
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

func isWSL() bool {
	return os.Getenv("WSL_INTEROP") != ""
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd.exe"
		args = []string{"/c", "rundll32", "url.dll,FileProtocolHandler", strings.ReplaceAll(url, "&", "^&")}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		args = []string{url}
		if isWSL() {
			cmd = "/mnt/c/Windows/explorer.exe"
		} else {
			cmd = "xdg-open"
		}
	}

	c := exec.Command(cmd, args...)
	if err := c.Start(); err != nil {
		return err
	}

	if err := c.Wait(); err != nil {
		return err
	}

	return nil
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
	if err := openBrowser(url); err != nil {
		fmt.Fprintf(os.Stderr, "failed to '%s' browser: %v\n", url, err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(_main())
}
