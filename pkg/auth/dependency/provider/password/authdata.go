package password

type authDataChecker interface {
	isValid(authData map[string]string) bool
	isMatching(authData map[string]string) bool
}

type defaultAuthDataChecker struct {
	authRecordKeys [][]string
}

func (c defaultAuthDataChecker) isValid(authData map[string]string) bool {
	return len(toValidAuthDataList(c.authRecordKeys, authData)) > 0
}

func (c defaultAuthDataChecker) isMatching(authData map[string]string) bool {
	// authData requires exactly match to current authRecordKeys setting
	// if authRecordKeys is [["username"], ["email"]]
	// it will match authData is {"username": "someusername"} or {"email": "someemail@example.com"}
	// and will not match authData is {"username": "someusername", "email": "someemail@example.com"}
	for _, authRecordKeys := range c.authRecordKeys {
		if len(authRecordKeys) != len(authData) {
			continue
		}
		matched := true
		for _, key := range authRecordKeys {
			if _, ok := authData[key]; !ok {
				matched = false
				break
			}
		}
		if matched {
			return matched
		}
	}

	return false
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
//
// example 4: authRecordKeys = [["username"], ["email"]]
// - if authData is { "username": "john.doe", "emamil": "" },
// output is [{ "username": "john.doe" }}]
func toValidAuthDataList(authRecordKeys [][]string, authData map[string]string) []map[string]string {
	outputs := make([]map[string]string, 0)

	for _, ks := range authRecordKeys {
		m := make(map[string]string)
		for _, k := range ks {
			for dk := range authData {
				if k == dk && authData[dk] != "" {
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
