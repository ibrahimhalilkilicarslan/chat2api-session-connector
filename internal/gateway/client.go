package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/pairing"
)

const maxResponseBytes = 64 * 1024

type Client struct {
	httpClient *http.Client
	version    string
}

type completionRequest struct {
	SessionID string `json:"sessionId"`
	Secret    string `json:"secret"`
	Token     string `json:"token"`
}

type completionResponse struct {
	Status string `json:"status"`
	Error  *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func New(version string) *Client {
	return NewWithHTTPClient(version, &http.Client{
		Timeout: 45 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	})
}

func NewWithHTTPClient(version string, httpClient *http.Client) *Client {
	return &Client{httpClient: httpClient, version: version}
}

func (client *Client) Complete(ctx context.Context, payload pairing.Payload, token string) error {
	token = strings.TrimSpace(token)
	if len(token) == 0 || len(token) > 16_384 {
		return errors.New("DeepSeek oturum bilgisi geçersiz")
	}
	if err := payload.Validate(time.Now()); err != nil {
		return err
	}

	body, err := json.Marshal(completionRequest{
		SessionID: payload.SessionID,
		Secret:    payload.Secret,
		Token:     token,
	})
	if err != nil {
		return errors.New("gateway isteği hazırlanamadı")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, payload.Endpoint, bytes.NewReader(body))
	if err != nil {
		return errors.New("gateway isteği hazırlanamadı")
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Cache-Control", "no-store")
	request.Header.Set("X-Chat2API-Connector", "native-v1")
	request.Header.Set("User-Agent", fmt.Sprintf("Chat2API-Session-Connector/%s", client.version))

	response, err := client.httpClient.Do(request)
	if err != nil {
		return errors.New("gateway bağlantısı kurulamadı")
	}
	defer response.Body.Close()

	decoded := completionResponse{}
	reader := io.LimitReader(response.Body, maxResponseBytes)
	if err := json.NewDecoder(reader).Decode(&decoded); err != nil {
		return errors.New("gateway geçerli bir yanıt vermedi")
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 && decoded.Status == "complete" {
		return nil
	}
	if decoded.Error != nil {
		switch decoded.Error.Code {
		case "invalid_link_session":
			return errors.New("bağlantı kodu geçersiz veya süresi dolmuş")
		case "link_session_busy":
			return errors.New("bu bağlantı başka bir doğrulama işlemi tarafından kullanılıyor")
		case "credential_validation_failed":
			return errors.New("DeepSeek oturumu doğrulanamadı")
		case "provider_unavailable":
			return errors.New("DeepSeek sağlayıcısı gateway üzerinde kullanılamıyor")
		}
	}
	return fmt.Errorf("gateway bağlantıyı reddetti (HTTP %d)", response.StatusCode)
}
