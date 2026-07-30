package ui

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/diagnostic"
	"github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/pairing"
)

const (
	maxRequestBytes = 40 * 1024
	idleLifetime    = 15 * time.Minute
)

type ConnectionProgress struct {
	Stage       string
	Message     string
	BrowserName string
}

type ProgressFunc func(ConnectionProgress)
type ConnectFunc func(context.Context, pairing.Payload, ProgressFunc) (string, error)
type OpenURLFunc func(string) error

type statusView struct {
	Phase       string `json:"phase"`
	Message     string `json:"message"`
	Hint        string `json:"hint,omitempty"`
	ErrorCode   string `json:"errorCode,omitempty"`
	Stage       string `json:"stage,omitempty"`
	GatewayHost string `json:"gatewayHost,omitempty"`
	ExpiresAt   int64  `json:"expiresAt,omitempty"`
	BrowserName string `json:"browserName,omitempty"`
	CandidateID string `json:"candidateId,omitempty"`
}

type Options struct {
	InitialPayload    *pairing.Payload
	InitialError      error
	Notice            string
	InstallationReady bool
	Version           string
}

type candidate struct {
	id      string
	payload pairing.Payload
}

type server struct {
	mu                sync.Mutex
	status            statusView
	candidate         *candidate
	connecting        bool
	connect           ConnectFunc
	origin            string
	host              string
	basePath          string
	nonce             string
	notice            string
	installationReady bool
	version           string
	rootContext       context.Context
	shutdown          chan struct{}
	shutdownOnce      sync.Once
}

