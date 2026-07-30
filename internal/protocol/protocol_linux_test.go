//go:build linux

package protocol

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureRegistersPerUserLinuxHandler(t *testing.T) {
	home := t.TempDir()
	bin := filepath.Join(home, "bin")
	if err := os.MkdirAll(bin, 0o700); err != nil {
		t.Fatal(err)
	}
	xdgMime := filepath.Join(bin, "xdg-mime")
	script := "#!/bin/sh\nprintf '%s\\n' \"$*\" > \"$HOME/xdg-mime.args\"\n"
	if err := os.WriteFile(xdgMime, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	t.Setenv("PATH", bin)

	result := ensure()
	if !result.Supported || !result.Ready || result.Message != "" {
		t.Fatalf("ensure() = %#v", result)
	}

	desktopPath := filepath.Join(
		home,
		".local",
		"share",
		"applications",
		"chat2api-session-connector.desktop",
	)
	desktop, err := os.ReadFile(desktopPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(desktop), "MimeType=x-scheme-handler/chat2api-connector;") {
		t.Fatalf("desktop entry = %s", desktop)
	}
	if !strings.Contains(string(desktop), "Exec=\""+home+"/.local/lib/chat2api-session-connector/chat2api-connector\" %u") {
		t.Fatalf("desktop entry = %s", desktop)
	}

	args, err := os.ReadFile(filepath.Join(home, "xdg-mime.args"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(args)) !=
		"default chat2api-session-connector.desktop x-scheme-handler/chat2api-connector" {
		t.Fatalf("xdg-mime args = %q", args)
	}
}
