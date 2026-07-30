package ui

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/diagnostic"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/pairing"
)

func TestLoopbackUIRequiresExactHostAndOrigin(t *testing.T) {
	instance := testServer(t)
	code := testCapability(t)

	untrusted := request(t, http.MethodPost, instance.basePath+"inspect", `{"code":"`+code+`"}`)
	untrusted.Header.Set("Content-Type", "application/json")
	untrusted.Header.Set("Origin", "https://untrusted.example")
	recorder := httptest.NewRecorder()
	instance.ServeHTTP(recorder, untrusted)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d", recorder.Code)
	}

	wrongHost := request(t, http.MethodGet, instance.basePath, "")
	wrongHost.Host = "attacker.example"
	recorder = httptest.NewRecorder()
	instance.ServeHTTP(recorder, wrongHost)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d", recorder.Code)
	}
}

func TestInspectReturnsOnlySanitizedCapabilityMetadata(t *testing.T) {
	instance := testServer(t)
	code := testCapability(t)
	request := request(t, http.MethodPost, instance.basePath+"inspect", `{"code":"`+code+`"}`)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", instance.origin)
	recorder := httptest.NewRecorder()
	instance.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"gatewayHost":"gateway.example.com"`) {
		t.Fatalf("body = %s", body)
	}
	if strings.Contains(body, strings.Repeat("s", 43)) || strings.Contains(body, code) {
		t.Fatal("inspect response leaked capability material")
	}
}

func TestInspectRejectsInvalidCapabilityWithActionableJSON(t *testing.T) {
	instance := testServer(t)
	request := request(t, http.MethodPost, instance.basePath+"inspect", `{"code":"not-a-capability"}`)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", instance.origin)
	recorder := httptest.NewRecorder()
	instance.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	var status statusView
	if err := json.Unmarshal(recorder.Body.Bytes(), &status); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if status.Phase != "error" || status.ErrorCode != "PAIRING_CODE_INVALID" || status.Hint == "" {
		t.Fatalf("status = %#v", status)
	}
	if strings.Contains(recorder.Body.String(), "not-a-capability") {
		t.Fatal("inspect response leaked rejected capability")
	}
}

func TestPageUsesNonceCSPAndDoesNotEnableCORS(t *testing.T) {
	instance := testServer(t)
	request := request(t, http.MethodGet, instance.basePath, "")
	recorder := httptest.NewRecorder()
	instance.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	csp := recorder.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "script-src 'nonce-") || strings.Contains(csp, "unsafe-inline") {
		t.Fatalf("unexpected CSP: %s", csp)
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("loopback UI unexpectedly enabled CORS")
	}
	if !strings.Contains(recorder.Body.String(), "Chat2API Session Connector") {
		t.Fatal("connector page missing")
	}
	if !strings.Contains(
		recorder.Body.String(),
		`const base = "/session/local-test/";`,
	) {
		t.Fatalf("connector page rendered an invalid local API base: %s", recorder.Body.String())
	}
}

func TestRunServesFunctionalLocalStatusEndpoint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	openedURL := make(chan string, 1)
	runError := make(chan error, 1)

	go func() {
		runError <- Run(
			ctx,
			func(context.Context, pairing.Payload, ProgressFunc) (string, error) {
				return "Test Browser", nil
			},
			func(rawURL string) error {
				openedURL <- rawURL
				return nil
			},
			Options{},
		)
	}()

	var rawURL string
	select {
	case rawURL = <-openedURL:
	case <-time.After(3 * time.Second):
		t.Fatal("loopback UI did not open")
	}

	response, err := http.Get(rawURL + "status")
	if err != nil {
		t.Fatalf("local status endpoint unavailable: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("local status endpoint returned %d", response.StatusCode)
	}
	var status statusView
	if err := json.NewDecoder(response.Body).Decode(&status); err != nil {
		t.Fatalf("local status response is not JSON: %v", err)
	}
	if status.Phase != "idle" {
		t.Fatalf("unexpected local status: %#v", status)
	}

	cancel()
	select {
	case err := <-runError:
		if err != nil {
			t.Fatalf("loopback UI shutdown failed: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("loopback UI did not stop")
	}
}

func TestConnectionFailureKeepsValidCapabilityForLocalRetry(t *testing.T) {
	instance := testServer(t)
	payload, err := pairing.Parse(testCapability(t), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	instance.connecting = true
	instance.connect = func(context.Context, pairing.Payload, ProgressFunc) (string, error) {
		return "Test Browser", diagnostic.New(
			"BROWSER_LAUNCH_FAILED",
			"DeepSeek giriş penceresi açılamadı.",
			"Tarayıcıyı kontrol edin.",
		)
	}

	instance.runConnection(payload)

	if instance.status.Phase != "error" ||
		instance.status.ErrorCode != "BROWSER_LAUNCH_FAILED" ||
		instance.status.CandidateID == "" ||
		instance.candidate == nil {
		t.Fatalf("status = %#v candidate = %#v", instance.status, instance.candidate)
	}
	serialized, err := json.Marshal(instance.status)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(serialized), payload.Secret) {
		t.Fatal("retry status leaked capability secret")
	}
}

func testServer(t *testing.T) *server {
	t.Helper()
	return &server{
		status: statusView{Phase: "idle"},
		connect: func(context.Context, pairing.Payload, ProgressFunc) (string, error) {
			return "Test Browser", nil
		},
		origin:      "http://127.0.0.1:41883",
		host:        "127.0.0.1:41883",
		basePath:    "/session/local-test/",
		nonce:       "test-nonce",
		rootContext: context.Background(),
		shutdown:    make(chan struct{}),
	}
}

func request(t *testing.T, method string, path string, body string) *http.Request {
	t.Helper()
	request := httptest.NewRequest(method, "http://127.0.0.1:41883"+path, strings.NewReader(body))
	request.Host = "127.0.0.1:41883"
	request.RemoteAddr = "127.0.0.1:52341"
	return request
}

func testCapability(t *testing.T) string {
	t.Helper()
	payload := pairing.Payload{
		Version:   1,
		Transport: "native",
		Endpoint:  "https://gateway.example.com/admin/api/deepseek-link/native-complete",
		SessionID: "6f3e75fd-e65f-4f6f-95d8-958bc4fdb759",
		Secret:    strings.Repeat("s", 43),
		ExpiresAt: time.Now().Add(5 * time.Minute).UnixMilli(),
	}
	value, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return pairing.Prefix + base64.RawURLEncoding.EncodeToString(value)
}
