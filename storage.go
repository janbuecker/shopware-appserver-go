package appserver

import (
	"errors"
	"golang.org/x/oauth2"
	"sync"
)

var (
	ErrCredentialsNotFound = errors.New("credentials for shop not found")
)

type CredentialStore interface {
	Store(credentials *Credentials) error
	Get(shopID string) (*Credentials, error)
	Delete(shopID string) error
}

type tokenStore struct {
	accessTokens   map[string]*oauth2.Token
	accessTokensMu sync.Mutex
}

func newTokenStore() *tokenStore {
	return &tokenStore{
		accessTokens: make(map[string]*oauth2.Token),
	}
}

func (s *tokenStore) Get(shopID string) (*oauth2.Token, bool) {
	if token, ok := s.accessTokens[shopID]; ok {
		return token, true
	}

	return nil, false
}

func (s *tokenStore) Store(shopID string, token *oauth2.Token) {
	s.accessTokensMu.Lock()
	defer s.accessTokensMu.Unlock()
	s.accessTokens[shopID] = token
}
