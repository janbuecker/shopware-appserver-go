package appserver

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type SignatureVerificationError struct {
	err error
}

func (e SignatureVerificationError) Error() string {
	return "invalid signature"
}

func (e SignatureVerificationError) Unwrap() error {
	return e.err
}

func (srv *Server) verifyPayloadSignature(req *http.Request) error {
	body, err := extractBody(req)
	if err != nil {
		return SignatureVerificationError{err: fmt.Errorf("extract request body: %w", err)}
	}

	if len(body) == 0 {
		return SignatureVerificationError{err: errors.New("empty payload")}
	}

	// copy body back to the request
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	appReq := AppRequest{}
	if err := json.Unmarshal(body, &appReq); err != nil {
		return SignatureVerificationError{err: fmt.Errorf("parse body: %w", err)}
	}

	credentials, err := srv.credentialStore.Get(req.Context(), appReq.Source.ShopID)
	if err != nil {
		return SignatureVerificationError{err: fmt.Errorf("get shop credentials: %w", err)}
	}

	signature, err := hex.DecodeString(req.Header.Get(ShopSignatureKey))
	if err != nil {
		return SignatureVerificationError{err: fmt.Errorf("decode signature: %w", err)}
	}

	if err := verifySignature(body, signature, credentials.ShopSecret); err != nil {
		return SignatureVerificationError{err: err}
	}

	return nil
}

func (srv *Server) verifyQuerySignature(req *http.Request) error {
	shopID := req.URL.Query().Get("shop-id")
	if shopID == "" {
		return SignatureVerificationError{err: errors.New("missing query parameter: shop-id")}
	}

	signature, err := hex.DecodeString(req.URL.Query().Get(ShopSignatureKey))
	if err != nil {
		return SignatureVerificationError{err: fmt.Errorf("decode signature: %w", err)}
	}

	// Go sorts the query internally by key, so it's hard to replicate Shopware's behaviour here.
	// The signature needs be calculated on the raw query string, because that's what Shopware does and expects. The
	// tests in Shopware indicate, that the signature key/value should just be replaced with an empty string. It works.
	query := strings.ReplaceAll(req.URL.RawQuery, fmt.Sprintf("&%s=%s", ShopSignatureKey, req.URL.Query().Get(ShopSignatureKey)), "")
	query, err = url.QueryUnescape(query)
	if err != nil {
		return SignatureVerificationError{err: fmt.Errorf("encode query: %w", err)}
	}

	credentials, err := srv.credentialStore.Get(req.Context(), shopID)
	if err != nil {
		return SignatureVerificationError{err: fmt.Errorf("get shop credentials: %w", err)}
	}

	if err := verifySignature([]byte(query), signature, credentials.ShopSecret); err != nil {
		return SignatureVerificationError{err: err}
	}

	return nil
}

func verifySignature(data []byte, signature []byte, key string) error {
	if len(data) == 0 {
		return errors.New("empty data")
	}

	if len(signature) == 0 {
		return errors.New("empty signature")
	}

	if key == "" {
		return errors.New("empty key")
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)

	if !hmac.Equal(signature, h.Sum(nil)) {
		return errors.New("signature mismatch")
	}

	return nil
}
