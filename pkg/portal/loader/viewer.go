package loader

import (
	"context"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/session"
)

type ViewerLoader struct {
	Context context.Context
}

func (l *ViewerLoader) Get() (interface{}, error) {
	sessionInfo := session.GetValidSessionInfo(l.Context)
	if sessionInfo == nil {
		return nil, nil
	}

	return &model.User{
		ID: sessionInfo.UserID,
	}, nil
}
