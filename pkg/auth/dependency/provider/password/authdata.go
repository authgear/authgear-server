package password

type authDataChecker interface {
	isValid(authData map[string]interface{}) bool
}

type defaultAuthDataChecker struct {
	authRecordKeys [][]string
}

func (c defaultAuthDataChecker) isValid(authData map[string]interface{}) bool {
	for dk := range authData {
		found := false
		for _, k := range c.allKeys() {
			if dk == k {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return len(c.usingKeys(authData)) > 0
}

func (c defaultAuthDataChecker) allKeys() []string {
	keyMap := map[string]bool{}
	for _, keys := range c.authRecordKeys {
		for _, key := range keys {
			keyMap[key] = true
		}
	}

	keys := []string{}
	for k := range keyMap {
		keys = append(keys, k)
	}

	return keys
}

func (c defaultAuthDataChecker) usingKeys(authData map[string]interface{}) []string {
	for _, ks := range c.authRecordKeys {
		count := 0
		for _, k := range ks {
			for dk := range authData {
				if k == dk && authData[dk] != nil {
					count = count + 1
				}
			}
		}

		if len(ks) == count {
			return ks
		}
	}

	return []string{}
}

// this ensures that our structure conform to certain interfaces.
var (
	_ authDataChecker = &defaultAuthDataChecker{}
)

// toValidAuthDataList converts authData to a list of authData depending on authRecordKeys
//
// example 1: authRecordKeys = [["username"], ["email"]]
// - if authData is { "username": "john.doe" }, output is [{ "username": "john.doe" }]
// - if authData is { "username": "john.doe", "email": "john.doe@example.com" }, output is [{ "username": "john.doe" }, { "email": "john.doe@example.com" }]
//
// example 2: authRecordKeys = [["username", "nickname"], ["email"]]
// - if authData is { "username": "john.doe", "email": "john.doe@example.com", "nickname": "john.doe" },
// output is [{ "username": "john.doe", "nickname": "john.doe" }, { "email": "john.doe@example.com" }]
//
// example 3: authRecordKeys = [["username", "email"], ["nickname"]]
// - if authData is { "username": "john.doe", "nickname": "john.doe" },
// output is [{ "nickname": "john.doe" }}]
func toValidAuthDataList(authRecordKeys [][]string, authData map[string]interface{}) []map[string]interface{} {
	outputs := make([]map[string]interface{}, 0)

	for _, ks := range authRecordKeys {
		m := make(map[string]interface{})
		for _, k := range ks {
			for dk := range authData {
				if k == dk && authData[dk] != nil {
					m[k] = authData[dk]
				}
			}
		}
		if len(m) == len(ks) {
			outputs = append(outputs, m)
		}
	}

	return outputs
}
