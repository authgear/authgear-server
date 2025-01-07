package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/api/model"
	identityservice "github.com/authgear/authgear-server/pkg/lib/authn/identity/service"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type UserQueries interface {
	Get(ctx context.Context, userID string, role accesscontrol.Role) (*model.User, error)
}

type ElasticsearchServiceLogger struct{ *log.Logger }

func NewElasticsearchServiceLogger(lf *log.Factory) *ElasticsearchServiceLogger {
	return &ElasticsearchServiceLogger{lf.New("elasticsearch-service")}
}

type Service struct {
	Clock           clock.Clock
	Database        *appdb.Handle
	Logger          *ElasticsearchServiceLogger
	AppID           config.AppID
	Client          *elasticsearch.Client
	Users           UserQueries
	UserStore       *user.Store
	IdentityService *identityservice.Service
	RolesGroups     *rolesgroups.Store
}

type queryUserResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source model.SearchUserSource `json:"_source"`
			Sort   interface{}            `json:"sort"`
		} `json:"hits"`
	} `json:"hits"`
}

func (s *Service) QueryUser(
	searchKeyword string,
	filterOptions libuser.FilterOptions,
	sortOption libuser.SortOption,
	pageArgs graphqlutil.PageArgs,
) ([]model.PageItemRef, *Stats, error) {
	if s.Client == nil {
		return nil, nil, ErrMissingCredential
	}

	// Prepare body
	bodyJSONValue := MakeSearchBody(s.AppID, searchKeyword, filterOptions, sortOption)

	// Prepare search_after
	searchAfter, err := CursorToSearchAfter(model.PageCursor(pageArgs.After))
	if err != nil {
		return nil, nil, err
	}
	if searchAfter != nil {
		bodyJSONValue["search_after"] = searchAfter
	}

	bodyJSONBytes, err := json.Marshal(bodyJSONValue)
	if err != nil {
		return nil, nil, err
	}
	body := bytes.NewReader(bodyJSONBytes)

	// Prepare size
	//nolint:gosec // G115
	size := int(*pageArgs.First)
	if size == 0 {
		size = 20
	}

	res, err := s.Client.Search(func(o *esapi.SearchRequest) {
		o.Index = []string{IndexNameUser}
		o.Body = body
		o.Size = &size
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
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
		cursor, err := SortToCursor(u.Sort)
		if err != nil {
			return nil, nil, err
		}
		items[i] = model.PageItemRef{ID: user.ID, Cursor: cursor}
	}

	return items, &Stats{
		TotalCount: r.Hits.Total.Value,
	}, nil
}

func (s *Service) ReindexUser(user *model.SearchUserSource) error {

	documentID := fmt.Sprintf("%s:%s", user.AppID, user.ID)
	s.Logger.WithFields(logrus.Fields{
		"app_id":  user.AppID,
		"user_id": user.ID,
	}).Info("reindexing user")

	var res *esapi.Response

	var sourceBytes []byte
	sourceBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}

	body := &bytes.Buffer{}
	_, err = body.Write(sourceBytes)
	if err != nil {
		return err
	}

	res, err = s.Client.Index(IndexNameUser, body, func(o *esapi.IndexRequest) {
		o.DocumentID = documentID
	})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		err = fmt.Errorf("%v", res)
		return err
	}
	return nil
}

func (s *Service) DeleteUser(userID string) error {
	appID := s.AppID
	s.Logger.WithFields(logrus.Fields{
		"app_id":  appID,
		"user_id": userID,
	}).Info("removing user from index")

	documentID := fmt.Sprintf("%s:%s", appID, userID)

	var res *esapi.Response

	res, err := s.Client.Delete(IndexNameUser, documentID)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		err = fmt.Errorf("%v", res)
		return err
	}
	return nil
}
