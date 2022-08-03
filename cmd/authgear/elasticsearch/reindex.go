package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
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
		raw := &model.ElasticsearchUserRaw{
			ID:                 u.ID,
			AppID:              string(q.AppID),
			CreatedAt:          u.CreatedAt,
			UpdatedAt:          u.UpdatedAt,
			LastLoginAt:        u.MostRecentLoginAt,
			IsDisabled:         u.IsDisabled,
			StandardAttributes: rawStandardAttributes,
		}

		var arrClaims []map[identity.ClaimKey]interface{}
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

func (q *Reindexer) Reindex(es *elasticsearch.Client) (err error) {
	var first uint64 = 50
	var after model.PageCursor = ""
	var items []Item

	for {
		err = q.Handle.WithTx(func() (err error) {
			items, err = q.QueryPage(after, first)
			if err != nil {
				return
			}
			return nil
		})
		if err != nil {
			return
		}

		// Termination condition
		if len(items) <= 0 {
			break
		}

		// Prepare for next iteration
		after = items[len(items)-1].Cursor

		// Process the items
		buf := &bytes.Buffer{}
		for _, item := range items {
			user := item.Value.(*model.ElasticsearchUserRaw)
			fmt.Printf("Indexing app (%s) user (%s)\n", user.AppID, user.ID)
			err = q.writeBody(buf, user)
			if err != nil {
				return
			}
		}

		var res *esapi.Response
		res, err = es.Bulk(buf, func(o *esapi.BulkRequest) {
			o.Index = libes.IndexNameUser
		})
		if err != nil {
			return
		}
		defer res.Body.Close()
		if res.IsError() {
			err = fmt.Errorf("%v", res)
			return
		}
	}

	return nil
}

func (q *Reindexer) writeBody(buf io.Writer, raw *model.ElasticsearchUserRaw) (err error) {
	source := libes.RawToSource(raw)
	id := fmt.Sprintf("%s:%s", source.AppID, source.ID)
	action := map[string]interface{}{
		"index": map[string]interface{}{
			"_id": id,
		},
	}
	actionBytes, err := json.Marshal(action)
	if err != nil {
		return
	}

	_, err = buf.Write(actionBytes)
	if err != nil {
		return
	}

	_, err = buf.Write([]byte("\n"))
	if err != nil {
		return
	}

	sourceBytes, err := json.Marshal(source)
	if err != nil {
		return
	}

	_, err = buf.Write(sourceBytes)
	if err != nil {
		return
	}

	_, err = buf.Write([]byte("\n"))
	if err != nil {
		return
	}

	return nil
}
