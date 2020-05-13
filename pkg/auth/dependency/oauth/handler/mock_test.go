package handler_test

import (
	"net/url"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/oauth/protocol"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/webapp"
)

type mockEndpointsProvider struct{}

func (mockEndpointsProvider) AuthorizeURI(r protocol.AuthorizationRequest) *url.URL {
	u, _ := url.Parse("https://auth/authorize")
	return u
}

func (mockEndpointsProvider) AuthenticateURI(opts webapp.AuthenticateURLOptions) *url.URL {
	u, _ := url.Parse("https://auth/authenticate")
	return u
}

type mockAuthzStore struct {
	authzs []oauth.Authorization
}

func (m *mockAuthzStore) Get(userID, clientID string) (*oauth.Authorization, error) {
	for _, a := range m.authzs {
		if a.UserID == userID && a.ClientID == clientID {
			return &a, nil
		}
	}
	return nil, oauth.ErrAuthorizationNotFound
}

func (m *mockAuthzStore) GetByID(id string) (*oauth.Authorization, error) {
	for _, a := range m.authzs {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, oauth.ErrAuthorizationNotFound
}

func (m *mockAuthzStore) Create(authz *oauth.Authorization) error {
	m.authzs = append(m.authzs, *authz)
	return nil
}

func (m *mockAuthzStore) Delete(authz *oauth.Authorization) error {
	n := 0
	for _, a := range m.authzs {
		if a.ID != authz.ID {
			m.authzs[n] = a
			n++
		}
	}
	m.authzs = m.authzs[:n]
	return nil
}

func (m *mockAuthzStore) UpdateScopes(authz *oauth.Authorization) error {
	for i, a := range m.authzs {
		if a.ID == authz.ID {
			a.Scopes = authz.Scopes
			a.UpdatedAt = authz.UpdatedAt
			m.authzs[i] = a
		}
	}
	return nil
}

type mockCodeGrantStore struct {
	grants []oauth.CodeGrant
}

func (m *mockCodeGrantStore) GetCodeGrant(codeHash string) (*oauth.CodeGrant, error) {
	for _, g := range m.grants {
		if g.CodeHash == codeHash {
			return &g, nil
		}
	}
	return nil, oauth.ErrGrantNotFound
}

func (m *mockCodeGrantStore) CreateCodeGrant(grant *oauth.CodeGrant) error {
	m.grants = append(m.grants, *grant)
	return nil
}

func (m *mockCodeGrantStore) DeleteCodeGrant(grant *oauth.CodeGrant) error {
	n := 0
	for _, g := range m.grants {
		if g.CodeHash != grant.CodeHash {
			m.grants[n] = g
			n++
		}
	}
	m.grants = m.grants[:n]
	return nil
}
