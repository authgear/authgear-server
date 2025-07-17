package resourcescope

import (
	"context"
	"regexp"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/oauth"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

type NewScopeOptions struct {
	ResourceURI string
	Scope       string
	Description *string
}

type UpdateScopeOptions struct {
	ResourceURI string
	Scope       string
	NewDesc     *string
}

type Scope struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ResourceID  string
	Scope       string
	Description *string
}

func (s *Scope) ToModel() *model.Scope {
	return &model.Scope{
		Meta: model.Meta{
			ID:        s.ID,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		},
		ResourceID:  s.ResourceID,
		Scope:       s.Scope,
		Description: s.Description,
	}
}

func ValidateScope(ctx context.Context, scope string) error {
	blacklist := oauth.AllowedScopes

	validationCtx := &validation.Context{}

	// See https://datatracker.ietf.org/doc/html/rfc6749#section-3.3
	tokenRe := regexp.MustCompile(`^[\x21\x23-\x5B\x5D-\x7E]+$`)
	if !tokenRe.MatchString(scope) {
		validationCtx.EmitError("format", map[string]interface{}{"error": "invalid scope", "scope": scope})
	}

	for _, blacklisted := range blacklist {
		if blacklisted == scope {
			validationCtx.EmitError("blocked", map[string]interface{}{"reason": "ReservedScope", "scope": scope})
		}
	}

	return validationCtx.Error("invalid scope")
}
