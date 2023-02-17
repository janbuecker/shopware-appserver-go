package appserver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCredentialStore_Store(t *testing.T) {
	store := NewMemoryCredentialStore()

	cred := Credentials{
		APIKey:    "foo",
		SecretKey: "bar",
		Timestamp: time.Now().Format(time.RFC3339),
		ShopURL:   "https://shopware.com",
		ShopID:    "aBCd21EF",
	}
	err := store.Store(context.Background(), cred)

	if assert.NoError(t, err) {
		assert.Len(t, store.credentials, 1)
		assert.Equal(t, cred, store.credentials[cred.ShopID])
	}

	credOverwrite := Credentials{
		APIKey:    "newkey",
		SecretKey: "newsecret",
		Timestamp: time.Now().Format(time.RFC3339),
		ShopURL:   "https://newURL.com",
		ShopID:    "aBCd21EF",
	}
	err = store.Store(context.Background(), credOverwrite)

	if assert.NoError(t, err) {
		assert.Len(t, store.credentials, 1)
		assert.Equal(t, credOverwrite, store.credentials[cred.ShopID])
	}
}

func TestMemoryCredentialStore_Delete(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryCredentialStore()

	cred := Credentials{
		APIKey:    "foo",
		SecretKey: "bar",
		Timestamp: time.Now().Format(time.RFC3339),
		ShopURL:   "https://shopware.com",
		ShopID:    "aBCd21EF",
	}
	err := store.Store(ctx, cred)

	if assert.NoError(t, err) {
		assert.Len(t, store.credentials, 1)
	}

	err = store.Delete(ctx, cred.ShopID)
	if assert.NoError(t, err) {
		assert.Len(t, store.credentials, 0)
	}

	// delete unknown key
	err = store.Delete(ctx, cred.ShopID)
	assert.EqualError(t, err, ErrCredentialsNotFound.Error())
}

func TestMemoryCredentialStore_Get(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryCredentialStore()

	cred := Credentials{
		APIKey:    "foo",
		SecretKey: "bar",
		Timestamp: time.Now().Format(time.RFC3339),
		ShopURL:   "https://shopware.com",
		ShopID:    "aBCd21EF",
	}
	err := store.Store(ctx, cred)

	if assert.NoError(t, err) {
		assert.Len(t, store.credentials, 1)
	}

	c, err := store.Get(ctx, cred.ShopID)
	if assert.NoError(t, err) {
		assert.NotNil(t, c)
		assert.Equal(t, c, cred)
	}

	// get unknown
	_, err = store.Get(ctx, "foo")
	assert.EqualError(t, err, ErrCredentialsNotFound.Error())
}
