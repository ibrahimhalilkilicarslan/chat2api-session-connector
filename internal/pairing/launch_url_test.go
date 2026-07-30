package pairing

import (
	"net/url"
	"testing"
	"time"
)

func TestParseLaunchURLAcceptsEncodedCapability(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	code := encode(t, validTestPayload(now))
	raw := LaunchScheme + "://" + LaunchHost + "?code=" + url.QueryEscape(code)

	parsed, err := ParseLaunchURL(raw, now)
	if err != nil {
		t.Fatalf("ParseLaunchURL() error = %v", err)
	}
	if parsed.GatewayHost() != "gateway.example.com" {
		t.Fatalf("GatewayHost() = %q", parsed.GatewayHost())
	}
}

func TestParseLaunchURLRejectsAmbiguousOrUnsafeLinks(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	code := url.QueryEscape(encode(t, validTestPayload(now)))
	tests := []string{
		"https://pair.example/?code=" + code,
		LaunchScheme + "://other?code=" + code,
		LaunchScheme + "://" + LaunchHost + "/unexpected?code=" + code,
		LaunchScheme + "://" + LaunchHost + "?code=" + code + "&next=evil",
		LaunchScheme + "://" + LaunchHost + "?code=" + code + "&code=" + code,
		LaunchScheme + "://user@" + LaunchHost + "?code=" + code,
		LaunchScheme + "://" + LaunchHost + "?code=" + code + "#fragment",
	}
	for _, raw := range tests {
		if _, err := ParseLaunchURL(raw, now); err == nil {
			t.Fatalf("ParseLaunchURL(%q) unexpectedly succeeded", raw)
		}
	}
}

func validTestPayload(now time.Time) Payload {
	return Payload{
		Version:   1,
		Transport: "native",
		Endpoint:  "https://gateway.example.com/admin/api/deepseek-link/native-complete",
		SessionID: "6f3e75fd-e65f-4f6f-95d8-958bc4fdb759",
		Secret:    "sssssssssssssssssssssssssssssssssssssssssss",
		ExpiresAt: now.Add(10 * time.Minute).UnixMilli(),
	}
}
