package rolesgroups

func computeKeyDifference(originalKeys []string, incomingKeys []string) (keysToAdd []string, keysToRemove []string) {
	keysMap := make(map[string]int)
	// -1: delete, 0: no ops, 1: add
	for _, v := range originalKeys {
		keysMap[v] = -1
	}
	for _, v := range incomingKeys {
		if keysMap[v] == -1 {
			keysMap[v] = 0
		} else {
			keysMap[v] = 1
		}
	}

	for k, v := range keysMap {
		if v == -1 {
			keysToRemove = append(keysToRemove, k)
		}
		if v == 1 {
			keysToAdd = append(keysToAdd, k)
		}
	}

	return keysToAdd, keysToRemove
}
