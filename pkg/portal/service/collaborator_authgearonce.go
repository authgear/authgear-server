//go:build authgearonce
// +build authgearonce

package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

func (s *CollaboratorService) createAccountForInvitee(ctx context.Context, actorUserID string, inviteeEmail string) (err error) {
	params := graphqlutil.DoParams{
		OperationName: "createAccount",
		Query: `
		mutation createAccount($email: String!) {
			createUser(input: {
				definition: {
					loginID: {
						key: "email"
						value: $email
					}
				}
				sendPassword: true
				setPasswordExpired: true
			}) {
				user {
					id
				}
			}
		}
		`,
		Variables: map[string]interface{}{
			"email": inviteeEmail,
		},
	}

	r, err := http.NewRequestWithContext(ctx, "POST", "/graphql", nil)
	if err != nil {
		return err
	}

	director, err := s.AdminAPI.SelfDirector(ctx, actorUserID, UsageInternal)
	if err != nil {
		return err
	}

	director(r)

	result, err := graphqlutil.HTTPDo(s.HTTPClient.Client, r, params)
	if err != nil {
		return err
	}

	if result.HasErrors() {
		return fmt.Errorf("unexpected graphql errors: %v", result.Errors)
	}

	return
}
