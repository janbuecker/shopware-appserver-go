package appserver

import (
	"net/http"
)

func (srv *Server) VerifyPageSignature(req *http.Request) error {
	if err := srv.verifyQuerySignature(req); err != nil {
		return err
	}

	return nil
}
