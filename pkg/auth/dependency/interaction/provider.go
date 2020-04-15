package interaction

type Provider struct {
}

func (p *Provider) GetInteraction(token string) (*Interaction, error) {
	// TODO(interaction): do something
	return nil, nil
}

func (p *Provider) SaveInteraction(i *Interaction) (string, error) {
	// TODO(interaction): do something
	return "", nil
}

func (p *Provider) Commit(i *Interaction) error {
	// TODO(interaction): do something
	return nil
}
