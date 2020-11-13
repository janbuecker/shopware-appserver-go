package appserver

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

var (
	ErrWebhookMissingEvent = errors.New("missing event")
)

type ErrWebhookHandlerNotFound struct {
	event string
}

func (e ErrWebhookHandlerNotFound) Error() string {
	return fmt.Sprintf("no webhook handler found for event: %s", e.event)
}

type WebhookHandler func(webhook WebhookRequest, api *ApiClient) error

type WebhookRequest struct {
	Data struct {
		Payload map[string]interface{} `json:"payload"`
		Event   string                 `json:"event"`
	} `json:"data"`

	Source struct {
		ShopID     string `json:"shopId"`
		ShopURL    string `json:"url"`
		AppVersion string `json:"appVersion"`
	} `json:"source"`
}

func (srv *Server) webhookHandler(c echo.Context) error {
	req := WebhookRequest{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if len(req.Data.Event) == 0 {
		return ErrWebhookMissingEvent
	}

	h, ok := srv.webhooks[req.Data.Event]
	if !ok {
		return ErrWebhookHandlerNotFound{event: req.Data.Event}
	}

	credentials, err := srv.credentialStore.Get(req.Source.ShopID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	err = h(req, newApiClient(srv.appName, credentials, srv.tokenStore))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	return c.NoContent(http.StatusOK)
}
