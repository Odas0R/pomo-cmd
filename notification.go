package pomo

import (
	"fmt"
	"os/exec"
)

const appID = "Microsoft.WSL"

func sendNotification(title, message string) error {
	psScript := fmt.Sprintf(`New-BurntToastNotification -Text "%s", "%s" -AppId %s`, title, message, appID)
	cmd := exec.Command("pwsh.exe", "-Command", psScript)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("notification error: %v, output: %s", err, string(output))
	}
	return nil
}

func playAlert() error {
	psScript := `[Console]::Beep(800, 500)` // frequency: 800Hz, duration: 500ms
	cmd := exec.Command("pwsh.exe", "-Command", psScript)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("alert sound error: %v, output: %s", err, string(output))
	}
	return nil
}
