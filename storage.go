package appserver

import (
	"context"
	"errors"
	"sync"

	"golang.org/x/oauth2"
)

var ErrCredentialsNotFound = errors.New("credentials for shop not found")

type CredentialStore interface {
	Store(ctx context.Context, credentials Credentials) error
	Get(ctx context.Context, shopID string) (Credentials, error)
	Delete(ctx context.Context, shopID string) error
}

type tokenStore struct {
	accessTokens   map[string]*oauth2.Token
	accessTokensMu sync.RWMutex
}

func newTokenStore() *tokenStore {
	return &tokenStore{
		accessTokens: make(map[string]*oauth2.Token),
	}
}

func (s *tokenStore) Get(shopID string) (*oauth2.Token, bool) {
	s.accessTokensMu.RLock()
	defer s.accessTokensMu.RUnlock()
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
