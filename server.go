package appserver

import (
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
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

	return srv
}

func WithCredentialStore(store CredentialStore) ServerOpt {
	return func(s *Server) {
		s.credentialStore = store
	}
}

func (srv *Server) Event(event string, handler WebhookHandler) {
	srv.webhooks[event] = handler
}

func (srv *Server) Action(entity string, action string, handler ActionHandler) {
	srv.actions[entity+action] = handler
}

func extractBody(req *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read body")
	}

	if err := req.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "close body")
	}

	return body, nil
}
