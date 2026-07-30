package app

import (
	"context"
	"fmt"
	"io"
	"runtime"

	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/browser"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/gateway"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/pairing"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/ui"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/version"
)

func Run(ctx context.Context, args []string, stdout io.Writer) error {
	if len(args) > 0 {
		switch args[0] {
		case "--version", "version":
			_, err := fmt.Fprintf(stdout, "chat2api-session-connector %s (%s)\n", version.Version, version.Commit)
			return err
		case "doctor":
			return doctor(stdout)
		default:
			return fmt.Errorf("desteklenmeyen komut: %s", args[0])
		}
	}

	client := gateway.New(version.Version)
	connect := func(connectContext context.Context, payload pairing.Payload) (string, error) {
		installed, err := browser.Find()
		if err != nil {
			return "", err
		}
		token, err := browser.CaptureToken(connectContext, installed)
		if err != nil {
			return installed.Name, err
		}
		if err := client.Complete(connectContext, payload, token); err != nil {
			return installed.Name, err
		}
		return installed.Name, nil
	}
	return ui.Run(ctx, connect, browser.OpenURL)
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
