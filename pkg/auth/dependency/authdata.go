package dependency

type AuthDataChecker interface {
	IsValid(authData map[string]interface{}) bool
}

type DefaultAuthDataChecker struct {
	AuthRecordKeys [][]string
}

func (c DefaultAuthDataChecker) IsValid(authData map[string]interface{}) bool {
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

func (c DefaultAuthDataChecker) allKeys() []string {
	keyMap := map[string]bool{}
	for _, keys := range c.AuthRecordKeys {
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

func (c DefaultAuthDataChecker) usingKeys(authData map[string]interface{}) []string {
	for _, ks := range c.AuthRecordKeys {
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
	_ AuthDataChecker = &DefaultAuthDataChecker{}
)
