package verification

import "time"

type Status string

const (
	StatusRequired Status = "required"
	StatusDisabled Status = "disabled"
	StatusPending  Status = "pending"
	StatusVerified Status = "verified"
)

type ClaimStatus struct {
	Name   string
	Status Status
}

type Claim struct {
	ID        string
	UserID    string
	Name      string
	Value     string
	CreatedAt time.Time
}

type claim struct {
	Name  string
	Value string
}
