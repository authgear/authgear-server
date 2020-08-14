package oauth

type GrantSessionKind string

const (
	GrantSessionKindOffline GrantSessionKind = "offline_grant"
	GrantSessionKindSession GrantSessionKind = "idp_session"
)

type Grant interface {
	Session() (kind GrantSessionKind, id string)
}
