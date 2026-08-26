package dcr

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/oauthclient"
)

// RegisterClientOptions is built by the registration HTTP handler after it
// has determined the caller's IAT-derived Kind (FIRST_PARTY/THIRD_PARTY,
// or THIRD_PARTY under open registration) and validated the request body
// via ValidateAndNormalize.
type RegisterClientOptions struct {
	Kind         oauthclient.Kind
	Registration *NormalizedRegistration
}

// RegisterClient generates the dcrc_ client_id, builds an
// oauthclient.Client with Source: DCR, and persists it via
// oauthclient.Commands.CreateClient.
func (c *Commands) RegisterClient(ctx context.Context, options *RegisterClientOptions) (*model.OAuthClient, error) {
	clientID := oauthclient.GenerateDCRClientID()

	client, err := c.OAuthClient.CreateClient(ctx, &oauthclient.NewClientOptions{
		ClientID:        clientID,
		Source:          model.OAuthClientSourceDCR,
		Kind:            options.Kind,
		ApplicationType: options.Registration.ApplicationType,
		ClientName:      options.Registration.ClientName,
		ClientURI:       options.Registration.ClientURI,
		LogoURI:         options.Registration.LogoURI,
		TOSURI:          options.Registration.TOSURI,
		PolicyURI:       options.Registration.PolicyURI,
		RedirectURIs:    options.Registration.RedirectURIs,
		GrantTypes:      options.Registration.GrantTypes,
		ResponseTypes:   options.Registration.ResponseTypes,
	})
	if err != nil {
		return nil, err
	}

	tokenLifetimes := oauthclient.ResolveTokenLifetimes(c.OAuthConfig, model.OAuthClientSourceDCR)
	return client.ToModel(tokenLifetimes), nil
}
