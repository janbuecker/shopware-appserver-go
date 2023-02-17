package appserver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/thanhpk/randstr"
)

type RegistrationResponse struct {
	Proof           string `json:"proof"`
	Secret          string `json:"secret"`
	ConfirmationURL string `json:"confirmation_url"`
}

func (srv Server) HandleRegistration(req *http.Request) (RegistrationResponse, error) {
	query, err := url.QueryUnescape(req.URL.Query().Encode())
	if err != nil {
		return RegistrationResponse{}, SignatureVerificationError{err: fmt.Errorf("encode query: %w", err)}
	}

	signature, err := hex.DecodeString(req.Header.Get(HeaderAppSignature))
	if err != nil {
		return RegistrationResponse{}, SignatureVerificationError{err: fmt.Errorf("decode signature: %w", err)}
	}

	if err := verifySignature([]byte(query), signature, srv.appSecret); err != nil {
		return RegistrationResponse{}, SignatureVerificationError{err: err}
	}

	credentials := Credentials{
		Timestamp: req.URL.Query().Get("timestamp"),
		ShopURL:   req.URL.Query().Get("shop-url"),
		ShopID:    req.URL.Query().Get("shop-id"),
	}

	h := hmac.New(sha256.New, []byte(srv.appSecret))
	h.Write([]byte(credentials.ShopID + credentials.ShopURL + srv.appName))

	credentials.ShopSecret = randstr.Base62(16)

	err = srv.credentialStore.Store(req.Context(), credentials)
	if err != nil {
		return RegistrationResponse{}, fmt.Errorf("store shop credentials: %w", err)
	}

	return RegistrationResponse{
		Secret:          credentials.ShopSecret,
		Proof:           hex.EncodeToString(h.Sum(nil)),
		ConfirmationURL: srv.confirmationURL,
	}, nil
}

func (srv Server) HandleConfirm(req *http.Request) error {
	body, err := extractBody(req)
	if err != nil {
		return fmt.Errorf("extract request body: %w", err)
	}

	if len(body) == 0 {
		return errors.New("empty payload")
	}

	confirmReq := Credentials{}
	err = json.Unmarshal(body, &confirmReq)
	if err != nil {
		return fmt.Errorf("parse body: %w", err)
	}

	credentials, err := srv.credentialStore.Get(req.Context(), confirmReq.ShopID)
	if err != nil {
		return fmt.Errorf("get shop credentials: %w", err)
	}

	credentials.APIKey = confirmReq.APIKey
	credentials.SecretKey = confirmReq.SecretKey

	err = srv.credentialStore.Store(req.Context(), credentials)
	if err != nil {
		return fmt.Errorf("store shop credentials: %w", err)
	}

	return nil
}
