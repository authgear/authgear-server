package tasks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/sirupsen/logrus"

	libes "github.com/authgear/authgear-server/pkg/lib/elasticsearch"
	"github.com/authgear/authgear-server/pkg/lib/infra/task"
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
	Client *elasticsearch.Client
	Logger ReindexUserLogger
}

func (t *ReindexUserTask) Run(ctx context.Context, param task.Param) (err error) {
	if t.Client == nil {
		t.Logger.Info("skip reindexing user")
		return
	}

	taskParam := param.(*tasks.ReindexUserParam)

	if taskParam.DeleteUserID != "" {
		appID := taskParam.DeleteUserAppID
		userID := taskParam.DeleteUserID

		t.Logger.WithFields(logrus.Fields{
			"app_id":  appID,
			"user_id": userID,
		}).Info("removing user from index")

		documentID := fmt.Sprintf("%s:%s", appID, userID)

		var res *esapi.Response

		res, err = t.Client.Delete(libes.IndexNameUser, documentID)
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

		res, err = t.Client.Index(libes.IndexNameUser, body, func(o *esapi.IndexRequest) {
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
