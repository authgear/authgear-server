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
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type Service struct {
	AppID  config.AppID
	Client *elasticsearch.Client
}

type queryUserResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source User `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
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
