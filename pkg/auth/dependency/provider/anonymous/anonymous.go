package anonymous

const providerAnonymous string = "anonymous"

type Provider interface {
	CreatePrincipal(principal Principal) error
}
