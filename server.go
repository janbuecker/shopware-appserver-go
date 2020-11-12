package appserver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

const (
	HeaderAppSignature     = "shopware-app-signature"
	HeaderWebhookSignature = "shopware-shop-signature"
	RouteRegister          = "/setup/register"
	RouteRegisterConfirm   = "/setup/register-confirm"
	RouteWebhook           = "/webhook"
)

var (
	ErrInvalidSignature = errors.New("invalid signature")
)

type ServerOpt func(s *Server)

type Server struct {
	srv *echo.Echo

	serverURL string
	appName   string
	appSecret string

	webhooks        map[string]WebhookHandler
	credentialStore CredentialStore
	tokenStore      *tokenStore
}

type RegistrationResponse struct {
	Proof           string `json:"proof"`
	Secret          string `json:"secret"`
	ConfirmationURL string `json:"confirmation_url"`
}

type Credentials struct {
	APIKey    string `json:"apiKey"`
	SecretKey string `json:"secretKey"`
	Timestamp string `json:"timestamp"`
	ShopURL   string `json:"shopUrl"`
	ShopID    string `json:"shopId"`
}

func NewServer(serverURL string, appName string, appSecret string, opts ...ServerOpt) *Server {
	e := echo.New()
	credentialStore := NewMemoryCredentialStore()

	srv := &Server{
		srv: e,

		webhooks:        make(map[string]WebhookHandler),
		credentialStore: credentialStore,
		tokenStore:      newTokenStore(),

		serverURL: serverURL,
		appName:   appName,
		appSecret: appSecret,
	}

	for _, o := range opts {
		o(srv)
	}

	// global middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	// registration workflow
	e.GET(RouteRegister, srv.registerHandler)
	e.POST(RouteRegisterConfirm, srv.confirmHandler)

	// webhooks
	e.POST(RouteWebhook, srv.webhookHandler)

	return srv
}

func WithCredentialStore(store CredentialStore) ServerOpt {
	return func(s *Server) {
		s.credentialStore = store
	}
}

func (srv *Server) Start(listenAddr string) error {
	return srv.srv.Start(listenAddr)
}

func (srv *Server) On(event string, handler WebhookHandler) {
	srv.webhooks[event] = handler
}

func (srv *Server) registerHandler(c echo.Context) error {
	if !srv.verifySignature([]byte(c.QueryString()), c.Request().Header.Get(HeaderAppSignature)) {
		return echo.NewHTTPError(http.StatusBadRequest, ErrInvalidSignature)
	}

	h := hmac.New(sha256.New, []byte(srv.appSecret))
	h.Write([]byte(c.QueryParam("shop-id") + c.QueryParam("shop-url") + srv.appName))

	return c.JSON(http.StatusOK, RegistrationResponse{
		Secret:          srv.appSecret,
		Proof:           hex.EncodeToString(h.Sum(nil)),
		ConfirmationURL: srv.serverURL + RouteRegisterConfirm,
	})
}

func (srv *Server) verifySignature(data []byte, signature string) bool {
	h := hmac.New(sha256.New, []byte(srv.appSecret))
	h.Write(data)

	return hex.EncodeToString(h.Sum(nil)) == signature
}

func (srv *Server) confirmHandler(c echo.Context) error {
	credentials := Credentials{}
	if err := c.Bind(&credentials); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := srv.credentialStore.Store(&credentials)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.NoContent(http.StatusOK)
}
