package userprofile

type MockUserProfileStoreImpl struct {
	Data map[string]map[string]interface{}
}

func NewMockUserProfileStore() *MockUserProfileStoreImpl {
	return &MockUserProfileStoreImpl{
		Data: map[string]map[string]interface{}{},
	}
}

func (u MockUserProfileStoreImpl) CreateUserProfile(userID string, userProfile map[string]interface{}) (err error) {
	u.Data[userID] = make(map[string]interface{})
	u.Data[userID] = userProfile
	return
}

func (u MockUserProfileStoreImpl) GetUserProfile(userID string, userProfile *map[string]interface{}) (err error) {
	*userProfile = make(map[string]interface{})
	data := u.Data[userID]
	for k := range data {
		(*userProfile)[k] = data[k]
	}
	(*userProfile)["_id"] = "user/" + userID
	(*userProfile)["_type"] = "record"
	(*userProfile)["_recordID"] = userID
	(*userProfile)["_recordType"] = "user"
	(*userProfile)["_access"] = nil
	(*userProfile)["_ownerID"] = userID
	(*userProfile)["_created_by"] = userID
	(*userProfile)["_updated_by"] = userID
	return
}