func Run(ctx context.Context, connect ConnectFunc, openURL OpenURLFunc, options Options) error {
	if connect == nil || openURL == nil {
		return errors.New("connector başlatılamadı")
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return errors.New("yerel connector portu açılamadı")
	}
	defer listener.Close()

	sessionSecret, err := randomValue(32)
	if err != nil {
		return errors.New("yerel connector oturumu oluşturulamadı")
	}
	nonce, err := randomValue(18)
	if err != nil {
		return errors.New("yerel connector güvenlik anahtarı oluşturulamadı")
	}

	instance := &server{
		status: statusView{
			Phase:   "idle",
			Message: "Chat2API panelindeki tek kullanımlık bağlantı kodunu girin.",
		},
		connect:           connect,
		origin:            "http://" + listener.Addr().String(),
		host:              listener.Addr().String(),
		basePath:          "/session/" + sessionSecret + "/",
		nonce:             nonce,
		notice:            options.Notice,
		installationReady: options.InstallationReady,
		version:           options.Version,
		rootContext:       ctx,
		shutdown:          make(chan struct{}),
	}
	if options.InitialError != nil {
		errorCode, hint := diagnostic.Details(options.InitialError)
		instance.status = statusView{
			Phase:     "error",
			Message:   options.InitialError.Error(),
			Hint:      hint,
			ErrorCode: errorCode,
		}
	}
	if options.InitialPayload != nil {
		candidateID, candidateError := randomValue(24)
		if candidateError != nil {
			return errors.New("bağlantı onayı hazırlanamadı")
		}
		instance.candidate = &candidate{id: candidateID, payload: *options.InitialPayload}
		instance.status = statusView{
			Phase:       "confirm",
			Message:     "Gateway adresini kontrol edip bağlantıyı başlatın.",
			GatewayHost: options.InitialPayload.GatewayHost(),
			ExpiresAt:   options.InitialPayload.ExpiresAt,
			CandidateID: candidateID,
		}
	}

	httpServer := &http.Server{
		Handler:           instance,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    16 * 1024,
	}
	serverErrors := make(chan error, 1)
	go func() {
		err := httpServer.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	if err := openURL(instance.origin + instance.basePath); err != nil {
		_ = httpServer.Shutdown(context.Background())
		return err
	}

	idleTimer := time.NewTimer(idleLifetime)
	defer idleTimer.Stop()
	select {
	case <-ctx.Done():
	case <-instance.shutdown:
	case <-idleTimer.C:
	case err := <-serverErrors:
		return err
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return httpServer.Shutdown(shutdownContext)
}

func (instance *server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	instance.securityHeaders(writer)
	if !instance.isLocalRequest(request) {
		http.Error(writer, "Forbidden", http.StatusForbidden)
		return
	}
	relative, ok := strings.CutPrefix(request.URL.Path, instance.basePath)
	if !ok || strings.Contains(relative, "/") {
		http.NotFound(writer, request)
		return
	}

	switch {
	case request.Method == http.MethodGet && relative == "":
		instance.servePage(writer)
	case request.Method == http.MethodPost && relative == "inspect":
		instance.requireSameOriginJSON(writer, request, instance.inspect)
	case request.Method == http.MethodPost && relative == "connect":
		instance.requireSameOriginJSON(writer, request, instance.startConnection)
	case request.Method == http.MethodGet && relative == "status":
		instance.writeStatus(writer)
	case request.Method == http.MethodPost && relative == "reset":
		instance.requireSameOriginJSON(writer, request, instance.reset)
	case request.Method == http.MethodPost && relative == "shutdown":
		instance.requireSameOriginJSON(writer, request, instance.stop)
	default:
		http.NotFound(writer, request)
	}
}

func (instance *server) inspect(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		Code string `json:"code"`
	}
	if !decodeJSON(writer, request, &body) {
		return
	}
	payload, err := pairing.Parse(body.Code, time.Now())
	body.Code = ""
	if err != nil {
		errorCode, hint := diagnostic.Details(err)
		status := statusView{
			Phase:     "error",
			Message:   err.Error(),
			Hint:      hint,
			ErrorCode: errorCode,
		}
		instance.setStatus(status)
		writeJSON(writer, http.StatusBadRequest, status)
		return
	}
	id, err := randomValue(24)
	if err != nil {
		status := statusView{
			Phase:     "error",
			Message:   "Bağlantı onayı hazırlanamadı.",
			Hint:      "Connector'ı kapatıp yeniden açın.",
			ErrorCode: "PAIRING_PREPARATION_FAILED",
		}
		instance.setStatus(status)
		writeJSON(writer, http.StatusInternalServerError, status)
		return
	}

	instance.mu.Lock()
	if instance.connecting {
		instance.mu.Unlock()
		http.Error(writer, "Connection already in progress", http.StatusConflict)
		return
	}
	instance.candidate = &candidate{id: id, payload: payload}
	instance.status = statusView{
		Phase:       "confirm",
		Message:     "Gateway adresini kontrol edip bağlantıyı başlatın.",
		GatewayHost: payload.GatewayHost(),
		ExpiresAt:   payload.ExpiresAt,
		CandidateID: id,
	}
	status := instance.status
	instance.mu.Unlock()
	writeJSON(writer, http.StatusOK, status)
}

func (instance *server) startConnection(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		CandidateID string `json:"candidateId"`
	}
	if !decodeJSON(writer, request, &body) {
		return
	}

	instance.mu.Lock()
	if instance.connecting {
		instance.mu.Unlock()
		http.Error(writer, "Connection already in progress", http.StatusConflict)
		return
	}
	if instance.candidate == nil || body.CandidateID == "" || body.CandidateID != instance.candidate.id {
		instance.mu.Unlock()
		http.Error(writer, "Invalid candidate", http.StatusBadRequest)
		return
	}
	payload := instance.candidate.payload
	instance.candidate = nil
	instance.connecting = true
	instance.status = statusView{
		Phase:       "connecting",
		Stage:       "starting",
		Message:     "DeepSeek giriş penceresi açılıyor. Girişi bu pencerede tamamlayın.",
		GatewayHost: payload.GatewayHost(),
		ExpiresAt:   payload.ExpiresAt,
	}
	status := instance.status
	instance.mu.Unlock()
	writeJSON(writer, http.StatusAccepted, status)

	go instance.runConnection(payload)
}

