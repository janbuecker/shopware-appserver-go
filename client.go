package appserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type APIClient struct {
	appName     string
	credentials *Credentials
	tokenStore  *tokenStore
	httpClient  *retryablehttp.Client
}

func newAPIClient(appName string, credentials *Credentials, tokenStore *tokenStore) *APIClient {
	return &APIClient{
		appName:     appName,
		credentials: credentials,
		tokenStore:  tokenStore,
		httpClient:  retryablehttp.NewClient(),
	}
}

func (c *APIClient) Request(method string, path string, payload interface{}) (*http.Response, error) {
	token, err := c.getTokenForShop(c.credentials.ShopID)
	if err != nil {
		return nil, fmt.Errorf("get token: %v", err)
	}

	var pdata []byte
	if payload != nil {
		pdata, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encode payload: %v", err)
		}
	}

	req, err := retryablehttp.NewRequest(method, c.credentials.ShopURL+path, pdata)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	return c.httpClient.Do(req)
}

func (c *APIClient) GetAppConfig() (map[string]interface{}, error) {
	resp, err := c.Request(http.MethodGet, "/api/_action/system-config?domain="+c.appName+".config", nil)
	if err != nil {
		return nil, err
	}

	bodyString, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()

	out := map[string]interface{}{}
	err = json.Unmarshal(bodyString, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (c *APIClient) getTokenForShop(shopID string) (*oauth2.Token, error) {
	if token, ok := c.tokenStore.Get(shopID); ok {
		return token, nil
	}

	cc := clientcredentials.Config{
		ClientID:     c.credentials.APIKey,
		ClientSecret: c.credentials.SecretKey,
		TokenURL:     c.credentials.ShopURL + "/api/oauth/token",
		AuthStyle:    oauth2.AuthStyleInParams,
	}

	token, err := cc.Token(context.Background())
	if err != nil {
		return nil, err
	}

	c.tokenStore.Store(shopID, token)

	return token, nil
}

const (
	TotalCountModeDefault  = 0
	TotalCountModeExact    = 1
	TotalCountModeNextPage = 2

	SearchFilterTypeEquals    = "equals"
	SearchFilterTypeEqualsAny = "equalsAny"

	SearchSortDirectionAscending  = "ASC"
	SearchSortDirectionDescending = "DESC"
)

type Search struct {
	Includes       map[string][]string `json:"includes,omitempty"`
	Page           int64               `json:"page,omitempty"`
	Limit          int64               `json:"limit,omitempty"`
	IDs            []string            `json:"ids,omitempty"`
	Filter         []SearchFilter      `json:"filter,omitempty"`
	PostFilter     []SearchFilter      `json:"postFilter,omitempty"`
	Sort           []SearchSort        `json:"sort,omitempty"`
	Term           string              `json:"term,omitempty"`
	TotalCountMode int                 `json:"totalCountMode,omitempty"`
}

type SearchFilter struct {
	Type  string      `json:"type"`
	Field string      `json:"field"`
	Value interface{} `json:"value"`
}

type SearchSort struct {
	Direction      string `json:"order"`
	Field          string `json:"field"`
	NaturalSorting bool   `json:"naturalSorting"`
}

type SearchResponse struct {
	Total        int64       `json:"total"`
	Data         interface{} `json:"data"`
	Aggregations interface{} `json:"aggregations"`
}
