package dependency

type MockUserProfileStoreImpl struct{}

func NewMockUserProfileStore() *MockUserProfileStoreImpl {
	return &MockUserProfileStoreImpl{}
}

func (u MockUserProfileStoreImpl) CreateUserProfile(userID string, userProfile map[string]interface{}) (err error) {
	return
}

func (u MockUserProfileStoreImpl) GetUserProfile(userID string, userProfile *map[string]interface{}) (err error) {
	return
}
