package appserver

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
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

type ActionHandler func(action ActionRequest, api *ApiClient) error

type ActionRequest struct {
	Data struct {
		IDs    []string `json:"ids"`
		Entity string   `json:"entity"`
		Action string   `json:"action"`
	} `json:"data"`

	Source struct {
		ShopID     string `json:"shopId"`
		ShopURL    string `json:"url"`
		AppVersion string `json:"appVersion"`
	} `json:"source"`

	Meta struct {
		Timestamp   int64  `json:"timestamp"`
		ReferenceID string `json:"reference"`
		LanguageID  string `json:"language"`
	} `json:"meta"`
}

func (srv *Server) actionHandler(c echo.Context) error {
	req := ActionRequest{}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if len(req.Data.Action) == 0 || len(req.Data.Entity) == 0 {
		return ErrActionMissingAction
	}

	h, ok := srv.actions[req.Data.Entity+req.Data.Action]
	if !ok {
		return ErrActionHandlerNotFound{entity: req.Data.Entity, action: req.Data.Action}
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
