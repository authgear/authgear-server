package oauth

import (
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Identity struct {
	ID                string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	UserID            string
	ProviderID        config.ProviderID
	ProviderSubjectID string
	UserProfile       map[string]interface{}
	Claims            map[string]interface{}
}
