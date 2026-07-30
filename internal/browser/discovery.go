package browser

import (
	"os"
	"os/exec"
	"runtime"
)

type candidate struct {
	Name   string
	Path   string
	Lookup bool
}

func candidates() []candidate {
	switch runtime.GOOS {
	case "windows":
		return []candidate{
			{Name: "Microsoft Edge", Path: `%PROGRAMFILES(X86)%\Microsoft\Edge\Application\msedge.exe`},
			{Name: "Microsoft Edge", Path: `%PROGRAMFILES%\Microsoft\Edge\Application\msedge.exe`},
			{Name: "Google Chrome", Path: `%PROGRAMFILES%\Google\Chrome\Application\chrome.exe`},
			{Name: "Google Chrome", Path: `%PROGRAMFILES(X86)%\Google\Chrome\Application\chrome.exe`},
			{Name: "Google Chrome", Path: `%LOCALAPPDATA%\Google\Chrome\Application\chrome.exe`},
			{Name: "Brave", Path: `%PROGRAMFILES%\BraveSoftware\Brave-Browser\Application\brave.exe`},
			{Name: "Brave", Path: `%LOCALAPPDATA%\BraveSoftware\Brave-Browser\Application\brave.exe`},
		}
	case "darwin":
		return []candidate{
			{Name: "Google Chrome", Path: "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"},
			{Name: "Microsoft Edge", Path: "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge"},
			{Name: "Brave", Path: "/Applications/Brave Browser.app/Contents/MacOS/Brave Browser"},
			{Name: "Chromium", Path: "/Applications/Chromium.app/Contents/MacOS/Chromium"},
			{Name: "Google Chrome", Path: "$HOME/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"},
		}
	default:
		return []candidate{
			{Name: "Google Chrome", Path: "google-chrome", Lookup: true},
			{Name: "Google Chrome", Path: "google-chrome-stable", Lookup: true},
			{Name: "Microsoft Edge", Path: "microsoft-edge", Lookup: true},
			{Name: "Microsoft Edge", Path: "microsoft-edge-stable", Lookup: true},
			{Name: "Brave", Path: "brave-browser", Lookup: true},
			{Name: "Chromium", Path: "chromium", Lookup: true},
			{Name: "Chromium", Path: "chromium-browser", Lookup: true},
		}
	}
}

func lookupExecutable(name string) (string, error) {
	return exec.LookPath(name)
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}
