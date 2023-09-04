package latte

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
)

func reauthUserIDHint(ctx context.Context) (string, error) {
	userID := workflow.GetUserIDHint(ctx)
	if userID == "" {
		return "", apierrors.NewInvalid("this workflow must be triggered in a reauthentication session")
	}
	return userID, nil
}
