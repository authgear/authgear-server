package loader

type AuthzService interface {
	CheckAccessOfViewer(appID string) (userID string, err error)
}
