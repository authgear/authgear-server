package authtoken

type Store interface {
	NewToken(authInfoID string) (Token, error)
	Get(accessToken string, token *Token) error
	Put(token *Token) error
	Delete(accessToken string) error
}
