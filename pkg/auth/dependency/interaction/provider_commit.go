package interaction

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

func (p *Provider) Commit(i *Interaction) (*authn.Attrs, error) {
	var err error
	switch intent := i.Intent.(type) {
	case *IntentLogin:
		break
	case *IntentSignup:
		err = p.onCommitSignup(i, intent)
	default:
		panic(fmt.Sprintf("interaction: unknown intent type %T", i.Intent))
	}
	if err != nil {
		return nil, err
	}

	// Create identities & authenticators
	if err := p.Identity.CreateAll(i.UserID, i.NewIdentities); err != nil {
		return nil, err
	}
	if err := p.Authenticator.CreateAll(i.UserID, i.NewAuthenticators); err != nil {
		return nil, err
	}

	err = p.Store.Delete(i)
	if err != nil {
		p.Logger.WithError(err).Warn("failed to cleanup interaction")
	}

	attrs := &authn.Attrs{
		UserID:         i.UserID,
		IdentityType:   i.Identity.Type,
		IdentityClaims: i.Identity.Claims,
		// TODO(interaction): populate acr & amr
	}
	return attrs, nil
}

func (p *Provider) onCommitSignup(i *Interaction, intent *IntentSignup) error {
	if intent.OnUserDuplicate == model.OnUserDuplicateAbort {
		emailIdentities := map[string]struct{}{}
		for _, i := range i.NewIdentities {
			email, hasEmail := i.Claims[string(metadata.Email)].(string)
			if !hasEmail {
				continue
			}

			if _, exists := emailIdentities[email]; exists {
				return ErrDuplicatedIdentity
			}
			emailIdentities[email] = struct{}{}
		}

		for email := range emailIdentities {
			is, err := p.Identity.ListByClaims(map[string]string{string(metadata.Email): email})
			if err != nil {
				return err
			} else if len(is) > 0 {
				return ErrDuplicatedIdentity
			}
		}
	}

	err := p.User.Create(i.UserID, intent.UserMetadata, i.NewIdentities)
	if err != nil {
		return err
	}

	return nil
}
