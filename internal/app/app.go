package app

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/browser"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/gateway"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/pairing"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/protocol"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/ui"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/version"
)

func Run(ctx context.Context, args []string, stdout io.Writer) error {
	standaloneLaunch := len(args) == 0
	options := ui.Options{Version: version.Version}
	if len(args) > 0 {
		switch args[0] {
		case "--version", "version":
			_, err := fmt.Fprintf(stdout, "chat2api-session-connector %s (%s)\n", version.Version, version.Commit)
			return err
		case "doctor":
			return doctor(stdout)
		default:
			if len(args) != 1 || !pairing.IsLaunchURL(args[0]) {
				return fmt.Errorf("desteklenmeyen connector komutu")
			}
			payload, err := pairing.ParseLaunchURL(args[0], time.Now())
			args[0] = ""
			if err != nil {
				options.InitialError = err
			} else {
				options.InitialPayload = &payload
			}
		}
	}
	registration := protocol.Ensure()
	options.InstallationReady = standaloneLaunch && registration.Ready
	if registration.Message != "" && options.InitialError == nil && !registration.Ready {
		options.Notice = registration.Message
	}

	client := gateway.New(version.Version)
	connect := func(
		connectContext context.Context,
		payload pairing.Payload,
		report ui.ProgressFunc,
	) (string, error) {
		report(ui.ConnectionProgress{
			Stage:   "browser-discovery",
			Message: "Desteklenen tarayıcı aranıyor.",
		})
		installed, err := browser.Find()
		if err != nil {
			return "", err
		}
		report(ui.ConnectionProgress{
			Stage:       "provider-login",
			Message:     installed.Name + " içinde güvenli DeepSeek oturumu açılıyor.",
			BrowserName: installed.Name,
		})
		token, err := browser.CaptureToken(connectContext, installed)
		if err != nil {
			return installed.Name, err
		}
		report(ui.ConnectionProgress{
			Stage:       "gateway-validation",
			Message:     "DeepSeek oturumu Chat2API gateway'inde doğrulanıyor.",
			BrowserName: installed.Name,
		})
		if err := client.Complete(connectContext, payload, token); err != nil {
			return installed.Name, err
		}
		return installed.Name, nil
	}
	return ui.Run(ctx, connect, browser.OpenURL, options)
}

func doctor(stdout io.Writer) error {
	_, _ = fmt.Fprintf(stdout, "Version: %s\n", version.Version)
	_, _ = fmt.Fprintf(stdout, "Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	installed, err := browser.Find()
	if err != nil {
		_, _ = fmt.Fprintln(stdout, "Browser: bulunamadı")
		return err
	}
	_, _ = fmt.Fprintf(stdout, "Browser: %s\n", installed.Name)
	_, _ = fmt.Fprintf(stdout, "Browser path: %s\n", installed.Path)
	return nil
}
