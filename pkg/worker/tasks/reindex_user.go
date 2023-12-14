package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/sirupsen/logrus"

	"github.com/authgear/authgear-server/pkg/lib/config"
	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
	"github.com/authgear/authgear-server/pkg/lib/search/pgsearch"
	"github.com/authgear/authgear-server/pkg/lib/tasks"
	"github.com/authgear/authgear-server/pkg/util/log"
)

func ConfigureReindexUserTask(registry task.Registry, t task.Task) {
	registry.Register(tasks.ReindexUser, t)
}

type ReindexUserLogger struct{ *log.Logger }

func NewReindexUserLogger(lf *log.Factory) ReindexUserLogger {
	return ReindexUserLogger{lf.New("reindex-user")}
}

type ReindexUserTask struct {
	ElasticsearchClient *elasticsearch.Client
	PGStore             *pgsearch.Store
	Logger              ReindexUserLogger
}

func (t *ReindexUserTask) Run(ctx context.Context, param task.Param) (err error) {
	taskParam := param.(*tasks.ReindexUserParam)

	switch taskParam.Implementation {
	case config.SearchImplementationElasticsearch:
		return t.runElasticsearch(taskParam)
	case config.SearchImplementationPostgresql:
		return t.runPostgresql(taskParam)
	}

	return nil
}

func (t *ReindexUserTask) runPostgresql(taskParam *tasks.ReindexUserParam) (err error) {
	if t.PGStore == nil {
		t.Logger.Warn("search database credential not provided, skip reindexing user")
		return nil
	}
	err = t.PGStore.Database.WithTx(func() error {
		if taskParam.DeleteUserID != "" {
			appID := taskParam.DeleteUserAppID
			userID := taskParam.DeleteUserID

			t.Logger.WithFields(logrus.Fields{
				"app_id":  appID,
				"user_id": userID,
			}).Info("removing user from search database")

			err := t.PGStore.DeleteUser(appID, userID)
			if err != nil {
				return err
			}
			return nil
		} else {
			user := taskParam.User
			t.Logger.WithFields(logrus.Fields{
				"app_id":  user.AppID,
				"user_id": user.ID,
			}).Info("reindexing user in search database")
			err := t.PGStore.UpsertUser(user)
			if err != nil {
				return err
			}
			return nil
		}
	})
	return
}

func (t *ReindexUserTask) runElasticsearch(taskParam *tasks.ReindexUserParam) (err error) {
	if t.ElasticsearchClient == nil {
		t.Logger.Warn("elasticsearch credential not provided, skip reindexing user")
		return nil
	}
	if taskParam.DeleteUserID != "" {
		appID := taskParam.DeleteUserAppID
		userID := taskParam.DeleteUserID

		t.Logger.WithFields(logrus.Fields{
			"app_id":  appID,
			"user_id": userID,
		}).Info("removing user from index")

		documentID := fmt.Sprintf("%s:%s", appID, userID)

		var res *esapi.Response

		res, err = t.ElasticsearchClient.Delete(libes.IndexNameUser, documentID)
		if err != nil {
			return
		}
		defer res.Body.Close()
		if res.IsError() {
			err = fmt.Errorf("%v", res)
			return
		}
	} else {
		user := taskParam.User
		documentID := fmt.Sprintf("%s:%s", user.AppID, user.ID)
		t.Logger.WithFields(logrus.Fields{
			"app_id":  user.AppID,
			"user_id": user.ID,
		}).Info("reindexing user")

		var res *esapi.Response

		var sourceBytes []byte
		sourceBytes, err = json.Marshal(user)
		if err != nil {
			return
		}

		body := &bytes.Buffer{}
		_, err = body.Write(sourceBytes)
		if err != nil {
			return
		}

		res, err = t.ElasticsearchClient.Index(libes.IndexNameUser, body, func(o *esapi.IndexRequest) {
			o.DocumentID = documentID
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
