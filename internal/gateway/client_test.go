package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/pairing"
)

func TestCompleteSubmitsNativeCapabilityWithoutLeakingToken(t *testing.T) {
	token := "private-token-value"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Chat2API-Connector") != "native-v1" {
			t.Error("missing native connector header")
		}
		var body map[string]string
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["token"] != token {
			t.Error("token was not submitted")
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"status":"complete"}`))
	}))
	defer server.Close()

	payload := validPayload(server.URL + "/admin/api/deepseek-link/native-complete")
	if err := New("test").Complete(context.Background(), payload, token); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
}

func TestCompleteRejectsRedirectWithoutForwardingCredential(t *testing.T) {
	var received atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		received.Store(true)
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer target.Close()

	redirect := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, target.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()

	payload := validPayload(redirect.URL + "/admin/api/deepseek-link/native-complete")
	err := New("test").Complete(context.Background(), payload, "private-token-value")
	if err == nil {
		t.Fatal("Complete() unexpectedly succeeded")
	}
	if received.Load() {
		t.Fatal("credential followed an HTTP redirect")
	}
	if strings.Contains(err.Error(), "private-token-value") {
		t.Fatal("error leaked the token")
	}
}

func TestCompleteDoesNotEchoRejectedToken(t *testing.T) {
	token := "private-token-value"
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = writer.Write([]byte(`{"error":{"code":"invalid","message":"private-token-value"}}`))
	}))
	defer server.Close()

	err := New("test").Complete(
		context.Background(),
		validPayload(server.URL+"/admin/api/deepseek-link/native-complete"),
		token,
	)
	if err == nil {
		t.Fatal("Complete() unexpectedly succeeded")
	}
	if strings.Contains(err.Error(), token) {
		t.Fatal("error leaked the token")
	}
}

func TestCompleteMapsSafeGatewayValidationReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = writer.Write([]byte(
			`{"error":{"code":"credential_validation_failed","message":"Provider session response could not be verified.","details":{"reason":"provider_protocol_changed"}}}`,
		))
	}))
	defer server.Close()

	err := New("test").Complete(
		context.Background(),
		validPayload(server.URL+"/admin/api/deepseek-link/native-complete"),
		"private-token-value",
	)
	if err == nil {
		t.Fatal("Complete() unexpectedly succeeded")
	}
	if !strings.Contains(err.Error(), "DeepSeek oturum yanıtı") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func validPayload(endpoint string) pairing.Payload {
	return pairing.Payload{
		Version:   1,
		Transport: "native",
		Endpoint:  endpoint,
		SessionID: "6f3e75fd-e65f-4f6f-95d8-958bc4fdb759",
		Secret:    strings.Repeat("s", 43),
		ExpiresAt: time.Now().Add(5 * time.Minute).UnixMilli(),
	}
}
