package appserver

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
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
		return SignatureVerificationError{err: errors.Wrap(err, "extract request body")}
	}

	if len(body) == 0 {
		return SignatureVerificationError{err: errors.New("empty payload")}
	}

	// copy body back to the request
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	appReq := AppRequest{}
	if err := json.Unmarshal(body, &appReq); err != nil {
		return SignatureVerificationError{err: errors.Wrap(err, "parse body")}
	}

	credentials, err := srv.credentialStore.Get(appReq.Source.ShopID)
	if err != nil {
		return SignatureVerificationError{err: errors.Wrap(err, "get shop credentials")}
	}

	signature, err := hex.DecodeString(req.Header.Get(HeaderPayloadSignature))
	if err != nil {
		return SignatureVerificationError{err: errors.Wrap(err, "decode signature")}
	}

	if err := verifySignature(body, signature, credentials.ShopSecret); err != nil {
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
