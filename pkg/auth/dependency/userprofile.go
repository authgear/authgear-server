package dependency

type UserProfileStore interface {
	CreateUserProfile(userProfile interface{}) error
	GetUserProfile(userID string, userProfile *interface{}) error
}
