package loader

type AuthzService interface {
	CheckAccessOfViewer(appID string) error
}
