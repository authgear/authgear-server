package userprofile

type MockUserProfileStoreImpl struct {
	Data map[string]map[string]interface{}
}

func NewMockUserProfileStore() *MockUserProfileStoreImpl {
	return &MockUserProfileStoreImpl{
		Data: map[string]map[string]interface{}{},
	}
}

func (u MockUserProfileStoreImpl) CreateUserProfile(userID string, data Data) (profile UserProfile, err error) {
	u.Data[userID] = make(map[string]interface{})
	u.Data[userID] = data
	profile = toProfile(userID, data)
	return
}

func (u MockUserProfileStoreImpl) GetUserProfile(userID string) (profile UserProfile, err error) {
	data := u.Data[userID]
	profile = toProfile(userID, data)
	return
}

func toProfile(userID string, data Data) map[string]interface{} {
	profile := make(map[string]interface{})

	profile["_id"] = "user/" + userID
	profile["_type"] = "record"
	profile["_recordID"] = userID
	profile["_recordType"] = "user"
	profile["_access"] = nil
	profile["_ownerID"] = userID
	profile["_created_by"] = userID
	profile["_updated_by"] = userID
	for k, v := range data {
		profile[k] = v
	}

	return profile
}
