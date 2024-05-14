package internalinterface

type LoginIDNormalizer interface {
	Normalize(loginID string) (string, error)
	ComputeUniqueKey(normalizeLoginID string) (string, error)
}
