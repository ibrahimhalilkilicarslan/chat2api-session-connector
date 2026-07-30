package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

func OpenURL(rawURL string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", rawURL)
	case "darwin":
		command = exec.Command("open", rawURL)
	default:
		command = exec.Command("xdg-open", rawURL)
	}
	if err := command.Start(); err != nil {
		return fmt.Errorf("yerel arayüz varsayılan tarayıcıda açılamadı")
	}
	return nil
}
