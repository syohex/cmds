package cmds

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const pwshPath = "/mnt/c/Program Files/PowerShell/7/pwsh.exe"

func IsWSL() bool {
	_, ok := os.LookupEnv("WSLENV")
	return ok
}

func wslOpenUrlByPowerShell(url string) error {
	if _, err := os.Stat(pwshPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("Powershell 7 is not installed")
	}

	// pwsh.exe -NoProfile -Command "Start-Process 'URL'"
	args := []string{
		"-NoProfile",
		"-Command",
		fmt.Sprintf("Start-Process '%s'", url),
	}
	cmd := exec.Command(pwshPath, args...)
	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

func translateToWindowsPath(path string) (string, error) {
	cmd := exec.Command("wslpath", "-w", path)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func WSLOpen(url string) error {
	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		return wslOpenUrlByPowerShell(url)
	}

	path, err := translateToWindowsPath(url)
	if err != nil {
		return nil
	}

	return wslOpenUrlByPowerShell(path)
}
