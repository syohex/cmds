package cmds

import (
	"os/exec"
	"runtime"
	"strings"
)

func escapeForCommandPrompt(s string) string {
	return strings.ReplaceAll(s, "&", "^&")
}

func OpenBrowser(url string) error {
	if IsWSL() {
		return WSLOpen(url)
	}

	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd.exe"
		args = []string{"/c", "rundll32", "url.dll,FileProtocolHandler", escapeForCommandPrompt(url)}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		args = []string{url}
		cmd = "xdg-open"
	}

	c := exec.Command(cmd, args...)
	if err := c.Start(); err != nil {
		return err
	}

	return c.Wait()
}
