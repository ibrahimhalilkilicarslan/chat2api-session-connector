package browser

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/diagnostic"
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
		return Installed{}, diagnostic.New(
			"BROWSER_PATH_INVALID",
			"Seçilen tarayıcı çalıştırılamıyor.",
			"CHAT2API_BROWSER_PATH değerini Chrome, Edge, Chromium veya Brave çalıştırılabilir dosyasına yönlendirin.",
		)
	}

	for _, candidate := range prioritizeCandidates(candidates(), defaultBrowserPreference()) {
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
	return Installed{}, diagnostic.New(
		"BROWSER_NOT_FOUND",
		"Desteklenen bir tarayıcı bulunamadı.",
		"Chrome, Edge, Chromium veya Brave kurun; ardından connector'ı yeniden açın.",
	)
}

func prioritizeCandidates(values []candidate, preferredName string) []candidate {
	preferredName = strings.TrimSpace(preferredName)
	if preferredName == "" {
		return values
	}

	prioritized := make([]candidate, 0, len(values))
	for _, value := range values {
		if value.Name == preferredName {
			prioritized = append(prioritized, value)
		}
	}
	for _, value := range values {
		if value.Name != preferredName {
			prioritized = append(prioritized, value)
		}
	}
	return prioritized
}

func browserNameForProgID(programID string) string {
	normalized := strings.ToLower(strings.TrimSpace(programID))
	switch {
	case strings.Contains(normalized, "chromehtml"):
		return "Google Chrome"
	case strings.Contains(normalized, "msedgehtm"),
		strings.Contains(normalized, "microsoftedge"):
		return "Microsoft Edge"
	case strings.Contains(normalized, "bravehtml"):
		return "Brave"
	case strings.Contains(normalized, "chromium"):
		return "Chromium"
	default:
		return ""
	}
}

func CaptureToken(ctx context.Context, installed Installed) (string, error) {
	if !isExecutable(installed.Path) {
		return "", diagnostic.New(
			"BROWSER_NOT_EXECUTABLE",
			installed.Name+" çalıştırılamıyor.",
			"Tarayıcı kurulumunu onarın veya başka bir Chromium tabanlı tarayıcı kurun.",
		)
	}
	profileDir, err := os.MkdirTemp("", "chat2api-deepseek-profile-*")
	if err != nil {
		return "", diagnostic.New(
			"TEMP_PROFILE_CREATE_FAILED",
			"Güvenli geçici tarayıcı profili oluşturulamadı.",
			"Disk alanını ve işletim sistemi geçici klasör izinlerini kontrol edin.",
		)
	}
	if err := os.Chmod(profileDir, 0o700); err != nil && runtime.GOOS != "windows" {
		_ = os.RemoveAll(profileDir)
		return "", diagnostic.New(
			"TEMP_PROFILE_PROTECTION_FAILED",
			"Geçici tarayıcı profili güvenli izinlerle açılamadı.",
			"Geçici klasör izinlerini kontrol edip connector'ı yeniden başlatın.",
		)
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
		return "", diagnostic.New(
			"BROWSER_LAUNCH_FAILED",
			"DeepSeek giriş penceresi açılamadı.",
			installed.Name+" kurulumunu ve güvenlik yazılımı engellerini kontrol edip yeniden deneyin. Sorun sürerse connector doctor komutunu çalıştırın.",
		)
	}

	ticker := time.NewTicker(900 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", diagnostic.New(
				"LOGIN_TIMEOUT",
				"DeepSeek girişi zamanında tamamlanamadı.",
				"Chat2API'den yeni bir bağlantı başlatın ve açılan penceredeki giriş ile doğrulamayı tamamlayın.",
			)
		case <-ticker.C:
			token, err := readToken(browserContext)
			if err != nil {
				if browserContext.Err() != nil {
					return "", diagnostic.New(
						"BROWSER_CLOSED",
						"DeepSeek bağlantısı tamamlanmadan tarayıcı penceresi kapandı.",
						"Yeni bir bağlantı başlatın ve hesap bağlanana kadar açılan pencereyi kapatmayın.",
					)
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
		return "", diagnostic.New(
			"TOKEN_SIZE_INVALID",
			"DeepSeek oturum bilgisi beklenen güvenli sınırı aşıyor.",
			"DeepSeek oturumunu kapatıp yeni bir oturumla tekrar deneyin.",
		)
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
