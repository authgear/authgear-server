package verification

type Status string

const (
	StatusRequired Status = "required"
	StatusDisabled Status = "disabled"
	StatusPending  Status = "pending"
	StatusVerified Status = "verified"
)
