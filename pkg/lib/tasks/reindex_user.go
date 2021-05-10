package tasks

import (
	apimodel "github.com/authgear/authgear-server/pkg/api/model"
)

const ReindexUser = "ReindexUser"

type ReindexUserParam struct {
	DeleteUserAppID string
	DeleteUserID    string
	User            *apimodel.ElasticsearchUser
}

func (p *ReindexUserParam) TaskName() string {
	return ReindexUser
}
