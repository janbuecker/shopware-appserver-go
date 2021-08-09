package appserver

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	// Signature for payload: {}
	verifyPayloadSignatureTestSignature = "5c4355936c2e25efca79f5e84a163b080a74f28bc185c159f7e0d9e3f40d47bc"
)

type mockCredendtialStore struct {
	credentials Credentials
}

func (m *mockCredendtialStore) Store(credentials *Credentials) error {
	return nil
}

func (m *mockCredendtialStore) Get(shopID string) (*Credentials, error) {
	return &m.credentials, nil
}

func (m *mockCredendtialStore) Delete(shopID string) error {
	return nil
}

func TestVerifyPayloadSignature(t *testing.T) {
	srv := NewServer(
		"",
		"",
		"mysecret",
		WithCredentialStore(&mockCredendtialStore{
			credentials: Credentials{
				APIKey:     "mysecret",
				SecretKey:  "mysecret",
				Timestamp:  "",
				ShopURL:    "",
				ShopID:     "",
				ShopSecret: "mysecret",
			},
		},
		),
	)

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.NoContent(200)
	})

	e.POST("/signature", func(c echo.Context) error {
		return c.NoContent(200)
	}, srv.verifyPayloadSignature())

	// test not matching
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// test without signature
	req = httptest.NewRequest("POST", "/signature", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, `{"message":"invalid signature"}`, strings.Trim(rec.Body.String(), "\n"))

	// test with invalid signature
	req = httptest.NewRequest("POST", "/signature", nil)
	req.Header.Set(HeaderPayloadSignature, "foo")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, `{"message":"invalid signature"}`, strings.Trim(rec.Body.String(), "\n"))

	// test with valid signature
	req = httptest.NewRequest("POST", "/signature", strings.NewReader(`{}`))
	req.Header.Set(HeaderPayloadSignature, verifyPayloadSignatureTestSignature)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
