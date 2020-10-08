package model

import (
	"time"
)

type Domain struct {
	ID                    string    `json:"id"`
	CreatedAt             time.Time `json:"createdAt"`
	Domain                string    `json:"domain"`
	ApexDomain            string    `json:"apexDomain"`
	VerificationDNSRecord string    `json:"verificationDNSRecord"`
	IsVerified            bool      `json:"isVerified"`
}
