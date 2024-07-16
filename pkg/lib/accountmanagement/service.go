package accountmanagement

import (
	"github.com/authgear/oauthrelyingparty/pkg/api/oauthrelyingparty"
)

type StartAddingInput struct {
	UserID                                          string
	Alias                                           string
	RedirectURI                                     string
	IncludeStateAuthorizationURLAndBindStateToToken bool
}

type StartAddingOutput struct {
	Token            string `json:"token,omitempty"`
	AuthorizationURL string `json:"authorization_url,omitempty"`
}

type FinishAddingInput struct {
	UserID string
	Token  string
	Query  string
}

type FinishAddingOutput struct {
	// It is intentionally empty.
}

type Store interface {
	GenerateToken(userID string, maybeState string) (string, error)
	ConsumeToken(tokenStr string) (*Token, error)
}

type OAuthProvider interface {
	GetProviderConfig(alias string) (oauthrelyingparty.ProviderConfig, error)
	GetAuthorizationURL(alias string, options oauthrelyingparty.GetAuthorizationURLOptions) (string, error)
}

type Service struct {
	Store         Store
	OAuthProvider OAuthProvider
}

func (s *Service) StartAdding(input *StartAddingInput) (*StartAddingOutput, error) {
	state := ""
	if input.IncludeStateAuthorizationURLAndBindStateToToken {
		state = GenerateRandomState()
	}

	token, err := s.Store.GenerateToken(input.UserID, state)
	if err != nil {
		return nil, err
	}

	param := oauthrelyingparty.GetAuthorizationURLOptions{
		RedirectURI: input.RedirectURI,
		State:       state,
	}

	authorizationURL, err := s.OAuthProvider.GetAuthorizationURL(input.Alias, param)
	if err != nil {
		return nil, err
	}

	return &StartAddingOutput{
		Token:            token,
		AuthorizationURL: authorizationURL,
	}, nil
}

func (s *Service) FinishAdding(input *FinishAddingInput) (*FinishAddingOutput, error) {
	token, err := s.Store.ConsumeToken(input.Token)
	if err != nil {
		return nil, err
	}

	err = token.CheckUser(input.UserID)
	if err != nil {
		return nil, err
	}

	state, err := ExtractStateFromQuery(input.Query)
	if err != nil {
		return nil, err
	}

	err = token.CheckState(state)
	if err != nil {
		return nil, err
	}

	// TODO: Do remaining things.
	return &FinishAddingOutput{}, nil
}
