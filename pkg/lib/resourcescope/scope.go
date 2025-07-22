package resourcescope

import (
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
)

type newScope struct {
	Value string
}

func NewScope(ctx context.Context, str string) newScope {
	err := FormatScopeToken{}.CheckFormat(ctx, str)
	if err != nil {
		// This is a programming error because you should always validate the user input before calling NewScope
		panic(fmt.Errorf("invalid scope"))
	}
	return newScope{Value: str}
}

type NewScopeOptions struct {
	ResourceURI string
	Scope       newScope
	Description *string
}

type UpdateScopeOptions struct {
	ResourceURI string
	Scope       string
	NewDesc     *string
}

type ListScopeOptions struct {
	SearchKeyword string
	ClientID      string
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

type ListScopeResult struct {
	Items      []*model.Scope
	Offset     uint64
	TotalCount uint64
}
