package appserver

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMemoryCredentialStore_Store(t *testing.T) {
	store := NewMemoryCredentialStore().(*MemoryCredentialStore)

	cred := &Credentials{
		APIKey:    "foo",
		SecretKey: "bar",
		Timestamp: time.Now().Format(time.RFC3339),
		ShopURL:   "https://shopware.com",
		ShopID:    "aBCd21EF",
	}
	err := store.Store(cred)

	if assert.NoError(t, err) {
		assert.Len(t, store.credentials, 1)
		assert.Equal(t, cred, store.credentials[cred.ShopID])
	}

	credOverwrite := &Credentials{
		APIKey:    "newkey",
		SecretKey: "newsecret",
		Timestamp: time.Now().Format(time.RFC3339),
		ShopURL:   "https://newURL.com",
		ShopID:    "aBCd21EF",
	}
	err = store.Store(credOverwrite)

	if assert.NoError(t, err) {
		assert.Len(t, store.credentials, 1)
		assert.Equal(t, credOverwrite, store.credentials[cred.ShopID])
	}
}

func TestMemoryCredentialStore_Delete(t *testing.T) {
	store := NewMemoryCredentialStore().(*MemoryCredentialStore)

	cred := &Credentials{
		APIKey:    "foo",
		SecretKey: "bar",
		Timestamp: time.Now().Format(time.RFC3339),
		ShopURL:   "https://shopware.com",
		ShopID:    "aBCd21EF",
	}
	err := store.Store(cred)

	if assert.NoError(t, err) {
		assert.Len(t, store.credentials, 1)
	}

	err = store.Delete(cred.ShopID)
	if assert.NoError(t, err) {
		assert.Len(t, store.credentials, 0)
	}

	// delete unknown key
	err = store.Delete(cred.ShopID)
	assert.EqualError(t, err, ErrCredentialsNotFound.Error())
}

func TestMemoryCredentialStore_Get(t *testing.T) {
	store := NewMemoryCredentialStore().(*MemoryCredentialStore)

	cred := &Credentials{
		APIKey:    "foo",
		SecretKey: "bar",
		Timestamp: time.Now().Format(time.RFC3339),
		ShopURL:   "https://shopware.com",
		ShopID:    "aBCd21EF",
	}
	err := store.Store(cred)

	if assert.NoError(t, err) {
		assert.Len(t, store.credentials, 1)
	}

	c, err := store.Get(cred.ShopID)
	if assert.NoError(t, err) {
		assert.NotNil(t, c)
		assert.Equal(t, c, cred)
	}

	// get unknown
	_, err = store.Get("foo")
	assert.EqualError(t, err, ErrCredentialsNotFound.Error())
}
