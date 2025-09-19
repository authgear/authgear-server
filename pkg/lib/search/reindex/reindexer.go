package reindex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/lib/infra/redisqueue"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

type ReindexRequest struct {
	UserID string `json:"user_id"`
}

type ReindexResult struct {
	UserID       string `json:"user_id"`
	IsSuccess    bool   `json:"is_success"`
	ErrorMessage string `json:"error_message,omitempty"`
}

var ReindexerLogger = slogutil.NewLogger("search-reindexer")

type ElasticsearchReindexer interface {
	ReindexUser(ctx context.Context, user *model.SearchUserSource) error
	DeleteUser(ctx context.Context, userID string) error
}

type PostgresqlReindexer interface {
	ReindexUser(ctx context.Context, user *model.SearchUserSource) error
	DeleteUser(ctx context.Context, userID string) error
}

type UserReindexCreateProducer interface {
	NewTask(appID string, input json.RawMessage, taskIDPrefix string) *redisqueue.Task
	EnqueueTask(ctx context.Context, task *redisqueue.Task) error
}

type Reindexer struct {
	AppID        config.AppID
	SearchConfig *config.SearchConfig
	Clock        clock.Clock
	Database     *appdb.Handle
	UserStore    *user.Store
	Producer     UserReindexCreateProducer

	SourceProvider *SourceProvider

	ElasticsearchReindexer     ElasticsearchReindexer
	PostgresqlReindexer        PostgresqlReindexer
	GlobalSearchImplementation config.GlobalSearchImplementation
}

type action string

const (
	actionReindex action = "reindex"
	actionDelete  action = "delete"
	actionSkip    action = "skip"
)

func (s *Reindexer) getSourceWithAction(ctx context.Context, userID string) (*model.SearchUserSource, action, error) {
	rawUser, err := s.UserStore.Get(ctx, userID)
	if errors.Is(err, user.ErrUserNotFound) {
		return nil, actionDelete, nil
	}
	if rawUser.LastIndexedAt != nil && rawUser.RequireReindexAfter != nil && rawUser.LastIndexedAt.After(*rawUser.RequireReindexAfter) {
		// Already latest state, skip the update
		return nil, actionSkip, nil
	}

	source, err := s.SourceProvider.getSource(ctx, rawUser)
	if err != nil {
		return nil, "", err
	}

	return source, actionReindex, nil
}

func (s *Reindexer) ExecReindexUser(ctx context.Context, request ReindexRequest) (result ReindexResult) {
	logger := ReindexerLogger.GetLogger(ctx)
	failure := func(err error) ReindexResult {
		logger.WithError(err).Error(ctx, "unknown error on reindexing user",
			slog.String("user_id", request.UserID),
		)
		return ReindexResult{
			UserID:       request.UserID,
			IsSuccess:    false,
			ErrorMessage: fmt.Sprintf("%v", err),
		}
	}

	startedAt := s.Clock.NowUTC()
	var source *model.SearchUserSource = nil
	var actionToExec action
	err := s.Database.ReadOnly(ctx, func(ctx context.Context) error {
		s, a, err := s.getSourceWithAction(ctx, request.UserID)
		if err != nil {
			return err
		}
		source = s
		actionToExec = a
		return nil
	})

	if err != nil {
		return failure(err)
	}

	switch actionToExec {
	case actionDelete:
		err = s.deleteUser(ctx, request.UserID)
		if err != nil {
			return failure(err)
		}

	case actionReindex:
		err = s.reindexUser(ctx, source)
		if err != nil {
			return failure(err)
		}
		err = s.Database.WithTx(ctx, func(ctx context.Context) error {
			return s.UserStore.UpdateLastIndexedAt(ctx, []string{request.UserID}, startedAt)
		})
		if err != nil {
			return failure(err)
		}

	case actionSkip:
		logger.Info(ctx, "skipping reindexing user because it is already up to date",
			slog.String("app_id", string(s.AppID)),
			slog.String("user_id", request.UserID),
		)
	default:
		panic(fmt.Errorf("search: unknown action %s", actionToExec))
	}

	return ReindexResult{
		UserID:    request.UserID,
		IsSuccess: true,
	}

}

func (s *Reindexer) MarkUsersAsReindexRequiredInTx(ctx context.Context, userIDs []string) error {
	return s.UserStore.MarkAsReindexRequired(ctx, userIDs)
}

func (s *Reindexer) EnqueueReindexUserTask(ctx context.Context, userID string) error {
	request := ReindexRequest{UserID: userID}
	rawMessage, err := json.Marshal(request)
	if err != nil {
		return err
	}
	task := s.Producer.NewTask(string(s.AppID), rawMessage, "task")
	err = s.Producer.EnqueueTask(ctx, task)
	if err != nil {
		return err
	}

	return nil
}

func (s *Reindexer) reindexUser(ctx context.Context, source *model.SearchUserSource) error {
	switch s.SearchConfig.GetImplementation(s.GlobalSearchImplementation) {
	case config.SearchImplementationElasticsearch:
		return s.ElasticsearchReindexer.ReindexUser(ctx, source)
	case config.SearchImplementationPostgresql:
		return s.PostgresqlReindexer.ReindexUser(ctx, source)
	case config.SearchImplementationNone:
		// Do nothing
		return nil
	}

	panic(fmt.Errorf("unknown search implementation %s", s.SearchConfig.GetImplementation(s.GlobalSearchImplementation)))
}

func (s *Reindexer) deleteUser(ctx context.Context, userID string) error {
	switch s.SearchConfig.GetImplementation(s.GlobalSearchImplementation) {
	case config.SearchImplementationElasticsearch:
		return s.ElasticsearchReindexer.DeleteUser(ctx, userID)
	case config.SearchImplementationPostgresql:
		return s.PostgresqlReindexer.DeleteUser(ctx, userID)
	case config.SearchImplementationNone:
		// Do nothing
		return nil
	}

	panic(fmt.Errorf("unknown search implementation %s", s.SearchConfig.GetImplementation(s.GlobalSearchImplementation)))
}
