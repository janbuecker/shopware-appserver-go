package appserver

import (
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
	"testing"
)

func TestTokenStore_Store(t *testing.T) {
	store := newTokenStore()

	token := &oauth2.Token{AccessToken: "accesstoken"}
	store.Store("shopA", token)

	assert.Len(t, store.accessTokens, 1)
	assert.Equal(t, token, store.accessTokens["shopA"])

	newToken := &oauth2.Token{AccessToken: "newtoken"}
	store.Store("shopA", newToken)

	assert.Len(t, store.accessTokens, 1)
	assert.Equal(t, newToken, store.accessTokens["shopA"])
}

func TestTokenStore_Get(t *testing.T) {
	store := newTokenStore()

	token := &oauth2.Token{AccessToken: "accesstoken"}
	store.Store("shopA", token)

	assert.Len(t, store.accessTokens, 1)

	tk, ok := store.Get("shopA")
	if assert.True(t, ok) {
		assert.NotNil(t, tk)
		assert.Equal(t, token, tk)
	}

	// get unknown
	_, ok = store.Get("foo")
	assert.False(t, ok)
}
