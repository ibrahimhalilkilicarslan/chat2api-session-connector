package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	cdpRuntime "github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/diagnostic"
)

const deepSeekURL = "https://chat.deepseek.com/"

const storedTokenExpression = `(() => {
	const origin = globalThis.location?.origin;
	if (origin !== "https://chat.deepseek.com") return {state: "empty", token: ""};
	const stored = globalThis.localStorage?.getItem("userToken") ?? "";
	return stored ? {state: "stored", token: stored} : {state: "empty", token: ""};
})()`

const tokenValidationExpression = `(async (token) => {
	try {
		const response = await fetch("/api/v0/users/current", {
			method: "GET",
			cache: "no-store",
			credentials: "omit",
			headers: {
				Accept: "application/json",
				Authorization: "Bearer " + token
			}
		});
		const payload = await response.json().catch(() => null);
		const root = payload && typeof payload === "object" ? payload : {};
		const data = root.data && typeof root.data === "object" ? root.data : {};
		const businessData = data.biz_data && typeof data.biz_data === "object"
			? data.biz_data
			: root.biz_data && typeof root.biz_data === "object"
				? root.biz_data
				: null;
		if (
			response.status === 401
			|| response.status === 403
			|| root.code === 40003
			|| data.biz_code === 40003
		) {
			return {state: "rejected", token: ""};
		}
		const acceptedCodes = (root.code === undefined || root.code === 0)
			&& (data.biz_code === undefined || data.biz_code === 0);
		if (
			response.ok
			&& acceptedCodes
			&& typeof businessData?.token === "string"
			&& businessData.token.length > 0
		) {
			return {state: "ready", token};
		}
		return {state: "unavailable", token: ""};
	} catch {
		return {state: "unavailable", token: ""};
	}
})`

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

	options := launchOptions(installed, profileDir)
	allocatorContext, cancelAllocator := chromedp.NewExecAllocator(ctx, options...)
	browserContext, cancelBrowser := chromedp.NewContext(allocatorContext)

	defer func() {
		cancelBrowser()
		cancelAllocator()
	}()

	if err := chromedp.Run(browserContext); err != nil {
		return "", diagnostic.New(
			"BROWSER_LAUNCH_FAILED",
			"DeepSeek giriş penceresi açılamadı.",
			installed.Name+" kurulumunu ve güvenlik yazılımı engellerini kontrol edip yeniden deneyin. Sorun sürerse connector doctor komutunu çalıştırın.",
		)
	}
	if err := openDeepSeek(browserContext); err != nil {
		return "", diagnostic.New(
			"BROWSER_CONTROL_FAILED",
			"DeepSeek giriş penceresi hazırlanamadı.",
			"Connector'ı kapatıp yeni bir bağlantı başlatın. Sorun sürerse BROWSER_CONTROL_FAILED kodunu paylaşın.",
		)
	}

	ticker := time.NewTicker(900 * time.Millisecond)
	defer ticker.Stop()
	rejectedChecks := 0
	unavailableChecks := 0
	for {
		select {
		case <-ctx.Done():
			return "", diagnostic.New(
				"LOGIN_TIMEOUT",
				"DeepSeek girişi zamanında tamamlanamadı.",
				"Chat2API'den yeni bir bağlantı başlatın ve açılan penceredeki giriş ile doğrulamayı tamamlayın.",
			)
		case <-ticker.C:
			probe, err := readToken(browserContext)
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
			switch probe.State {
			case "ready":
				return probe.Token, nil
			case "rejected":
				rejectedChecks++
				unavailableChecks = 0
				if rejectedChecks >= 3 {
					return "", diagnostic.New(
						"SESSION_LOCAL_REJECTED",
						"DeepSeek oturumu tarayıcı tarafından doğrulanamadı.",
						"Bu pencerede DeepSeek hesabından çıkış yapıp yeniden giriş yapın; ardından Chat2API panelinden yeni bir bağlantı başlatın.",
					)
				}
			case "unavailable":
				unavailableChecks++
				rejectedChecks = 0
				if unavailableChecks >= 12 {
					return "", diagnostic.New(
						"SESSION_LOCAL_CHECK_FAILED",
						"DeepSeek oturum kontrolü tamamlanamadı.",
						"DeepSeek sayfasının tamamen açıldığını ve güvenlik yazılımının chat.deepseek.com isteklerini engellemediğini kontrol edin.",
					)
				}
			default:
				rejectedChecks = 0
				unavailableChecks = 0
			}
		}
	}
}

