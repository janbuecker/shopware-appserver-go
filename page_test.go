package appserver_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	appserver "github.com/janbuecker/shopware-appserver-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const verifyPageSignatureTestSignature = "eb70ab79972ec497c6a23e24d202f005b371868445985eae53a25af2bf93f4e0"

func TestServer_VerifyPageSignature(t *testing.T) {
	store := appserver.NewMemoryCredentialStore()
	require.NoError(t, store.Store(context.Background(), appserver.Credentials{
		ShopID:     "123",
		ShopSecret: "mysecret",
	}))

	srv := appserver.NewServer("", "mysecret", "", appserver.WithCredentialStore(store))

	t.Run("without signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/page/foo?shop-id=123&timestamp=1234567890", nil)
		err := srv.VerifyPageSignature(req)
		assert.EqualError(t, err, "invalid signature")
	})

	// test with invalid signature
	t.Run("invalid signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/page/foo?shop-id=123&timestamp=1234567890&shopware-shop-signature=foo", nil)
		err := srv.VerifyPageSignature(req)
		assert.EqualError(t, err, "invalid signature")
	})

	// test with valid signature
	t.Run("valid signature", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/page/foo?shop-id=123&timestamp=1234567890&shopware-shop-signature="+verifyPageSignatureTestSignature, nil)
		err := srv.VerifyPageSignature(req)
		assert.NoError(t, err)
	})
}
