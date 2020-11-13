package appserver

import (
	"bytes"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	// Signature for payload: content
	verifyPayloadSignatureTestSignature = "e60478e5c6e5b8961b8fb0698cc0e7e9ce63ae3219874eb65c4a1bd2ec4c3bdd"
)

func TestVerifyPayloadSignature(t *testing.T) {
	srv := NewServer("", "", "mysecret")

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
	req = httptest.NewRequest("POST", "/signature", nil)
	req.Header.Set(HeaderPayloadSignature, verifyPayloadSignatureTestSignature)
	req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("content")))
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
