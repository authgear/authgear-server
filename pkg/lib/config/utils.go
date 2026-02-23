package config

func newBool(v bool) *bool { return &v }

func newInt(v int) *int { return &v }

func IntersectAllowlist(appAllowlist []string, featureAllowlist []string) []string {
	if len(featureAllowlist) == 0 {
		return appAllowlist
	}

	featureMap := make(map[string]struct{})
	for _, a := range featureAllowlist {
		featureMap[a] = struct{}{}
	}

	var combinedAllowlist []string
	for _, a := range appAllowlist {
		if _, ok := featureMap[a]; ok {
			combinedAllowlist = append(combinedAllowlist, a)
		}
	}
	return combinedAllowlist
}
