package resourcescope

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type NewScopeOptions struct {
	ResourceID  string
	Scope       string
	Description *string
}

type UpdateScopeOptions struct {
	ID       string
	NewScope *string
	NewDesc  *string
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
