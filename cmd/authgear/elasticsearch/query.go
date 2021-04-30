package elasticsearch

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
)

type Query struct {
	AppID   config.AppID
	Users   *user.Store
	OAuth   *identityoauth.Store
	LoginID *identityloginid.Store
}

func (q *Query) QueryPage(after model.PageCursor, first uint64) ([]model.PageItem, error) {
	users, offset, err := q.Users.QueryPage(after, model.PageCursor(""), &first, nil)
	if err != nil {
		return nil, err
	}

	models := make([]model.PageItem, len(users))
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
		val := &User{
			ID:          u.ID,
			AppID:       string(q.AppID),
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
			LastLoginAt: u.LastLoginAt,
			IsDisabled:  u.IsDisabled,
		}

		var arrClaims []map[string]interface{}
		for _, oauthI := range oauthIdentities {
			arrClaims = append(arrClaims, oauthI.Claims)
		}
		for _, loginIDI := range loginIDIdentities {
			arrClaims = append(arrClaims, loginIDI.Claims)
		}

		for _, claims := range arrClaims {
			email, ok := claims["email"].(string)
			if ok {
				val.Email = append(val.Email, email)
			}
			phoneNumber, ok := claims["phone_number"].(string)
			if ok {
				val.PhoneNumber = append(val.PhoneNumber, phoneNumber)
			}
			preferredUsername, ok := claims["preferred_username"].(string)
			if ok {
				val.PreferredUsername = append(val.PreferredUsername, preferredUsername)
			}
		}

		models[i] = model.PageItem{Value: val, Cursor: cursor}
	}

	return models, nil
}