func (instance *server) runConnection(payload pairing.Payload) {
	deadline := payload.Deadline()
	if deadline.After(time.Now().Add(6 * time.Minute)) {
		deadline = time.Now().Add(6 * time.Minute)
	}
	ctx, cancel := context.WithDeadline(instance.rootContext, deadline)
	defer cancel()

	report := func(progress ConnectionProgress) {
		instance.mu.Lock()
		if instance.connecting {
			instance.status = statusView{
				Phase:       "connecting",
				Stage:       progress.Stage,
				Message:     progress.Message,
				GatewayHost: payload.GatewayHost(),
				ExpiresAt:   payload.ExpiresAt,
				BrowserName: progress.BrowserName,
			}
		}
		instance.mu.Unlock()
	}
	browserName, err := instance.connect(ctx, payload, report)
	retryCandidateID := ""
	if err != nil && payload.Deadline().After(time.Now()) {
		retryCandidateID, _ = randomValue(24)
	}
	instance.mu.Lock()
	instance.connecting = false
	if err != nil {
		errorCode, hint := diagnostic.Details(err)
		if retryCandidateID != "" {
			instance.candidate = &candidate{id: retryCandidateID, payload: payload}
		}
		instance.status = statusView{
			Phase:       "error",
			Message:     err.Error(),
			Hint:        hint,
			ErrorCode:   errorCode,
			GatewayHost: payload.GatewayHost(),
			ExpiresAt:   payload.ExpiresAt,
			BrowserName: browserName,
			CandidateID: retryCandidateID,
		}
	} else {
		instance.status = statusView{
			Phase:       "complete",
			Stage:       "complete",
			Message:     "DeepSeek hesabı doğrulandı ve Chat2API hesabı oluşturuldu.",
			GatewayHost: payload.GatewayHost(),
			BrowserName: browserName,
		}
	}
	instance.mu.Unlock()
}

func (instance *server) writeStatus(writer http.ResponseWriter) {
	instance.mu.Lock()
	status := instance.status
	instance.mu.Unlock()
	writeJSON(writer, http.StatusOK, status)
}

func (instance *server) reset(writer http.ResponseWriter, request *http.Request) {
	if !decodeJSON(writer, request, &struct{}{}) {
		return
	}
	instance.mu.Lock()
	if instance.connecting {
		instance.mu.Unlock()
		http.Error(writer, "Connection already in progress", http.StatusConflict)
		return
	}
	instance.candidate = nil
	instance.status = statusView{
		Phase:   "idle",
		Message: "Chat2API panelindeki tek kullanımlık bağlantı kodunu girin.",
	}
	status := instance.status
	instance.mu.Unlock()
	writeJSON(writer, http.StatusOK, status)
}

func (instance *server) stop(writer http.ResponseWriter, request *http.Request) {
	if !decodeJSON(writer, request, &struct{}{}) {
		return
	}
	instance.mu.Lock()
	connecting := instance.connecting
	instance.mu.Unlock()
	if connecting {
		http.Error(writer, "Connection already in progress", http.StatusConflict)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "closing"})
	instance.shutdownOnce.Do(func() { close(instance.shutdown) })
}

func (instance *server) requireSameOriginJSON(
	writer http.ResponseWriter,
	request *http.Request,
	handler func(http.ResponseWriter, *http.Request),
) {
	if request.Header.Get("Origin") != instance.origin {
		http.Error(writer, "Forbidden", http.StatusForbidden)
		return
	}
	contentType := strings.ToLower(request.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "application/json") {
		http.Error(writer, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}
	handler(writer, request)
}

func (instance *server) isLocalRequest(request *http.Request) bool {
	if request.Host != instance.host {
		return false
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (instance *server) securityHeaders(writer http.ResponseWriter) {
	writer.Header().Set("Cache-Control", "no-store")
	writer.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'nonce-"+instance.nonce+"'; script-src 'nonce-"+instance.nonce+"'; connect-src 'self'; img-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
	writer.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
	writer.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
	writer.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
	writer.Header().Set("Referrer-Policy", "no-referrer")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("X-Frame-Options", "DENY")
}

func (instance *server) servePage(writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = pageTemplate.Execute(writer, struct {
		BasePath          string
		Nonce             string
		Notice            string
		InstallationReady bool
		Version           string
	}{
		BasePath:          instance.basePath,
		Nonce:             instance.nonce,
		Notice:            instance.notice,
		InstallationReady: instance.installationReady,
		Version:           instance.version,
	})
}

func (instance *server) setStatus(status statusView) {
	instance.mu.Lock()
	instance.status = status
	instance.mu.Unlock()
}

func decodeJSON(writer http.ResponseWriter, request *http.Request, target any) bool {
	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		http.Error(writer, "Invalid request", http.StatusBadRequest)
		return false
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		http.Error(writer, "Invalid request", http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func randomValue(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

var pageTemplate = template.Must(template.New("connector").Parse(pageHTML))
