package appserver

import "sync"

var _ CredentialStore = (*MemoryCredentialStore)(nil)

type MemoryCredentialStore struct {
	credentials map[string]*Credentials
	mapMu       sync.Mutex
}

func NewMemoryCredentialStore() CredentialStore {
	return &MemoryCredentialStore{
		credentials: make(map[string]*Credentials),
	}
}

func (m *MemoryCredentialStore) Store(credentials *Credentials) error {
	m.mapMu.Lock()
	defer m.mapMu.Unlock()

	m.credentials[credentials.ShopID] = credentials

	return nil
}

func (m *MemoryCredentialStore) Get(shopID string) (*Credentials, error) {
	if cred, ok := m.credentials[shopID]; ok {
		return cred, nil
	}

	return nil, ErrCredentialsNotFound
}

func (m *MemoryCredentialStore) Delete(shopID string) error {
	if _, ok := m.credentials[shopID]; !ok {
		return ErrCredentialsNotFound
	}

	m.mapMu.Lock()
	defer m.mapMu.Unlock()

	delete(m.credentials, shopID)

	return nil
}
