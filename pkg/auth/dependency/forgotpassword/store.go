package forgotpassword

type Store interface {
	Create(code *Code) error
	Get(codeStr string) (*Code, error)
	Update(code *Code) error
}
