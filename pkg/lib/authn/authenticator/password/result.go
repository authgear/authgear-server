package password

type VerifyResult struct {
	PolicyForceChange bool
	ExpiryForceChange bool
}

func (r *VerifyResult) RequireUpdate() bool {
	return r.PolicyForceChange || r.ExpiryForceChange
}
