package interaction

func IsAdminAPI(input interface{}) bool {
	isAdminAPI := false
	var adminInput interface{ IsAdminAPI() bool }
	if Input(input, &adminInput) {
		isAdminAPI = adminInput.IsAdminAPI()
	}
	return isAdminAPI
}
