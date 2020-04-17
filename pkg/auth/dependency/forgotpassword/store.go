package forgotpassword

type Store interface {
	StoreCode(code *Code) error
	Get(codeStr string) (*Code, error)
}
