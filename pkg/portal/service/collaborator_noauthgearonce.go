//go:build !authgearonce
// +build !authgearonce

package service

import (
	"context"
)

func (s *CollaboratorService) createAccountForInvitee(ctx context.Context, actorUserID string, inviteeEmail string) (err error) {
	// It is noop
	return
}
