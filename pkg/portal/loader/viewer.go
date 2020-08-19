package loader

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/upstreamapp"
	"github.com/authgear/authgear-server/pkg/portal/model"
)

type ViewerLoader struct {
	Context context.Context
}

func (l *ViewerLoader) Get() (interface{}, error) {
	sessionInfo := upstreamapp.GetValidSessionInfo(l.Context)
	if sessionInfo == nil {
		return nil, nil
	}

	return &model.User{
		ID: sessionInfo.UserID,
	}, nil
}
