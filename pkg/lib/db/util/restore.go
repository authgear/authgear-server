package util

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
	ConnectionInfo db.ConnectionInfo
	DatabaseSchema string
	InputDir       string
	AppIDs         []string
	TableNames     []string

	dbHandle    *db.HookHandle
	sqlExecutor *db.SQLExecutor
	sqlBuilder  *db.SQLBuilder
	logger      *log.Logger
}

func NewRestorer(
	connectionInfo db.ConnectionInfo,
	databaseSchema string,
	inputDir string,
	appIDs []string,
	tableNames []string,
) *Restorer {
	loggerFactory := log.NewFactory(
		log.LevelDebug,
	)
	logger := loggerFactory.New("restorer")
	pool := db.NewPool()
	handle := db.NewHookHandle(
		pool,
		connectionInfo,
		db.ConnectionOptions{
			MaxOpenConnection:     1,
			MaxIdleConnection:     1,
			MaxConnectionLifetime: 1800 * time.Second,
			IdleConnectionTimeout: 300 * time.Second,
		},
		loggerFactory,
	)
	sqlExecutor := &db.SQLExecutor{}
	sqlBuilder := db.NewSQLBuilder(databaseSchema)
	return &Restorer{
		ConnectionInfo: connectionInfo,
		DatabaseSchema: databaseSchema,
		InputDir:       inputDir,
		AppIDs:         appIDs,
		TableNames:     tableNames,

		dbHandle:    handle,
		sqlExecutor: sqlExecutor,
		sqlBuilder:  &sqlBuilder,
		logger:      logger,
	}
}

func (r *Restorer) Restore(ctx context.Context) error {
	inputPathAbs, err := filepath.Abs(r.InputDir)
	if err != nil {
		panic(err)
	}
	r.logger.Info(fmt.Sprintf("Restoring from %s", inputPathAbs))

	return r.dbHandle.WithTx(ctx, func(ctx context.Context) error {
		for _, tableName := range r.TableNames {
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
				_, err := r.sqlExecutor.ExecWith(ctx, q)
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
