package dpop

var SupportedAlgorithms = []string{"ES256", "RS256"}

func IsSupportedAlgorithms(alg string) bool {
	for _, sa := range SupportedAlgorithms {
		if sa == alg {
			return true
		}
	}
	return false
}
