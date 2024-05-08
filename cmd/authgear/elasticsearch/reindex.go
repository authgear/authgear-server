package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/esutil"

	"github.com/authgear/authgear-server/pkg/api/model"
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type queryUserResponse struct {
	Hits struct {
		Hits []struct {
			UnderscoreID string `json:"_id"`
			Source       struct {
				ID string `json:"id"`
			} `json:"_source"`
			Sort interface{} `json:"sort"`
		} `json:"hits"`
	} `json:"hits"`
}

type Item struct {
	Value  interface{}
	Cursor model.PageCursor
}

type ReindexedTimestamp struct {
	UserID      string
	ReindexedAt time.Time
}

type ReindexedTimestamps struct {
	timestamps []*ReindexedTimestamp
	mutex      sync.Mutex
}

func NewReindexedTimestamps() *ReindexedTimestamps {
	return &ReindexedTimestamps{
		timestamps: []*ReindexedTimestamp{},
		mutex:      sync.Mutex{},
	}
}

func (r *ReindexedTimestamps) Append(userID string, timestamp time.Time) {
	r.mutex.Lock()
	t := &ReindexedTimestamp{
		UserID:      userID,
		ReindexedAt: timestamp,
	}
	r.timestamps = append(r.timestamps, t)
	r.mutex.Unlock()
}

func (r *ReindexedTimestamps) Flush(
	dbHandle *appdb.Handle,
	userStore *user.Store) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, t := range r.timestamps {
		err := dbHandle.WithTx(func() error {
			return userStore.UpdateLastIndexedAt([]string{t.UserID}, t.ReindexedAt)
		})
		if err != nil {
			return err
		}
	}
	r.timestamps = []*ReindexedTimestamp{}
	return nil
}

type Reindexer struct {
	Clock               clock.Clock
	Handle              *appdb.Handle
	AppID               config.AppID
	Users               *user.Store
	OAuth               *identityoauth.Store
	LoginID             *identityloginid.Store
	RolesGroups         *rolesgroups.Store
	ReindexedTimestamps *ReindexedTimestamps
}

func (q *Reindexer) QueryPage(after model.PageCursor, first uint64) ([]Item, error) {
	users, offset, err := q.Users.QueryPage(user.ListOptions{}, graphqlutil.PageArgs{
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

		effectiveRoles, err := q.RolesGroups.ListEffectiveRolesByUserID(u.ID)
		if err != nil {
			return nil, err
		}

		groups, err := q.RolesGroups.ListGroupsByUserID(u.ID)
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
			EffectiveRoles:     slice.Map(effectiveRoles, func(r *rolesgroups.Role) *model.Role { return r.ToModel() }),
			Groups:             slice.Map(groups, func(g *rolesgroups.Group) *model.Group { return g.ToModel() }),
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

func (q *Reindexer) Reindex(es *elasticsearch.Client) (err error) {
	ctx := context.Background()
	bulkIndexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:     es,
		Index:      libes.IndexNameUser,
		FlushBytes: 1000000, // 1MB,
	})
	if err != nil {
		return
	}

	allUserIDs, err := q.reindex(ctx, bulkIndexer)
	if err != nil {
		return
	}

	err = q.cleanupDeletedUsers(ctx, es, bulkIndexer, allUserIDs)
	if err != nil {
		return err
	}

	err = bulkIndexer.Close(ctx)
	if err != nil {
		return err
	}

	// Flush timestamps once after closed bulkindexer to ensure all rows are updated
	err = q.ReindexedTimestamps.Flush(q.Handle, q.Users)
	if err != nil {
		return
	}

	stats := bulkIndexer.Stats()
	log.Printf("App (%v): %v indexed; %v deleted; %v failed\n", q.AppID, stats.NumIndexed, stats.NumDeleted, stats.NumFailed)
	return nil
}

func (q *Reindexer) reindex(ctx context.Context, bulkIndexer esutil.BulkIndexer) (allUserIDs map[string]struct{}, err error) {
	allUserIDs = make(map[string]struct{})

	var first uint64 = 50
	var after model.PageCursor = ""
	var items []Item
	startAt := q.Clock.NowUTC()

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
		for _, item := range items {
			user := item.Value.(*model.ElasticsearchUserRaw)
			source := libes.RawToSource(user)
			allUserIDs[user.ID] = struct{}{}

			var bodyBytes []byte
			bodyBytes, err = json.Marshal(source)
			if err != nil {
				return
			}

			err = bulkIndexer.Add(
				ctx,
				esutil.BulkIndexerItem{
					Action:     "index",
					DocumentID: fmt.Sprintf("%s:%s", source.AppID, source.ID),
					Body:       bytes.NewReader(bodyBytes),
					OnFailure: func(
						ctx context.Context,
						item esutil.BulkIndexerItem,
						res esutil.BulkIndexerResponseItem,
						err error,
					) {
						if err != nil {
							fmt.Fprintf(os.Stderr, "%v\n", err)
						} else {
							fmt.Fprintf(os.Stderr, "%v: %v\n", res.Error.Type, res.Error.Reason)
						}
					},
					OnSuccess: func(
						ctx context.Context,
						item esutil.BulkIndexerItem,
						res esutil.BulkIndexerResponseItem) {
						q.ReindexedTimestamps.Append(source.ID, startAt)
					},
				},
			)
			if err != nil {
				return
			}
		}

		err = q.ReindexedTimestamps.Flush(q.Handle, q.Users)
		if err != nil {
			return
		}
	}

	return allUserIDs, nil
}

