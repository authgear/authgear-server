package latte

import (
	"context"
	"errors"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func init() {
	workflow.RegisterNode(&NodeDoUpdateIdentity{})
}

type NodeDoUpdateIdentity struct {
	IdentityBeforeUpdate *identity.Info `json:"identity_before_update"`
	IdentityAfterUpdate  *identity.Info `json:"identity_after_update"`
	IsAdminAPI           bool           `json:"is_admin_api"`
}

func (n *NodeDoUpdateIdentity) Kind() string {
	return "latte.NodeDoUpdateIdentity"
}

func (n *NodeDoUpdateIdentity) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.RunEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			if _, err := deps.Identities.CheckDuplicated(n.IdentityAfterUpdate); err != nil {
				if errors.Is(err, identity.ErrIdentityAlreadyExists) {
					s1 := n.IdentityBeforeUpdate.ToSpec()
					s2 := n.IdentityAfterUpdate.ToSpec()
					return identityFillDetails(api.ErrDuplicatedIdentity, &s2, &s1)
				}
				return err
			}

			if err := deps.Identities.Update(n.IdentityAfterUpdate); err != nil {
				s1 := n.IdentityBeforeUpdate.ToSpec()
				s2 := n.IdentityAfterUpdate.ToSpec()
				return identityFillDetails(err, &s2, &s1)
			}

			return nil
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			userRef := model.UserRef{
				Meta: model.Meta{
					ID: n.IdentityAfterUpdate.UserID,
				},
			}

			var e event.Payload
			switch n.IdentityAfterUpdate.Type {
			case model.IdentityTypeLoginID:
				loginIDType := n.IdentityAfterUpdate.LoginID.LoginIDType
				if payload, ok := nonblocking.NewIdentityLoginIDUpdatedEventPayload(
					userRef,
					n.IdentityAfterUpdate.ToModel(),
					n.IdentityBeforeUpdate.ToModel(),
					string(loginIDType),
					n.IsAdminAPI,
				); ok {
					e = payload
				}
			}

			if e != nil {
				err := deps.Events.DispatchEvent(e)
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}, nil
}

func (*NodeDoUpdateIdentity) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	return nil, workflow.ErrEOF
}

func (*NodeDoUpdateIdentity) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	return nil, workflow.ErrIncompatibleInput
}

func (n *NodeDoUpdateIdentity) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}
