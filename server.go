package appserver

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"time"
)

const (
	HeaderAppSignature     = "shopware-app-signature"
	HeaderPayloadSignature = "shopware-shop-signature"
)

type ServerOpt func(s *Server)

type Server struct {
	confirmationURL string
	appName         string
	appSecret       string

	webhooks map[string]WebhookHandler
	actions  map[string]ActionHandler

	credentialStore CredentialStore
	tokenStore      *tokenStore

	httpClient *http.Client
}

type Credentials struct {
	APIKey     string `json:"apiKey"`
	SecretKey  string `json:"secretKey"`
	Timestamp  string `json:"timestamp" query:"timestamp"`
	ShopURL    string `json:"shopUrl" query:"shop-url"`
	ShopID     string `json:"shopId" query:"shop-id"`
	ShopSecret string `json:"shopSecret"`
}

type Source struct {
	ShopID     string `json:"shopId"`
	ShopURL    string `json:"url"`
	AppVersion string `json:"appVersion"`
}

type AppRequest struct {
	Source Source `json:"source"`
}

func NewServer(appName string, appSecret string, confirmationURL string, opts ...ServerOpt) *Server {
	credentialStore := NewMemoryCredentialStore()

	srv := &Server{
		webhooks: make(map[string]WebhookHandler),
		actions:  make(map[string]ActionHandler),

		credentialStore: credentialStore,
		tokenStore:      newTokenStore(),

		confirmationURL: confirmationURL,
		appName:         appName,
		appSecret:       appSecret,
	}

	for _, o := range opts {
		o(srv)
	}

	if srv.httpClient == nil {
		srv.httpClient = createDefaultHTTPClient()
	}

	return srv
}

func WithCredentialStore(store CredentialStore) ServerOpt {
	return func(s *Server) {
		s.credentialStore = store
	}
}

func WithHTTPClient(client *http.Client) ServerOpt {
	return func(s *Server) {
		s.httpClient = client
	}
}

func (srv *Server) Event(event string, handler WebhookHandler) {
	srv.webhooks[event] = handler
}

func (srv *Server) Action(entity string, action string, handler ActionHandler) {
	srv.actions[entity+action] = handler
}

func extractBody(req *http.Request) ([]byte, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if err := req.Body.Close(); err != nil {
		return nil, fmt.Errorf("close body: %w", err)
	}

	return body, nil
}

func createDefaultHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ForceAttemptHTTP2:     true,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	}
}
