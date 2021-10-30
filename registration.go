package appserver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
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
		return RegistrationResponse{}, SignatureVerificationError{err: errors.Wrap(err, "encode query")}
	}

	signature, err := hex.DecodeString(req.Header.Get(HeaderAppSignature))
	if err != nil {
		return RegistrationResponse{}, SignatureVerificationError{err: errors.Wrap(err, "decode signature")}
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

	err = srv.credentialStore.Store(&credentials)
	if err != nil {
		return RegistrationResponse{}, errors.Wrap(err, "store shop credentials")
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
		return errors.Wrap(err, "extract request body")
	}

	if len(body) == 0 {
		return errors.New("empty payload")
	}

	confirmReq := Credentials{}
	err = json.Unmarshal(body, &confirmReq)
	if err != nil {
		return errors.Wrap(err, "parse body")
	}

	credentials, err := srv.credentialStore.Get(confirmReq.ShopID)
	if err != nil {
		return errors.Wrap(err, "get shop credentials")
	}

	credentials.APIKey = confirmReq.APIKey
	credentials.SecretKey = confirmReq.SecretKey

	err = srv.credentialStore.Store(credentials)
	if err != nil {
		return errors.Wrap(err, "store shop credentials")
	}

	return nil
}
