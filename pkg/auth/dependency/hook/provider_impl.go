package hook

type providerImpl struct {
}

func NewProvider() Provider {
	return &providerImpl{}
}
