package latte

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/uuid"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPublicIntent(&IntentMigrate{})
}

var IntentMigrateSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentMigrate struct{}

func (*IntentMigrate) Kind() string {
	return "latte.IntentMigrate"
}

func (*IntentMigrate) JSONSchema() *validation.SimpleSchema {
	return IntentMigrateSchema
}

func (*IntentMigrate) CanReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) ([]workflow.Input, error) {
	switch len(w.Nodes) {
	case 0:
		// Generate a new user ID.
		return nil, nil
	case 1:
		// Migrate from the migration token.
		return nil, nil
	default:
		return nil, workflow.ErrEOF
	}
}

func (i *IntentMigrate) ReactTo(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow, input workflow.Input) (*workflow.Node, error) {
	// Check the migration token
	switch len(w.Nodes) {
	case 0:
		// The token will be taken in on-commit effect.
		bucket := AntiSpamSignupBucket(string(deps.RemoteIP))
		pass, _, err := deps.RateLimiter.CheckToken(bucket)
		if err != nil {
			return nil, err
		}
		if !pass {
			return nil, bucket.BucketError()
		}
		return workflow.NewNodeSimple(&NodeDoCreateUser{
			UserID: uuid.New(),
		}), nil
	case 1:
		return workflow.NewSubWorkflow(&IntentMigrateAccount{
			UseID: i.userID(w),
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (*IntentMigrate) GetEffects(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			// Apply ratelimit on sign up.
			bucket := AntiSpamSignupBucket(string(deps.RemoteIP))
			err := deps.RateLimiter.TakeToken(bucket)
			if err != nil {
				return err
			}
			return nil
		}),
	}, nil
}

func (*IntentMigrate) OutputData(ctx context.Context, deps *workflow.Dependencies, w *workflow.Workflow) (interface{}, error) {
	return nil, nil
}

func (i *IntentMigrate) userID(w *workflow.Workflow) string {
	node, ok := workflow.FindSingleNode[*NodeDoCreateUser](w)
	if !ok {
		panic(fmt.Errorf("expected userID"))
	}
	return node.UserID
}
