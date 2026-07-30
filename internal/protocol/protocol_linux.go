//go:build linux

package protocol

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ensure() Result {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return Result{Supported: true, Message: "Linux kullanıcı dizini bulunamadı."}
	}
	executable := filepath.Join(home, ".local", "lib", "chat2api-session-connector", "chat2api-connector")
	if strings.ContainsAny(executable, "\r\n") {
		return Result{Supported: true, Message: "Linux kullanıcı dizini güvenli bir uygulama yolu oluşturmuyor."}
	}
	if err := installCurrentExecutable(executable); err != nil {
		return Result{Supported: true, Message: "Connector kullanıcı dizinine kurulamadı."}
	}

	applications := filepath.Join(home, ".local", "share", "applications")
	if err := os.MkdirAll(applications, 0o700); err != nil {
		return Result{Supported: true, Message: "Masaüstü uygulama dizini oluşturulamadı."}
	}
	desktopPath := filepath.Join(applications, "chat2api-session-connector.desktop")
	desktop := strings.Join([]string{
		"[Desktop Entry]",
		"Name=Chat2API Session Connector",
		"Comment=Link a DeepSeek session to Chat2API",
		"Type=Application",
		"Terminal=false",
		"NoDisplay=true",
		"Exec=" + desktopQuote(executable) + " %u",
		"MimeType=x-scheme-handler/chat2api-connector;",
		"",
	}, "\n")
	if err := os.WriteFile(desktopPath, []byte(desktop), 0o600); err != nil {
		return Result{Supported: true, Message: "Connector masaüstü kaydı yazılamadı."}
	}
	xdgMime, err := exec.LookPath("xdg-mime")
	if err != nil {
		return Result{
			Supported: true,
			Message:   "xdg-mime bulunamadı; manuel bağlantı kodu kullanılabilir.",
		}
	}
	command := exec.Command(xdgMime, "default", filepath.Base(desktopPath), "x-scheme-handler/chat2api-connector")
	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if err := command.Run(); err != nil {
		return Result{
			Supported: true,
			Message:   "Connector bağlantı protokolü kaydedilemedi; manuel kod kullanılabilir.",
		}
	}
	return Result{Supported: true, Ready: true}
}

func desktopQuote(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	value = strings.ReplaceAll(value, "`", "\\`")
	value = strings.ReplaceAll(value, "$", "\\$")
	return "\"" + value + "\""
}
