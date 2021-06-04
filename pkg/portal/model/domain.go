package model

import (
	"time"
)

// Domain represents a domain of an app.
// The keys in JSON struct tags are in camel case
// because this struct is directly returned in the GraphQL endpoint.
// Making the keys in camel case saves us from writing boilerplate resolver code.
type Domain struct {
	ID                    string    `json:"id"`
	AppID                 string    `json:"appID"`
	CreatedAt             time.Time `json:"createdAt"`
	Domain                string    `json:"domain"`
	ApexDomain            string    `json:"apexDomain"`
	VerificationDNSRecord string    `json:"verificationDNSRecord"`
	IsCustom              bool      `json:"isCustom"`
	IsVerified            bool      `json:"isVerified"`
}
