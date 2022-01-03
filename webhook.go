package appserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
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

type WebhookHandler func(webhook WebhookRequest, api *APIClient) error

type WebhookRequest struct {
	*AppRequest

	Data struct {
		Payload []map[string]interface{} `json:"payload"`
		Event   string                 `json:"event"`
	} `json:"data"`
}

func (srv Server) HandleWebhook(req *http.Request) error {
	if err := srv.verifyPayloadSignature(req); err != nil {
		return err
	}

	body, err := extractBody(req)
	if err != nil {
		return errors.Wrap(err, "extract body")
	}

	if len(body) == 0 {
		return errors.New("empty payload")
	}

	webhookReq := WebhookRequest{}
	err = json.Unmarshal(body, &webhookReq)
	if err != nil {
		return errors.Wrap(err, "parse body")
	}

	if len(webhookReq.Data.Event) == 0 {
		return ErrWebhookMissingEvent
	}

	h, ok := srv.webhooks[webhookReq.Data.Event]
	if !ok {
		return ErrWebhookHandlerNotFound{event: webhookReq.Data.Event}
	}

	credentials, err := srv.credentialStore.Get(webhookReq.Source.ShopID)
	if err != nil {
		return errors.Wrap(err, "get shop credentials")
	}

	err = h(webhookReq, newAPIClient(srv.appName, credentials, srv.tokenStore))
	if err != nil {
		return errors.Wrap(err, "handler")
	}

	return nil
}
