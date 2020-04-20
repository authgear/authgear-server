package interaction

import (
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/core/time"
)

// TODO(interaction): configurable lifetime
const interactionIdleTimeout = 5 * gotime.Minute

type Provider struct {
	Store Store
	Time  time.Provider
}

func (p *Provider) GetInteraction(token string) (*Interaction, error) {
	i, err := p.Store.Get(token)
	if err != nil {
		return nil, err
	}

	// TODO(interaction): populate identity & authenticator infos
	return i, nil
}

func (p *Provider) SaveInteraction(i *Interaction) (string, error) {
	if i.Token == "" {
		i.Token = generateToken()
		i.CreatedAt = p.Time.NowUTC()
		i.ExpireAt = i.CreatedAt.Add(interactionIdleTimeout)
		if err := p.Store.Create(i); err != nil {
			return "", err
		}
	} else {
		i.ExpireAt = p.Time.NowUTC().Add(interactionIdleTimeout)
		if err := p.Store.Update(i); err != nil {
			return "", err
		}
	}
	return i.Token, nil
}

func (p *Provider) Commit(i *Interaction) error {
	// TODO(interaction): do something
	return nil
}
