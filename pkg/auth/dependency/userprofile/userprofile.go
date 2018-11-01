package userprofile

type UserProfileStore interface {
	CreateUserProfile(userID string, userProfile map[string]interface{}) error
	GetUserProfile(userID string, userProfile *map[string]interface{}) error
}
