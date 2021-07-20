package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"

	"github.com/authgear/authgear-server/pkg/api/model"
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Service struct {
	AppID     config.AppID
	Client    *elasticsearch.Client
	Users     *libuser.Store
	OAuth     *identityoauth.Store
	LoginID   *identityloginid.Store
	TaskQueue task.Queue
}

type queryUserResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source model.ElasticsearchUserSource `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (s *Service) ReindexUser(userID string, isDelete bool) (err error) {
	if isDelete {
		s.TaskQueue.Enqueue(&tasks.ReindexUserParam{
			DeleteUserAppID: string(s.AppID),
			DeleteUserID:    userID,
		})
		return nil
	}

	u, err := s.Users.Get(userID)
	if err != nil {
		return
	}
	oauthIdentities, err := s.OAuth.List(u.ID)
	if err != nil {
		return
	}
	loginIDIdentities, err := s.LoginID.List(u.ID)
	if err != nil {
		return
	}

	raw := &model.ElasticsearchUserRaw{
		ID:          u.ID,
		AppID:       string(s.AppID),
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		LastLoginAt: u.MostRecentLoginAt,
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

	s.TaskQueue.Enqueue(&tasks.ReindexUserParam{
		User: RawToSource(raw),
	})
	return nil
}

func (s *Service) QueryUser(
	searchKeyword string,
	sortOption libuser.SortOption,
	pageArgs graphqlutil.PageArgs,
) ([]model.PageItemRef, *Stats, error) {
	if s.Client == nil {
		return nil, &Stats{TotalCount: 0}, nil
	}

	// Prepare body
	bodyJSONValue := MakeSearchBody(s.AppID, searchKeyword, sortOption)
	bodyJSONBytes, err := json.Marshal(bodyJSONValue)
	if err != nil {
		return nil, nil, err
	}
	body := bytes.NewReader(bodyJSONBytes)

	// Prepare size
	size := int(*pageArgs.First)
	if size == 0 {
		size = 20
	}

	// Prepare from
	pageKey, err := db.NewFromPageCursor(model.PageCursor(pageArgs.After))
	if err != nil {
		return nil, nil, err
	}
	from := 0
	if pageKey != nil {
		from = int(pageKey.Offset) + 1
	}

	res, err := s.Client.Search(func(o *esapi.SearchRequest) {
		o.Index = []string{IndexNameUser}
		o.Body = body
		o.Size = &size
		o.From = &from
	})
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, fmt.Errorf("failed to query user: %v", string(bytes))
	}

	var r queryUserResponse
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		return nil, nil, err
	}

	items := make([]model.PageItemRef, len(r.Hits.Hits))
	for i, u := range r.Hits.Hits {
		user := u.Source
		pageKey := db.PageKey{Offset: uint64(from) + uint64(i)}
		cursor, err := pageKey.ToPageCursor()
		if err != nil {
			return nil, nil, err
		}

		items[i] = model.PageItemRef{ID: user.ID, Cursor: cursor}
	}

	return items, &Stats{
		TotalCount: r.Hits.Total.Value,
	}, nil
}
