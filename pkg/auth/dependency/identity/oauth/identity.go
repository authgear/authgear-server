package oauth

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/auth/config"
)

type Identity struct {
	ID                string
	UserID            string
	ProviderID        config.ProviderID
	ProviderSubjectID string
	UserProfile       map[string]interface{}
	Claims            map[string]interface{}
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
