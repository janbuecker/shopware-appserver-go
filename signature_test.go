package appserver_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	appserver "github.com/shopwareLabs/GoAppserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Signature for payload below.
	verifyPayloadSignatureTestSignature = "f88bce849a86f16b9740eceb9190bff7d2a58c0a930d3afad5abcdb2162abacb"
	verifyPayloadSignatureTestPayload   = `{"data":{"event":"foo"},"source":{"shopId":"123"}}`
)

func TestVerifyPayloadSignature(t *testing.T) {
	store := appserver.NewMemoryCredentialStore()
	require.NoError(t, store.Store(&appserver.Credentials{
		ShopID:     "123",
		ShopSecret: "mysecret",
	}))

	srv := appserver.NewServer("", "", "mysecret", appserver.WithCredentialStore(store))
	srv.Event("foo", func(webhook appserver.WebhookRequest, api *appserver.APIClient) error {
		return nil
	})

	// test without signature
	req := httptest.NewRequest("GET", "/", strings.NewReader(verifyPayloadSignatureTestPayload))
	err := srv.HandleWebhook(req)
	assert.EqualError(t, err, "invalid signature")

	// test with invalid signature
	req = httptest.NewRequest("POST", "/signature", strings.NewReader(verifyPayloadSignatureTestPayload))
	req.Header.Set(appserver.HeaderPayloadSignature, "foo")
	err = srv.HandleWebhook(req)
	assert.EqualError(t, err, "invalid signature")

	// test with valid signature
	req = httptest.NewRequest("POST", "/signature", strings.NewReader(verifyPayloadSignatureTestPayload))
	req.Header.Set(appserver.HeaderPayloadSignature, verifyPayloadSignatureTestSignature)
	err = srv.HandleWebhook(req)
	assert.NoError(t, err)
}
