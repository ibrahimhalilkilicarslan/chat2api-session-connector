package pairing

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/diagnostic"
)

const (
	Prefix             = "c2a-ds-native-v1."
	completionPath     = "/admin/api/deepseek-link/native-complete"
	maxCodeLength      = 32 * 1024
	maxCapabilityAhead = 12 * time.Minute
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

type Payload struct {
	Version   int    `json:"v"`
	Transport string `json:"transport"`
	Endpoint  string `json:"endpoint"`
	SessionID string `json:"sessionId"`
	Secret    string `json:"secret"`
	ExpiresAt int64  `json:"expiresAt"`
}

func Parse(code string, now time.Time) (Payload, error) {
	var payload Payload
	code = strings.TrimSpace(code)
	if len(code) == 0 || len(code) > maxCodeLength || !strings.HasPrefix(code, Prefix) {
		return payload, diagnostic.New(
			"PAIRING_CODE_INVALID",
			"Bağlantı kodu biçimi geçersiz.",
			"Chat2API panelinden yeni bir bağlantı oluşturun ve kodun tamamını kullanın.",
		)
	}

	encoded := strings.TrimPrefix(code, Prefix)
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return payload, diagnostic.New(
			"PAIRING_CODE_DECODE_FAILED",
			"Bağlantı kodu okunamadı.",
			"Chat2API panelinden yeni bir bağlantı oluşturup tekrar deneyin.",
		)
	}

	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return Payload{}, diagnostic.New(
			"PAIRING_CODE_PAYLOAD_INVALID",
			"Bağlantı kodunun içeriği geçersiz.",
			"Yalnız Chat2API panelinin oluşturduğu connector bağlantısını kullanın.",
		)
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return Payload{}, diagnostic.New(
			"PAIRING_CODE_PAYLOAD_INVALID",
			"Bağlantı kodunun içeriği geçersiz.",
			"Chat2API panelinden yeni bir bağlantı oluşturup tekrar deneyin.",
		)
	}
	if err := payload.Validate(now); err != nil {
		return Payload{}, err
	}
	return payload, nil
}

func (payload Payload) Validate(now time.Time) error {
	if payload.Version != 1 || payload.Transport != "native" {
		return diagnostic.New(
			"PAIRING_VERSION_UNSUPPORTED",
			"Bu bağlantı kodu connector sürümüyle uyumlu değil.",
			"Connector'ın en güncel sürümünü kurup Chat2API'den yeni bağlantı oluşturun.",
		)
	}
	if !uuidPattern.MatchString(strings.ToLower(payload.SessionID)) {
		return diagnostic.New(
			"PAIRING_SESSION_INVALID",
			"Bağlantı oturumu geçersiz.",
			"Chat2API panelinden yeni bir bağlantı oluşturun.",
		)
	}
	if len(payload.Secret) < 32 || len(payload.Secret) > 512 {
		return diagnostic.New(
			"PAIRING_SECRET_INVALID",
			"Bağlantı yetkisi geçersiz.",
			"Chat2API panelinden yeni bir bağlantı oluşturun.",
		)
	}

	nowMillis := now.UnixMilli()
	if payload.ExpiresAt <= nowMillis {
		return diagnostic.New(
			"PAIRING_CODE_EXPIRED",
			"Bağlantı kodunun süresi dolmuş.",
			"Chat2API panelinden yeni bir bağlantı oluşturun.",
		)
	}
	if payload.ExpiresAt > now.Add(maxCapabilityAhead).UnixMilli() {
		return diagnostic.New(
			"PAIRING_LIFETIME_INVALID",
			"Bağlantı kodunun geçerlilik süresi güvenli sınırı aşıyor.",
			"Chat2API ve connector sürümlerini güncelleyip yeni bağlantı oluşturun.",
		)
	}
	_, err := ParseEndpoint(payload.Endpoint)
	return err
}

func ParseEndpoint(raw string) (*url.URL, error) {
	endpoint, err := url.Parse(raw)
	if err != nil || endpoint.Host == "" {
		return nil, diagnostic.New(
			"GATEWAY_URL_INVALID",
			"Chat2API gateway adresi geçersiz.",
			"Bağlantıyı yalnız güvendiğiniz Chat2API panelinden başlatın.",
		)
	}
	if endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return nil, diagnostic.New(
			"GATEWAY_URL_UNSAFE",
			"Chat2API gateway adresi güvenlik kontrolünden geçemedi.",
			"Bağlantıyı iptal edin ve doğru Chat2API panelinden yeniden başlatın.",
		)
	}
	if endpoint.Path != completionPath {
		return nil, diagnostic.New(
			"GATEWAY_PATH_UNSUPPORTED",
			"Chat2API gateway endpointi desteklenmiyor.",
			"Connector ve Chat2API gateway sürümlerini güncelleyin.",
		)
	}
	if endpoint.Scheme == "https" {
		return endpoint, nil
	}
	if endpoint.Scheme == "http" && isLoopbackHost(endpoint.Hostname()) {
		return endpoint, nil
	}
	return nil, diagnostic.New(
		"GATEWAY_HTTPS_REQUIRED",
		"Chat2API gateway güvenli HTTPS bağlantısı kullanmıyor.",
		"Yalnız HTTPS kullanan bir Chat2API gateway'ine bağlanın.",
	)
}

func (payload Payload) GatewayHost() string {
	endpoint, err := ParseEndpoint(payload.Endpoint)
	if err != nil {
		return ""
	}
	return endpoint.Host
}

func (payload Payload) Deadline() time.Time {
	return time.UnixMilli(payload.ExpiresAt)
}

func isLoopbackHost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); err == io.EOF {
		return nil
	}
	return fmt.Errorf("unexpected trailing JSON")
}
