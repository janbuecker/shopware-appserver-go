package appserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var ErrActionMissingAction = errors.New("missing action or entity")

type ActionHandlerNotFoundError struct {
	entity string
	action string
}

func (e ActionHandlerNotFoundError) Error() string {
	return fmt.Sprintf("no action handler found for entity %s, action %s", e.entity, e.action)
}

type ActionHandler func(ctx context.Context, action ActionRequest, api *APIClient) error

type ActionRequest struct {
	*AppRequest

	Data struct {
		IDs    []string `json:"ids"`
		Entity string   `json:"entity"`
		Action string   `json:"action"`
	} `json:"data"`

	Meta struct {
		Timestamp   int64  `json:"timestamp"`
		ReferenceID string `json:"reference"`
		LanguageID  string `json:"language"`
	} `json:"meta"`
}

func (srv *Server) HandleAction(req *http.Request) error {
	if err := srv.verifyPayloadSignature(req); err != nil {
		return err
	}

	body, err := extractBody(req)
	if err != nil {
		return fmt.Errorf("extract body: %w", err)
	}

	actionReq := ActionRequest{}
	err = json.Unmarshal(body, &actionReq)
	if err != nil {
		return fmt.Errorf("parse body: %w", err)
	}

	if len(actionReq.Data.Action) == 0 || len(actionReq.Data.Entity) == 0 {
		return ErrActionMissingAction
	}

	h, ok := srv.actions[actionReq.Data.Entity+actionReq.Data.Action]
	if !ok {
		return ActionHandlerNotFoundError{entity: actionReq.Data.Entity, action: actionReq.Data.Action}
	}

	credentials, err := srv.credentialStore.Get(req.Context(), actionReq.Source.ShopID)
	if err != nil {
		return fmt.Errorf("get shop credentials: %w", err)
	}

	err = h(req.Context(), actionReq, newAPIClient(srv.httpClient, srv.appName, credentials, srv.tokenStore))
	if err != nil {
		return fmt.Errorf("handler: %w", err)
	}

	return nil
}
