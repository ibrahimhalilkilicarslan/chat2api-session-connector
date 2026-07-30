package browser

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const deepSeekURL = "https://chat.deepseek.com/"

type Installed struct {
	Name string
	Path string
}

func Find() (Installed, error) {
	if custom := strings.TrimSpace(os.Getenv("CHAT2API_BROWSER_PATH")); custom != "" {
		if isExecutable(custom) {
			return Installed{Name: "Custom Chromium", Path: custom}, nil
		}
		return Installed{}, errors.New("CHAT2API_BROWSER_PATH çalıştırılabilir bir tarayıcı göstermiyor")
	}

	for _, candidate := range candidates() {
		path := expand(candidate.Path)
		if candidate.Lookup {
			resolved, err := lookupExecutable(path)
			if err == nil {
				return Installed{Name: candidate.Name, Path: resolved}, nil
			}
			continue
		}
		if isExecutable(path) {
			return Installed{Name: candidate.Name, Path: path}, nil
		}
	}
	return Installed{}, errors.New("desteklenen Chrome, Edge, Chromium veya Brave tarayıcısı bulunamadı")
}

func CaptureToken(ctx context.Context, installed Installed) (string, error) {
	if !isExecutable(installed.Path) {
		return "", errors.New("seçilen tarayıcı çalıştırılamıyor")
	}
	profileDir, err := os.MkdirTemp("", "chat2api-deepseek-profile-*")
	if err != nil {
		return "", errors.New("geçici tarayıcı profili oluşturulamadı")
	}
	if err := os.Chmod(profileDir, 0o700); err != nil && runtime.GOOS != "windows" {
		_ = os.RemoveAll(profileDir)
		return "", errors.New("geçici tarayıcı profili korunamadı")
	}
	defer removeProfile(profileDir)

	options := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(installed.Path),
		chromedp.UserDataDir(profileDir),
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("remote-debugging-address", "127.0.0.1"),
		chromedp.Flag("remote-debugging-port", 0),
		chromedp.Flag("window-size", "1180,820"),
	)
	allocatorContext, cancelAllocator := chromedp.NewExecAllocator(ctx, options...)
	browserContext, cancelBrowser := chromedp.NewContext(allocatorContext)

	defer func() {
		cancelBrowser()
		cancelAllocator()
	}()

	if err := chromedp.Run(browserContext, chromedp.Navigate(deepSeekURL)); err != nil {
		return "", errors.New("DeepSeek giriş sayfası açılamadı")
	}

	ticker := time.NewTicker(900 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", errors.New("DeepSeek giriş süresi doldu veya işlem iptal edildi")
		case <-ticker.C:
			token, err := readToken(browserContext)
			if err != nil {
				if browserContext.Err() != nil {
					return "", errors.New("tarayıcı bağlantısı kapandı")
				}
				continue
			}
			if token != "" {
				return token, nil
			}
		}
	}
}

func readToken(ctx context.Context) (string, error) {
	const expression = `(() => {
		if (globalThis.location?.origin !== "https://chat.deepseek.com") return "";
		const token = globalThis.localStorage?.getItem("userToken");
		return typeof token === "string" ? token.trim() : "";
	})()`
	var token string
	if err := chromedp.Run(ctx, chromedp.Evaluate(expression, &token)); err != nil {
		return "", err
	}
	if len(token) > 16_384 {
		return "", fmt.Errorf("DeepSeek oturum bilgisi beklenen sınırı aşıyor")
	}
	return token, nil
}

func removeProfile(path string) {
	for attempt := 0; attempt < 8; attempt++ {
		if err := os.RemoveAll(path); err == nil {
			return
		}
		time.Sleep(time.Duration(attempt+1) * 125 * time.Millisecond)
	}
}

func expand(path string) string {
	replacer := strings.NewReplacer(
		"%LOCALAPPDATA%", os.Getenv("LOCALAPPDATA"),
		"%PROGRAMFILES%", os.Getenv("PROGRAMFILES"),
		"%PROGRAMFILES(X86)%", os.Getenv("PROGRAMFILES(X86)"),
		"$HOME", userHome(),
	)
	return filepath.Clean(replacer.Replace(path))
}

func userHome() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}
