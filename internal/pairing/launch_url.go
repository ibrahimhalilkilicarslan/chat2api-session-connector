package pairing

import (
	"net/url"
	"strings"
	"time"

	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/diagnostic"
)

const (
	LaunchScheme       = "chat2api-connector"
	LaunchHost         = "pair"
	maxLaunchURLLength = 8 * 1024
)

func IsLaunchURL(raw string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(raw)), LaunchScheme+":")
}

func ParseLaunchURL(raw string, now time.Time) (Payload, error) {
	if len(raw) == 0 || len(raw) > maxLaunchURLLength {
		return Payload{}, invalidLaunchURL()
	}
	parsed, err := url.Parse(raw)
	if err != nil ||
		!strings.EqualFold(parsed.Scheme, LaunchScheme) ||
		!strings.EqualFold(parsed.Host, LaunchHost) ||
		(parsed.Path != "" && parsed.Path != "/") ||
		parsed.User != nil ||
		parsed.Fragment != "" {
		return Payload{}, invalidLaunchURL()
	}

	values, err := url.ParseQuery(parsed.RawQuery)
	if err != nil || len(values) != 1 {
		return Payload{}, invalidLaunchURL()
	}
	codes, ok := values["code"]
	if !ok || len(codes) != 1 {
		return Payload{}, invalidLaunchURL()
	}
	return Parse(codes[0], now)
}

func invalidLaunchURL() error {
	return diagnostic.New(
		"LAUNCH_LINK_INVALID",
		"Connector açılış bağlantısı geçersiz.",
		"Chat2API panelinden yeni bir bağlantı oluşturun veya bağlantı kodunu manuel olarak kullanın.",
	)
}
