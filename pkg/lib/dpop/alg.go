package dpop

import "slices"

var SupportedAlgorithms = []string{"ES256", "RS256"}

func IsSupportedAlgorithms(alg string) bool {
	return slices.Contains(SupportedAlgorithms, alg)
}
