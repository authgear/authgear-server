package cmddatabase

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Restorer struct {
	Context        context.Context
	DatabaseURL    string
	DatabaseSchema string
	InputDir       string
	AppIDs         []string

	dbHandle    *db.HookHandle
	sqlExecutor *db.SQLExecutor
	sqlBuilder  *db.SQLBuilder
	logger      *log.Logger
}

func NewRestorer(
	context context.Context,
	databaseURL string,
	databaseSchema string,
	inputDir string,
	appIDs []string,
) *Restorer {
	loggerFactory := log.NewFactory(
		log.LevelDebug,
	)
	logger := loggerFactory.New("restorer")
	pool := db.NewPool()
	handle := db.NewHookHandle(
		context,
		pool,
		db.ConnectionOptions{
			DatabaseURL:           databaseURL,
			MaxOpenConnection:     1,
			MaxIdleConnection:     1,
			MaxConnectionLifetime: 1800 * time.Second,
			IdleConnectionTimeout: 300 * time.Second,
		},
		loggerFactory,
	)
	sqlExecutor := &db.SQLExecutor{
		Context:  context,
		Database: handle,
	}
	sqlBuilder := db.NewSQLBuilder(databaseSchema)
	return &Restorer{
		Context:        context,
		DatabaseURL:    databaseURL,
		DatabaseSchema: databaseSchema,
		InputDir:       inputDir,
		AppIDs:         appIDs,

		dbHandle:    handle,
		sqlExecutor: sqlExecutor,
		sqlBuilder:  &sqlBuilder,
		logger:      logger,
	}
}

func (r *Restorer) Restore() error {
	inputPathAbs, err := filepath.Abs(r.InputDir)
	if err != nil {
		panic(err)
	}
	r.logger.Info(fmt.Sprintf("Restoring from %s", inputPathAbs))

	return r.dbHandle.WithTx(func() error {
		for _, tableName := range tableNames {
			inputFile := filepath.Join(inputPathAbs, fmt.Sprintf("%s.csv", tableName))
			f, err := os.Open(inputFile)
			if err != nil {
				r.logger.Warn(fmt.Sprintf("Restoration of %s skipped: failed to open %s", tableName, inputFile))
				continue
			}
			defer f.Close()

			r.logger.Info(fmt.Sprintf("Restoring %s", tableName))
			csvReader := csv.NewReader(f)
			records, err := csvReader.ReadAll()
			if err != nil {
				panic(err)
			}
			columns, data, err := r.convertToDatabaseData(records)
			if err != nil {
				r.logger.WithError(err).Error(fmt.Sprintf("Error on restoring %s from %s", tableName, inputFile))
				return err
			}
			for _, row := range data {
				q := r.sqlBuilder.
					Insert(r.sqlBuilder.TableName(tableName)).
					Columns(columns...).
					Values(row...)
				_, err := r.sqlExecutor.ExecWith(q)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})

}

func (r *Restorer) convertToDatabaseData(csvRecords [][]string) (columns []string, rows [][]interface{}, err error) {
	if len(csvRecords) == 0 {
		err = fmt.Errorf("csv missing header")
		return
	}
	columns = csvRecords[0]
	remainings := csvRecords[1:]
	for _, csvRow := range remainings {
		if len(csvRow) != len(columns) {
			err = fmt.Errorf("invalid format")
			return
		}
		row := []interface{}{}
		for _, value := range csvRow {
			v := value
			if value == NULL {
				var nullptr *string = nil
				row = append(row, nullptr)
			} else {
				vptr := &v
				row = append(row, vptr)
			}
		}
		rows = append(rows, row)
	}
	return
}
