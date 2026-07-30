//go:build windows

package protocol

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const windowsRegistryPath = `Software\Classes\chat2api-connector`

func ensure() Result {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return Result{
			Supported: true,
			Message:   "Windows kullanıcı uygulama dizini bulunamadı; manuel bağlantı kodu kullanılabilir.",
		}
	}
	executable := filepath.Join(localAppData, "Chat2API", "SessionConnector", "Chat2API-Connector.exe")
	if err := installCurrentExecutable(executable); err != nil {
		return Result{
			Supported: true,
			Message:   "Connector kullanıcı dizinine kurulamadı; manuel bağlantı kodu kullanılabilir.",
		}
	}

	root, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		windowsRegistryPath,
		registry.SET_VALUE|registry.CREATE_SUB_KEY,
	)
	if err != nil {
		return Result{Supported: true, Message: "Connector bağlantı protokolü kaydedilemedi."}
	}
	if err := root.SetStringValue("", "URL:Chat2API Session Connector"); err != nil {
		root.Close()
		return Result{Supported: true, Message: "Connector bağlantı protokolü kaydedilemedi."}
	}
	if err := root.SetStringValue("URL Protocol", ""); err != nil {
		root.Close()
		return Result{Supported: true, Message: "Connector bağlantı protokolü kaydedilemedi."}
	}
	root.Close()

	icon, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		windowsRegistryPath+`\DefaultIcon`,
		registry.SET_VALUE,
	)
	if err != nil {
		return Result{Supported: true, Message: "Connector bağlantı protokolü tamamlanamadı."}
	}
	_ = icon.SetStringValue("", executable+",0")
	icon.Close()

	command, _, err := registry.CreateKey(
		registry.CURRENT_USER,
		windowsRegistryPath+`\shell\open\command`,
		registry.SET_VALUE,
	)
	if err != nil {
		return Result{Supported: true, Message: "Connector açılış komutu kaydedilemedi."}
	}
	err = command.SetStringValue("", `"`+executable+`" "%1"`)
	command.Close()
	if err != nil {
		return Result{Supported: true, Message: "Connector açılış komutu kaydedilemedi."}
	}
	return Result{Supported: true, Ready: true}
}
