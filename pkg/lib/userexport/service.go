package userexport

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type UserQueries interface{}

type UserExportService struct {
	AppDatabase *appdb.Handle
	UserQueries UserQueries
	Logger      Logger
}

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger {
	return Logger{lf.New("user-export")}
}

func (s *UserExportService) ExportRecords(ctx context.Context, request *Request) *Response {
	// TODO: Add query logic
	return nil
}
