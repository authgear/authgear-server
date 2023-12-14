package tasks

import (
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

const ReindexUser = "ReindexUser"

type ReindexUserParam struct {
	Implementation  config.SearchImplementation
	DeleteUserAppID string
	DeleteUserID    string
	User            *apimodel.SearchUserSource
}

func (p *ReindexUserParam) TaskName() string {
	return ReindexUser
}
