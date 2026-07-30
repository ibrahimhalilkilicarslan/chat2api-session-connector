//go:build windows

package browser

import "golang.org/x/sys/windows/registry"

const windowsHTTPSUserChoice = `Software\Microsoft\Windows\Shell\Associations\UrlAssociations\https\UserChoice`

func defaultBrowserPreference() string {
	key, err := registry.OpenKey(registry.CURRENT_USER, windowsHTTPSUserChoice, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer key.Close()

	programID, _, err := key.GetStringValue("ProgId")
	if err != nil {
		return ""
	}
	return browserNameForProgID(programID)
}
