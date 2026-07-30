package browser

import (
	"os/exec"
	"runtime"

	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/diagnostic"
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
		return diagnostic.New(
			"LOCAL_UI_OPEN_FAILED",
			"Connector arayüzü varsayılan tarayıcıda açılamadı.",
			"Varsayılan tarayıcı ayarını kontrol edip connector'ı yeniden açın.",
		)
	}
	return nil
}
