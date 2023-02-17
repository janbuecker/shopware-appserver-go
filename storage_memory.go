package appserver

import (
	"sync"

	"golang.org/x/net/context"
)

var _ CredentialStore = (*MemoryCredentialStore)(nil)

type MemoryCredentialStore struct {
	credentials map[string]Credentials
	mapMu       sync.Mutex
}

func NewMemoryCredentialStore() *MemoryCredentialStore {
	return &MemoryCredentialStore{
		credentials: make(map[string]Credentials),
	}
}

func (m *MemoryCredentialStore) Store(ctx context.Context, credentials Credentials) error {
	m.mapMu.Lock()
	defer m.mapMu.Unlock()

	m.credentials[credentials.ShopID] = credentials

	return nil
}

func (m *MemoryCredentialStore) Get(ctx context.Context, shopID string) (Credentials, error) {
	if cred, ok := m.credentials[shopID]; ok {
		return cred, nil
	}

	return Credentials{}, ErrCredentialsNotFound
}

func (m *MemoryCredentialStore) Delete(ctx context.Context, shopID string) error {
	if _, ok := m.credentials[shopID]; !ok {
		return ErrCredentialsNotFound
	}

	m.mapMu.Lock()
	defer m.mapMu.Unlock()

	delete(m.credentials, shopID)

	return nil
}
