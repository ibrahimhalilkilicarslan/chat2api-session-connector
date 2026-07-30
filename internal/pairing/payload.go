package pairing

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	Prefix             = "c2a-ds-native-v1."
	completionPath     = "/admin/api/deepseek-link/native-complete"
	maxCodeLength      = 32 * 1024
	maxCapabilityAhead = 10 * time.Minute
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
		return payload, errors.New("bağlantı kodu biçimi geçersiz")
	}

	encoded := strings.TrimPrefix(code, Prefix)
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return payload, errors.New("bağlantı kodu çözülemedi")
	}

	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return Payload{}, errors.New("bağlantı kodu içeriği geçersiz")
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return Payload{}, errors.New("bağlantı kodu içeriği geçersiz")
	}
	if err := payload.Validate(now); err != nil {
		return Payload{}, err
	}
	return payload, nil
}

func (payload Payload) Validate(now time.Time) error {
	if payload.Version != 1 || payload.Transport != "native" {
		return errors.New("desteklenmeyen bağlantı kodu sürümü")
	}
	if !uuidPattern.MatchString(strings.ToLower(payload.SessionID)) {
		return errors.New("bağlantı oturumu geçersiz")
	}
	if len(payload.Secret) < 32 || len(payload.Secret) > 512 {
		return errors.New("bağlantı yetkisi geçersiz")
	}

	nowMillis := now.UnixMilli()
	if payload.ExpiresAt <= nowMillis {
		return errors.New("bağlantı kodunun süresi dolmuş")
	}
	if payload.ExpiresAt > now.Add(maxCapabilityAhead).UnixMilli() {
		return errors.New("bağlantı kodunun geçerlilik süresi güvenli sınırı aşıyor")
	}
	_, err := ParseEndpoint(payload.Endpoint)
	return err
}

func ParseEndpoint(raw string) (*url.URL, error) {
	endpoint, err := url.Parse(raw)
	if err != nil || endpoint.Host == "" {
		return nil, errors.New("gateway adresi geçersiz")
	}
	if endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return nil, errors.New("gateway adresi güvenli değil")
	}
	if endpoint.Path != completionPath {
		return nil, errors.New("gateway endpoint yolu desteklenmiyor")
	}
	if endpoint.Scheme == "https" {
		return endpoint, nil
	}
	if endpoint.Scheme == "http" && isLoopbackHost(endpoint.Hostname()) {
		return endpoint, nil
	}
	return nil, errors.New("gateway HTTPS kullanmalı")
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
