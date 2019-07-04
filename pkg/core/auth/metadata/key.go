package metadata

type StandardKey string

const (
	Email StandardKey = "email"
	Phone StandardKey = "phone"
)

func AllKeys() []StandardKey {
	return []StandardKey{
		Email,
		Phone,
	}
}
