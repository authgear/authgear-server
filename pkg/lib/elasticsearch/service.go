package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"

	"github.com/authgear/authgear-server/pkg/api/model"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/search/reindex"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Service struct {
	AppID     config.AppID
	Client    *elasticsearch.Client
	Reindexer *reindex.Reindexer
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

func (s *Service) ReindexUser(userID string, isDelete bool) (err error) {
	return s.Reindexer.ReindexUser(config.SearchImplementationElasticsearch, userID, isDelete)
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
