//go:build !windows

package browser

func defaultBrowserPreference() string {
	return ""
}
