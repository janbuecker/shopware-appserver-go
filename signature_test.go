package appserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appserver "github.com/janbuecker/shopware-appserver-go"
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
	require.NoError(t, store.Store(context.Background(), appserver.Credentials{
		ShopID:     "123",
		ShopSecret: "mysecret",
	}))

	srv := appserver.NewServer("", "mysecret", "", appserver.WithCredentialStore(store))
	srv.Event("foo", func(_ context.Context, webhook appserver.WebhookRequest, api *appserver.APIClient) error {
		return nil
	})

	t.Run("without signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(verifyPayloadSignatureTestPayload))
		err := srv.HandleWebhook(req)
		assert.EqualError(t, err, "invalid signature")
	})

	// test with invalid signature
	t.Run("invalid signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/signature", strings.NewReader(verifyPayloadSignatureTestPayload))
		req.Header.Set(appserver.ShopSignatureKey, "foo")
		err := srv.HandleWebhook(req)
		assert.EqualError(t, err, "invalid signature")
	})

	// test with valid signature
	t.Run("valid signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/signature", strings.NewReader(verifyPayloadSignatureTestPayload))
		req.Header.Set(appserver.ShopSignatureKey, verifyPayloadSignatureTestSignature)
		err := srv.HandleWebhook(req)
		assert.NoError(t, err)
	})
}
