package mfa

type providerImpl struct {
	store Store
}

func NewProvider(store Store) Provider {
	return &providerImpl{
		store: store,
	}
}

func (p *providerImpl) GetRecoveryCode(userID string) ([]string, error) {
	aa, err := p.store.GetRecoveryCode(userID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, len(aa))
	for i, a := range aa {
		codes[i] = a.Code
	}
	return codes, nil
}

func (p *providerImpl) GenerateRecoveryCode(userID string) ([]string, error) {
	aa, err := p.store.GenerateRecoveryCode(userID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, len(aa))
	for i, a := range aa {
		codes[i] = a.Code
	}
	return codes, nil
}

func (p *providerImpl) ListAuthenticators(userID string) ([]interface{}, error) {
	authenticators, err := p.store.ListAuthenticators(userID)
	if err != nil {
		return nil, err
	}
	return MaskAuthenticators(authenticators), nil
}

var (
	_ Provider = &providerImpl{}
)
