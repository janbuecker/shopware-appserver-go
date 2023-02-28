package appserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var ErrWebhookMissingEvent = errors.New("missing event")

type WebhookHandlerNotFoundError struct {
	event string
}

func (e WebhookHandlerNotFoundError) Error() string {
	return fmt.Sprintf("no webhook handler found for event: %s", e.event)
}

type WebhookHandler func(ctx context.Context, webhook WebhookRequest, api *APIClient) error

type WebhookRequest struct {
	*AppRequest

	Data struct {
		Payload map[string]interface{} `json:"payload"`
		Event   string                 `json:"event"`
	} `json:"data"`
}

func (srv *Server) HandleWebhook(req *http.Request) error {
	if err := srv.verifyPayloadSignature(req); err != nil {
		return err
	}

	body, err := extractBody(req)
	if err != nil {
		return fmt.Errorf("extract body: %w", err)
	}

	if len(body) == 0 {
		return errors.New("empty payload")
	}

	webhookReq := WebhookRequest{}
	err = json.Unmarshal(body, &webhookReq)
	if err != nil {
		return fmt.Errorf("parse body: %w", err)
	}

	if len(webhookReq.Data.Event) == 0 {
		return ErrWebhookMissingEvent
	}

	h, ok := srv.webhooks[webhookReq.Data.Event]
	if !ok {
		return WebhookHandlerNotFoundError{event: webhookReq.Data.Event}
	}

	credentials, err := srv.credentialStore.Get(req.Context(), webhookReq.Source.ShopID)
	if err != nil {
		return fmt.Errorf("get shop credentials: %w", err)
	}

	err = h(req.Context(), webhookReq, newAPIClient(srv.httpClient, srv.appName, credentials, srv.tokenStore))
	if err != nil {
		return fmt.Errorf("handler: %w", err)
	}

	return nil
}
