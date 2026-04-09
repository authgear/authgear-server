package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/graphql-go/graphql"

	relay "github.com/authgear/authgear-server/pkg/graphqlgo/relay"
	portalservice "github.com/authgear/authgear-server/pkg/portal/service"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type SiteAdminAdminAPI interface {
	SelfDirector(ctx context.Context, actorUserID string, usage portalservice.Usage) (func(*http.Request), error)
}

type SiteAdminHTTPClient struct {
	*http.Client
}

type AdminAPIService struct {
	AdminAPI   SiteAdminAdminAPI
	HTTPClient SiteAdminHTTPClient
}

func (s *AdminAPIService) FindUserIDsByEmail(ctx context.Context, email string) ([]string, error) {
	params := graphqlutil.DoParams{
		OperationName: "getUsersByStandardAttribute",
		Query: `
		query getUsersByStandardAttribute($name: String!, $value: String!) {
			users: getUsersByStandardAttribute(attributeName: $name, attributeValue: $value) {
				id
			}
		}
		`,
		Variables: map[string]interface{}{
			"name":  "email",
			"value": email,
		},
	}

	result, err := s.do(ctx, params)
	if err != nil {
		return nil, err
	}
	if result.HasErrors() {
		return nil, fmt.Errorf("failed to search users by email: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	users := data["users"].([]interface{})

	ids := make([]string, 0, len(users))
	for _, u := range users {
		userNode, ok := u.(map[string]interface{})
		if !ok {
			continue
		}
		globalID, _ := userNode["id"].(string)
		resolved := relay.FromGlobalID(globalID)
		if resolved == nil || resolved.ID == "" {
			continue
		}
		ids = append(ids, resolved.ID)
	}
	return ids, nil
}

func (s *AdminAPIService) ResolveUserEmails(ctx context.Context, userIDs []string) (map[string]string, error) {
	if len(userIDs) == 0 {
		return map[string]string{}, nil
	}

	globalIDs := make([]string, len(userIDs))
	for i, id := range userIDs {
		globalIDs[i] = relay.ToGlobalID("User", id)
	}

	params := graphqlutil.DoParams{
		OperationName: "getUserNodes",
		Query: `
		query getUserNodes($ids: [ID!]!) {
			nodes(ids: $ids) {
				... on User {
					id
					standardAttributes
				}
			}
		}
		`,
		Variables: map[string]interface{}{
			"ids": globalIDs,
		},
	}

	result, err := s.do(ctx, params)
	if err != nil {
		return nil, err
	}
	if result.HasErrors() {
		return nil, fmt.Errorf("failed to resolve user emails: %v", result.Errors)
	}

	emailMap := make(map[string]string, len(userIDs))
	data := result.Data.(map[string]interface{})
	nodes := data["nodes"].([]interface{})
	for _, node := range nodes {
		userNode, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		globalID, _ := userNode["id"].(string)
		resolvedID := relay.FromGlobalID(globalID)
		if resolvedID == nil || resolvedID.ID == "" {
			continue
		}
		attrs, ok := userNode["standardAttributes"].(map[string]interface{})
		if !ok {
			continue
		}
		email, _ := attrs["email"].(string)
		emailMap[resolvedID.ID] = email
	}
	return emailMap, nil
}

func (s *AdminAPIService) do(ctx context.Context, params graphqlutil.DoParams) (*graphql.Result, error) {
	r, err := http.NewRequestWithContext(ctx, "POST", "/graphql", nil)
	if err != nil {
		return nil, err
	}

	actorUserID := session.GetValidSessionInfo(ctx).UserID
	director, err := s.AdminAPI.SelfDirector(ctx, actorUserID, portalservice.UsageInternal)
	if err != nil {
		return nil, err
	}
	director(r)

	return graphqlutil.HTTPDo(s.HTTPClient.Client, r, params)
}