func (q *Reindexer) cleanupDeletedUsers(ctx context.Context, es *elasticsearch.Client, bulkIndexer esutil.BulkIndexer, userIDs map[string]struct{}) error {
	// Each user ID is 128-bit UUID == 16 bytes
	// 1M users take 16MiB memory
	size := 1000
	var searchAfter interface{}

	underscoreIDsToDelete := make(map[string]struct{})
	for {
		bodyJSONValue := map[string]interface{}{
			"size": size,
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"filter": []interface{}{
						map[string]interface{}{
							"term": map[string]interface{}{
								"app_id": string(q.AppID),
							},
						},
					},
				},
			},
			"sort": []interface{}{
				map[string]interface{}{
					"created_at": "asc",
				},
			},
		}
		if searchAfter != nil {
			bodyJSONValue["search_after"] = searchAfter
		}

		bodyBytes, err := json.Marshal(bodyJSONValue)
		if err != nil {
			return err
		}

		body := bytes.NewReader(bodyBytes)
		res, err := es.Search(func(o *esapi.SearchRequest) {
			o.Index = []string{libes.IndexNameUser}
			o.Body = body
		})
		if err != nil {
			return err
		}
		defer res.Body.Close()

		var response queryUserResponse
		err = json.NewDecoder(res.Body).Decode(&response)
		if err != nil {
			return err
		}

		// Reached the end.
		if len(response.Hits.Hits) == 0 {
			break
		}

		for _, hit := range response.Hits.Hits {
			userID := hit.Source.ID
			_, ok := userIDs[userID]
			if !ok {
				underscoreIDsToDelete[hit.UnderscoreID] = struct{}{}
			}
			searchAfter = hit.Sort
		}
	}

	for underscoreID := range underscoreIDsToDelete {
		err := bulkIndexer.Add(
			ctx,
			esutil.BulkIndexerItem{
				Action:     "delete",
				DocumentID: underscoreID,
				OnFailure: func(
					ctx context.Context,
					item esutil.BulkIndexerItem,
					res esutil.BulkIndexerResponseItem,
					err error,
				) {
					if err != nil {
						fmt.Fprintf(os.Stderr, "%v\n", err)
					} else {
						fmt.Fprintf(os.Stderr, "%v: %v\n", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}
