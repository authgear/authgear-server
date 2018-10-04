package dependency

type UserProfileStore interface {
	CreateUserProfile(userProfile interface{}) error
}
