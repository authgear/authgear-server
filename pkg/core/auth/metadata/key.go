package metadata

// TODO: Merge with config.LoginIDKeyType
type StandardKey string

const (
	Email    StandardKey = "email"
	Phone    StandardKey = "phone"
	Username StandardKey = "username"
)

func AllKeys() []StandardKey {
	return []StandardKey{
		Email,
		Phone,
		Username,
	}
}
