package facade

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/userexport"
)

type UserExportService interface {
	ExportRecords(ctx context.Context, request *userexport.Request) *userexport.Response
}

type UserExportFacade struct {
	Service UserExportService
}

func (s *UserExportFacade) ExportRecords(ctx context.Context, request *userexport.Request) *userexport.Response {
	return s.Service.ExportRecords(ctx, request)
}
