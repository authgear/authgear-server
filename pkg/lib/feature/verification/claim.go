package verification

import "time"

type ClaimStatus struct {
	Name                       string
	Verified                   bool
	RequiredToVerifyOnCreation bool
	EndUserTriggerable         bool
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

func (s ClaimStatus) IsVerifiable() bool {
	return s.RequiredToVerifyOnCreation || s.EndUserTriggerable
}
