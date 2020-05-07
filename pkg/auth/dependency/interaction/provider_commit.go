package interaction

import (
	"fmt"

	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth/metadata"
	"github.com/skygeario/skygear-server/pkg/core/authn"
)

func (p *Provider) Commit(i *Interaction) (*authn.Attrs, error) {
	if i.saved {
		panic("interaction: see NOTE(interaction): save-commit")
	}

	var err error
	switch intent := i.Intent.(type) {
	case *IntentLogin:
		err = p.onCommitLogin(i, intent)
	case *IntentSignup:
		err = p.onCommitSignup(i, intent)
	case *IntentAddIdentity:
		err = p.onCommitAddIdentity(i, intent, i.UserID)
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

	// Update identities
	if err := p.Identity.UpdateAll(i.UserID, i.UpdateIdentities); err != nil {
		return nil, err
	}

	identity, err := p.Identity.Get(i.UserID, i.Identity.Type, i.Identity.ID)
	if err != nil {
		return nil, err
	}

	err = p.Store.Delete(i)
	if err != nil {
		p.Logger.WithError(err).Warn("failed to cleanup interaction")
	}

	attrs := &authn.Attrs{
		UserID:         i.UserID,
		IdentityType:   identity.Type,
		IdentityClaims: identity.Claims,
		// TODO(interaction): populate acr & amr
	}

	i.committed = true

	return attrs, nil
}

func (p *Provider) onCommitLogin(i *Interaction, intent *IntentLogin) error {
	if intent.Identity.Type == authn.IdentityTypeOAuth {
		// skip update if login is triggered by signup
		if intent.OriginalIntentType == IntentTypeSignup {
			return nil
		}

		ii, err := p.Identity.Get(i.UserID, intent.Identity.Type, i.Identity.ID)
		if err != nil {
			p.Logger.WithError(err).Warn("failed to new identity for update")
			return err
		}
		ui := p.Identity.WithClaims(i.UserID, ii, intent.Identity.Claims)
		i.UpdateIdentities = append(i.UpdateIdentities, ui)
	}

	return nil
}

func (p *Provider) onCommitSignup(i *Interaction, intent *IntentSignup) error {
	// TODO(interaction-sso): handle OnUserDuplicateMerge
	if intent.OnUserDuplicate == model.OnUserDuplicateAbort {
		err := p.checkIdentitiesDuplicated(i.NewIdentities)
		if err != nil {
			return err
		}
	}

	err := p.User.Create(i.UserID, intent.UserMetadata, i.NewIdentities)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provider) onCommitAddIdentity(i *Interaction, intent *IntentAddIdentity, userID string) error {
	err := p.checkIdentitiesDuplicated(i.NewIdentities)
	if err != nil {
		return err
	}

	user, err := p.User.Get(userID)
	if err != nil {
		return err
	}

	for _, i := range i.NewIdentities {
		identity := model.Identity{
			Type:   string(i.Type),
			Claims: i.Claims,
		}
		err = p.Hooks.DispatchEvent(
			event.IdentityCreateEvent{
				User:     *user,
				Identity: identity,
			},
			user,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provider) checkIdentitiesDuplicated(iis []*IdentityInfo) error {
	emailIdentities := map[string]struct{}{}
	for _, i := range iis {
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

	return nil
}