func openDeepSeek(ctx context.Context) error {
	var lastError error
	for attempt := 0; attempt < 3; attempt++ {
		// Login redirects may replace the first document; sending the CDP command
		// directly avoids treating that transient page-load abort as browser death.
		err := chromedp.Run(ctx, chromedp.ActionFunc(func(actionContext context.Context) error {
			_, _, _, _, commandError := page.Navigate(deepSeekURL).Do(actionContext)
			return commandError
		}))
		if err == nil {
			return nil
		}
		lastError = err
		if ctx.Err() != nil {
			return err
		}
		time.Sleep(time.Duration(attempt+1) * 200 * time.Millisecond)
	}
	return lastError
}

func launchOptions(installed Installed, profileDir string) []chromedp.ExecAllocatorOption {
	options := append(
		[]chromedp.ExecAllocatorOption{},
		chromedp.DefaultExecAllocatorOptions[:]...,
	)
	return append(
		options,
		chromedp.ExecPath(installed.Path),
		chromedp.UserDataDir(profileDir),
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("remote-debugging-address", "127.0.0.1"),
		chromedp.WindowSize(1180, 820),
	)
}

type tokenProbe struct {
	State string `json:"state"`
	Token string `json:"token"`
}

func readToken(ctx context.Context) (tokenProbe, error) {
	var stored tokenProbe
	if err := chromedp.Run(
		ctx,
		chromedp.Evaluate(
			storedTokenExpression,
			&stored,
		),
	); err != nil {
		return tokenProbe{}, err
	}

	normalized, err := normalizeStoredToken(stored)
	if err != nil || normalized.State != "stored" {
		return normalized, err
	}

	encodedToken, err := json.Marshal(normalized.Token)
	if err != nil {
		return tokenProbe{}, diagnostic.New(
			"TOKEN_INVALID",
			"DeepSeek oturum bilgisi güvenli biçimde çözümlenemedi.",
			"DeepSeek hesabından çıkış yapıp connector'ın açtığı pencerede yeniden giriş yapın.",
		)
	}
	expression := fmt.Sprintf("(%s)(%s)", tokenValidationExpression, encodedToken)
	var validated tokenProbe
	if err := chromedp.Run(
		ctx,
		chromedp.Evaluate(
			expression,
			&validated,
			func(parameters *cdpRuntime.EvaluateParams) *cdpRuntime.EvaluateParams {
				return parameters.WithAwaitPromise(true)
			},
		),
	); err != nil {
		return tokenProbe{}, err
	}
	return validated, nil
}

func normalizeStoredToken(stored tokenProbe) (tokenProbe, error) {
	if stored.State != "stored" {
		return stored, nil
	}

	raw := strings.TrimSpace(stored.Token)
	if len(raw) > 16_384 {
		return tokenProbe{}, diagnostic.New(
			"TOKEN_SIZE_INVALID",
			"DeepSeek oturum bilgisi beklenen güvenli sınırı aşıyor.",
			"DeepSeek oturumunu kapatıp yeni bir oturumla tekrar deneyin.",
		)
	}
	if raw == "" {
		return tokenProbe{State: "empty"}, nil
	}

	token := raw
	if strings.HasPrefix(raw, "{") {
		var envelope struct {
			Value string `json:"value"`
		}
		if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
			return tokenProbe{State: "rejected"}, nil
		}
		token = strings.TrimSpace(envelope.Value)
	}

	if token == "" || len(token) > 16_384 {
		return tokenProbe{State: "rejected"}, nil
	}
	if value := strings.ToLower(token); value == "null" || value == "undefined" || value == "false" {
		return tokenProbe{State: "rejected"}, nil
	}
	return tokenProbe{State: "stored", Token: token}, nil
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
