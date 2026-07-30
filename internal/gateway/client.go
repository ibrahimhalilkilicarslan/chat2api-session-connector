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

	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/diagnostic"
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
		Details struct {
			Reason string `json:"reason"`
		} `json:"details"`
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
		return diagnostic.New(
			"TOKEN_INVALID",
			"DeepSeek oturum bilgisi alınamadı.",
			"DeepSeek hesabında girişin tamamlandığını doğrulayıp yeni bir bağlantı başlatın.",
		)
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
		return diagnostic.New(
			"GATEWAY_UNREACHABLE",
			"Chat2API gateway'ine bağlanılamadı.",
			"İnternet bağlantısını ve Chat2API adresinin tarayıcıdan açıldığını kontrol edip yeni bir bağlantı başlatın.",
		)
	}
	defer response.Body.Close()

	decoded := completionResponse{}
	reader := io.LimitReader(response.Body, maxResponseBytes)
	if err := json.NewDecoder(reader).Decode(&decoded); err != nil {
		return diagnostic.New(
			"GATEWAY_RESPONSE_INVALID",
			"Chat2API gateway'i geçerli bir yanıt vermedi.",
			"Gateway'in güncel ve erişilebilir olduğunu kontrol edin.",
		)
	}
	if response.StatusCode >= 200 && response.StatusCode < 300 && decoded.Status == "complete" {
		return nil
	}
	if decoded.Error != nil {
		switch decoded.Error.Code {
		case "invalid_link_session":
			return diagnostic.New(
				"LINK_EXPIRED",
				"Bağlantı oturumunun süresi dolmuş veya oturum kullanılmış.",
				"Chat2API panelinden yeni bir bağlantı başlatın.",
			)
		case "link_session_busy":
			return diagnostic.New(
				"LINK_BUSY",
				"Bu bağlantı başka bir doğrulama işlemi tarafından kullanılıyor.",
				"Devam eden connector penceresini tamamlayın veya Chat2API panelinden yeni bağlantı başlatın.",
			)
		case "credential_validation_failed":
			switch decoded.Error.Details.Reason {
			case "provider_rate_limited":
				return diagnostic.New(
					"PROVIDER_RATE_LIMITED",
					"DeepSeek oturumu geçerli ancak sağlayıcı şu anda yoğun.",
					"Birkaç dakika bekleyip Chat2API panelinden yeni bir bağlantı başlatın.",
				)
			case "provider_protocol_changed":
				return diagnostic.New(
					"PROVIDER_PROTOCOL_CHANGED",
					"DeepSeek oturum yanıtı güvenli biçimde doğrulanamadı.",
					"Connector ve Chat2API gateway sürümlerinin güncel olduğunu kontrol edin.",
				)
			case "provider_unavailable":
				return diagnostic.New(
					"PROVIDER_UNAVAILABLE",
					"DeepSeek oturum kontrolüne şu anda ulaşılamıyor.",
					"Birkaç dakika bekleyip yeniden deneyin; sorun sürerse gateway ağ erişimini kontrol edin.",
				)
			}
			return diagnostic.New(
				"SESSION_REJECTED",
				"DeepSeek oturumu doğrulanamadı.",
				"DeepSeek hesabından çıkış yapıp connector'ın açtığı pencerede yeniden giriş yapın.",
			)
		case "provider_unavailable":
			return diagnostic.New(
				"PROVIDER_UNAVAILABLE",
				"DeepSeek sağlayıcısı Chat2API üzerinde kullanılamıyor.",
				"Chat2API sağlayıcı ayarlarını kontrol edin.",
			)
		}
	}
	return diagnostic.New(
		"GATEWAY_REJECTED",
		fmt.Sprintf("Chat2API gateway'i bağlantıyı reddetti (HTTP %d).", response.StatusCode),
		"Chat2API panelinden yeni bir bağlantı oluşturup tekrar deneyin.",
	)
}
