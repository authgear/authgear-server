package verification

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type ClaimStatus struct {
	Name                       string
	Value                      string
	Verified                   bool
	RequiredToVerifyOnCreation bool
	EndUserTriggerable         bool
	VerifiedByChannel          model.AuthenticatorOOBChannel
}

type Claim struct {
	ID        string
	UserID    string
	Name      string
	Value     string
	CreatedAt time.Time
	Metadata  map[string]interface{}
}

type claim struct {
	Name  string
	Value string
}

func (s ClaimStatus) IsVerifiable() bool {
	return s.RequiredToVerifyOnCreation || s.EndUserTriggerable
}

func (c *Claim) VerifiedByChannel() model.AuthenticatorOOBChannel {
	if c.Metadata == nil {
		return ""
	}
	if verifiedByChannel, ok := c.Metadata["verified_by_channel"].(string); ok {
		return model.AuthenticatorOOBChannel(verifiedByChannel)
	}
	return ""
}

func (c *Claim) SetVerifiedByChannel(verifiedByChannel model.AuthenticatorOOBChannel) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]interface{})
	}
	if verifiedByChannel == "" {
		c.Metadata["verified_by_channel"] = nil
	} else {
		c.Metadata["verified_by_channel"] = string(verifiedByChannel)
	}
}
