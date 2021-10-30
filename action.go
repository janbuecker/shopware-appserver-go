package appserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

var (
	ErrActionMissingAction = errors.New("missing action or entity")
)

type ErrActionHandlerNotFound struct {
	entity string
	action string
}

func (e ErrActionHandlerNotFound) Error() string {
	return fmt.Sprintf("no action handler found for entity %s, action %s", e.entity, e.action)
}

type ActionHandler func(action ActionRequest, api *APIClient) error

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
		return errors.Wrap(err, "extract body")
	}

	actionReq := ActionRequest{}
	err = json.Unmarshal(body, &actionReq)
	if err != nil {
		return errors.Wrap(err, "parse body")
	}

	if len(actionReq.Data.Action) == 0 || len(actionReq.Data.Entity) == 0 {
		return ErrActionMissingAction
	}

	h, ok := srv.actions[actionReq.Data.Entity+actionReq.Data.Action]
	if !ok {
		return ErrActionHandlerNotFound{entity: actionReq.Data.Entity, action: actionReq.Data.Action}
	}

	credentials, err := srv.credentialStore.Get(actionReq.Source.ShopID)
	if err != nil {
		return errors.Wrap(err, "get shop credentials")
	}

	err = h(actionReq, newAPIClient(srv.appName, credentials, srv.tokenStore))
	if err != nil {
		return errors.Wrap(err, "handler")
	}

	return nil
}
