package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/api/model"
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	libuser "github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/rolesgroups"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type UserQueries interface {
	Get(userID string, role accesscontrol.Role) (*model.User, error)
}

type ElasticsearchServiceLogger struct{ *log.Logger }

func NewElasticsearchServiceLogger(lf *log.Factory) *ElasticsearchServiceLogger {
	return &ElasticsearchServiceLogger{lf.New("elasticsearch-service")}
}

type UserReindexCreateProducer interface {
	NewTask(appID string, input json.RawMessage) *redisqueue.Task
	EnqueueTask(ctx context.Context, task *redisqueue.Task) error
}

type Service struct {
	Clock       clock.Clock
	Context     context.Context
	Database    *appdb.Handle
	Logger      *ElasticsearchServiceLogger
	AppID       config.AppID
	Client      *elasticsearch.Client
	Users       UserQueries
	UserStore   *user.Store
	OAuth       *identityoauth.Store
	LoginID     *identityloginid.Store
	RolesGroups *rolesgroups.Store
	TaskQueue   task.Queue
	Producer    UserReindexCreateProducer
}

type queryUserResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source model.ElasticsearchUserSource `json:"_source"`
			Sort   interface{}                   `json:"sort"`
		} `json:"hits"`
	} `json:"hits"`
}

func (s *Service) EnqueueReindexUserTask(userID string) error {
	request := ReindexRequest{UserID: userID}

	rawMessage, err := json.Marshal(request)
	if err != nil {
		return err
	}

	task := s.Producer.NewTask(string(s.AppID), rawMessage)
	err = s.Producer.EnqueueTask(s.Context, task)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) getSource(userID string) (*model.ElasticsearchUserSource, error) {
	u, err := s.Users.Get(userID, accesscontrol.RoleGreatest)
	if errors.Is(err, libuser.ErrUserNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	oauthIdentities, err := s.OAuth.List(u.ID)
	if err != nil {
		return nil, err
	}
	loginIDIdentities, err := s.LoginID.List(u.ID)
	if err != nil {
		return nil, err
	}

	effectiveRoles, err := s.RolesGroups.ListEffectiveRolesByUserID(u.ID)
	if err != nil {
		return nil, err
	}

	groups, err := s.RolesGroups.ListGroupsByUserID(u.ID)
	if err != nil {
		return nil, err
	}

	raw := &model.ElasticsearchUserRaw{
		ID:                 u.ID,
		AppID:              string(s.AppID),
		CreatedAt:          u.CreatedAt,
		UpdatedAt:          u.UpdatedAt,
		LastLoginAt:        u.LastLoginAt,
		IsDisabled:         u.IsDisabled,
		StandardAttributes: u.StandardAttributes,
		EffectiveRoles:     slice.Map(effectiveRoles, func(r *rolesgroups.Role) *model.Role { return r.ToModel() }),
		Groups:             slice.Map(groups, func(g *rolesgroups.Group) *model.Group { return g.ToModel() }),
	}

	var arrClaims []map[model.ClaimName]string
	for _, oauthI := range oauthIdentities {
		arrClaims = append(arrClaims, oauthI.ToInfo().IdentityAwareStandardClaims())
		raw.OAuthSubjectID = append(raw.OAuthSubjectID, oauthI.ProviderSubjectID)
	}
	for _, loginIDI := range loginIDIdentities {
		arrClaims = append(arrClaims, loginIDI.ToInfo().IdentityAwareStandardClaims())
	}

	for _, claims := range arrClaims {
		if email, ok := claims[model.ClaimEmail]; ok {
			raw.Email = append(raw.Email, email)
		}
		if phoneNumber, ok := claims[model.ClaimPhoneNumber]; ok {
			raw.PhoneNumber = append(raw.PhoneNumber, phoneNumber)
		}
		if preferredUsername, ok := claims[model.ClaimPreferredUsername]; ok {
			raw.PreferredUsername = append(raw.PreferredUsername, preferredUsername)
		}
	}

	return RawToSource(raw), nil
}

func (s *Service) ExecReindexUser(request ReindexRequest) (result ReindexResult) {
	startedAt := s.Clock.NowUTC()
	var source *model.ElasticsearchUserSource = nil
	err := s.Database.ReadOnly(func() error {
		s, err := s.getSource(request.UserID)
		if err != nil {
			return err
		}
		source = s
		return nil
	})

	failure := func(err error) ReindexResult {
		s.Logger.WithFields(map[string]interface{}{"user_id": request.UserID}).
			WithError(err).
			Error("unknown error on reindexing user")
		return ReindexResult{
			UserID:       request.UserID,
			IsSuccess:    false,
			ErrorMessage: fmt.Sprintf("%v", err),
		}
	}

	if err != nil {
		return failure(err)
	}

	if source == nil {
		err = s.deleteUser(request.UserID)
	} else {
		err = s.reindexUser(source)
	}

	if err != nil {
		return failure(err)
	}

	err = s.Database.WithTx(func() error {
		return s.UserStore.UpdateLastIndexedAt([]string{request.UserID}, startedAt)
	})

	if err != nil {
		return failure(err)
	}

	return ReindexResult{
		UserID:    request.UserID,
		IsSuccess: true,
	}

}

func (s *Service) QueryUser(
	searchKeyword string,
	filterOptions libuser.FilterOptions,
	sortOption libuser.SortOption,
	pageArgs graphqlutil.PageArgs,
) ([]model.PageItemRef, *Stats, error) {
	if s.Client == nil {
		return nil, &Stats{TotalCount: 0}, nil
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

func (s *Service) reindexUser(user *model.ElasticsearchUserSource) error {

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

func (s *Service) deleteUser(userID string) error {
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

func (s *Service) MarkUsersAsReindexRequired(userIDs []string) error {
	return s.UserStore.MarkAsReindexRequired(userIDs)
}
