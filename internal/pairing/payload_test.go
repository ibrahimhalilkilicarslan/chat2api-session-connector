package pairing

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestParseAcceptsNativeCapability(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	payload := Payload{
		Version:   1,
		Transport: "native",
		Endpoint:  "https://gateway.example.com/admin/api/deepseek-link/native-complete",
		SessionID: "6f3e75fd-e65f-4f6f-95d8-958bc4fdb759",
		Secret:    strings.Repeat("s", 43),
		ExpiresAt: now.Add(5 * time.Minute).UnixMilli(),
	}
	parsed, err := Parse(encode(t, payload), now)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if parsed.GatewayHost() != "gateway.example.com" {
		t.Fatalf("GatewayHost() = %q", parsed.GatewayHost())
	}
}

func TestParseRejectsUnsafeCapabilities(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	base := Payload{
		Version:   1,
		Transport: "native",
		Endpoint:  "https://gateway.example.com/admin/api/deepseek-link/native-complete",
		SessionID: "6f3e75fd-e65f-4f6f-95d8-958bc4fdb759",
		Secret:    strings.Repeat("s", 43),
		ExpiresAt: now.Add(5 * time.Minute).UnixMilli(),
	}

	tests := map[string]func(*Payload){
		"expired": func(value *Payload) {
			value.ExpiresAt = now.Add(-time.Second).UnixMilli()
		},
		"too long": func(value *Payload) {
			value.ExpiresAt = now.Add(11 * time.Minute).UnixMilli()
		},
		"wrong transport": func(value *Payload) {
			value.Transport = "browser-extension"
		},
		"wrong path": func(value *Payload) {
			value.Endpoint = "https://gateway.example.com/collect"
		},
		"plain HTTP": func(value *Payload) {
			value.Endpoint = "http://gateway.example.com/admin/api/deepseek-link/native-complete"
		},
		"URL credentials": func(value *Payload) {
			value.Endpoint = "https://user:pass@gateway.example.com/admin/api/deepseek-link/native-complete"
		},
		"query": func(value *Payload) {
			value.Endpoint = "https://gateway.example.com/admin/api/deepseek-link/native-complete?next=evil"
		},
	}

	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			candidate := base
			mutate(&candidate)
			if _, err := Parse(encode(t, candidate), now); err == nil {
				t.Fatal("Parse() unexpectedly succeeded")
			}
		})
	}
}

func TestParseAllowsLoopbackDevelopment(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	payload := Payload{
		Version:   1,
		Transport: "native",
		Endpoint:  "http://127.0.0.1:8080/admin/api/deepseek-link/native-complete",
		SessionID: "6f3e75fd-e65f-4f6f-95d8-958bc4fdb759",
		Secret:    strings.Repeat("s", 43),
		ExpiresAt: now.Add(5 * time.Minute).UnixMilli(),
	}
	if _, err := Parse(encode(t, payload), now); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func encode(t *testing.T, payload Payload) string {
	t.Helper()
	value, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return Prefix + base64.RawURLEncoding.EncodeToString(value)
}
