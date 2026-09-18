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
	Scope       newScope
	Description *string
	// AccessPolicy is nil when the caller did not specify one, in which case
	// the scope is created with every key false (no access).
	AccessPolicy *model.AccessPolicyPatch
}

type UpdateScopeOptions struct {
	ResourceURI string
	Scope       string
	NewDesc     *string
	// AccessPolicy is nil when the caller did not specify one, in which case
	// the existing access policy is unchanged. A non-nil patch is merged
	// into the stored policy field by field; a field left nil in the patch
	// is unchanged even when other fields are set.
	AccessPolicy *model.AccessPolicyPatch
}

type ListScopeOptions struct {
	SearchKeyword string
	ClientID      string
}

type Scope struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ResourceID   string
	Scope        string
	Description  *string
	AccessPolicy model.AccessPolicy
}

func (s *Scope) ToModel() *model.Scope {
	return &model.Scope{
		Meta: model.Meta{
			ID:        s.ID,
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		},
		ResourceID:   s.ResourceID,
		Scope:        s.Scope,
		Description:  s.Description,
		AccessPolicy: s.AccessPolicy,
	}
}

type ListScopeResult struct {
	Items      []*model.Scope
	Offset     uint64
	TotalCount uint64
}
