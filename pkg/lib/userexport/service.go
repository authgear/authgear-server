package userexport

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type UserQueries interface {
	GetPageForExport(page uint64) (users []*UserForExport, err error)
	CountAll() (count uint64, err error)
}

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
	var total_user uint64 = 0
	err := s.AppDatabase.WithTx(func() (e error) {
		count, err := s.UserQueries.CountAll()
		if err != nil {
			return
		}

		total_user = count
		return
	})

	s.Logger.Infof("Export total users: %v", total_user)

	if total_user > 0 {
		// TODO: write to a tmp file
		writer := io.MultiWriter(io.Discard, os.Stdout)

		for offset := uint64(0); offset < total_user; offset += BatchSize {
			var page []*UserForExport = nil
			err = s.AppDatabase.WithTx(func() (e error) {
				result, pageErr := s.UserQueries.GetPageForExport(offset)
				if pageErr != nil {
					return
				}
				page = result
				return
			})

			for _, user := range page {
				// TODO: Convert user model to Record
				recordJson, json_err := json.Marshal(user)
				if json_err != nil {
					return nil
				}
				recordBytes := make([]byte, 0)
				recordBytes = append(recordBytes, []byte(recordJson)...)
				recordBytes = append(recordBytes, []byte("\n")...)
				writer.Write(recordBytes)
			}
		}

		// TODO: Upload tmp result output to cloud storage
	}

	if err != nil {
		return nil
	}

	// TODO: return worker task response
	now := time.Now()
	return &Response{
		ID:        "dummy_task_id",
		CreatedAt: &now,
		Status:    "pending",
		Request:   request,
	}
}
