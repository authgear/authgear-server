package loader

import (
	"context"

	"github.com/authgear/authgear-server/pkg/portal/model"
	"github.com/authgear/authgear-server/pkg/portal/session"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

type ViewerLoader struct {
	Context context.Context
}

func (l *ViewerLoader) Get() *graphqlutil.Lazy {
	sessionInfo := session.GetValidSessionInfo(l.Context)
	if sessionInfo == nil {
		return graphqlutil.NewLazyValue(nil)
	}
	return graphqlutil.NewLazyValue(&model.User{
		ID: sessionInfo.UserID,
	})
}
