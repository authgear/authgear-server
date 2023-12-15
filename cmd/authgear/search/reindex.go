package search

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Item struct {
	Value  interface{}
	Cursor model.PageCursor
}

type Reindexer struct {
	Handle  *appdb.Handle
	AppID   config.AppID
	Users   *user.Store
	OAuth   *identityoauth.Store
	LoginID *identityloginid.Store
}

func (q *Reindexer) QueryPage(after model.PageCursor, first uint64) ([]Item, error) {
	users, offset, err := q.Users.QueryPage(user.SortOption{}, graphqlutil.PageArgs{
		First: &first,
		After: graphqlutil.Cursor(after),
	})
	if err != nil {
		return nil, err
	}

	models := make([]Item, len(users))
	for i, u := range users {
		pageKey := db.PageKey{Offset: offset + uint64(i)}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, err
		}
		oauthIdentities, err := q.OAuth.List(u.ID)
		if err != nil {
			return nil, err
		}
		loginIDIdentities, err := q.LoginID.List(u.ID)
		if err != nil {
			return nil, err
		}
		// rawStandardAttributes is used in the re-index command
		// Since the fields that we use for search won't need processing
		// The re-index command should have greatest permission to access all fields.
		// To access standard attributes publicly, it should go through
		// DeriveStandardAttributes func.
		rawStandardAttributes := u.StandardAttributes
		raw := &model.SearchUserRaw{
			ID:                 u.ID,
			AppID:              string(q.AppID),
			CreatedAt:          u.CreatedAt,
			UpdatedAt:          u.UpdatedAt,
			LastLoginAt:        u.MostRecentLoginAt,
			IsDisabled:         u.IsDisabled,
			StandardAttributes: rawStandardAttributes,
		}

		var arrClaims []map[string]interface{}
		for _, oauthI := range oauthIdentities {
			arrClaims = append(arrClaims, oauthI.Claims)
			raw.OAuthSubjectID = append(raw.OAuthSubjectID, oauthI.ProviderSubjectID)
		}
		for _, loginIDI := range loginIDIdentities {
			arrClaims = append(arrClaims, loginIDI.Claims)
		}

		for _, claims := range arrClaims {
			if email, ok := claims["email"].(string); ok {
				raw.Email = append(raw.Email, email)
			}
			if phoneNumber, ok := claims["phone_number"].(string); ok {
				raw.PhoneNumber = append(raw.PhoneNumber, phoneNumber)
			}
			if preferredUsername, ok := claims["preferred_username"].(string); ok {
				raw.PreferredUsername = append(raw.PreferredUsername, preferredUsername)
			}
		}

		models[i] = Item{Value: raw, Cursor: cursor}
	}

	return models, nil
}
