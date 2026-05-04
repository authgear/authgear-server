package config

//go:fix inline
func newBool(v bool) *bool { return new(v) }

//go:fix inline
func newFloat64(v float64) *float64 { return new(v) }

//go:fix inline
func newInt(v int) *int { return new(v) }

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
