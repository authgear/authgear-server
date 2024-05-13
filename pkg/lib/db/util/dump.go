package util

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type Dumper struct {
	Context        context.Context
	DatabaseURL    string
	DatabaseSchema string
	OutputDir      string
	AppIDs         []string
	TableNames     []string

	dbHandle    *db.HookHandle
	sqlExecutor *db.SQLExecutor
	sqlBuilder  *db.SQLBuilder
	logger      *log.Logger
}

func NewDumper(
	context context.Context,
	databaseURL string,
	databaseSchema string,
	outputDir string,
	appIDs []string,
	tableNames []string,
) *Dumper {
	loggerFactory := log.NewFactory(
		log.LevelDebug,
	)
	logger := loggerFactory.New("dumper")
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
	return &Dumper{
		Context:        context,
		DatabaseURL:    databaseURL,
		DatabaseSchema: databaseSchema,
		OutputDir:      outputDir,
		AppIDs:         appIDs,
		TableNames:     tableNames,

		dbHandle:    handle,
		sqlExecutor: sqlExecutor,
		sqlBuilder:  &sqlBuilder,
		logger:      logger,
	}
}

func (d *Dumper) Dump() error {

	outputPathAbs, err := filepath.Abs(d.OutputDir)
	if err != nil {
		panic(err)
	}
	d.logger.Info(fmt.Sprintf("Dumping to %s", outputPathAbs))

	err = os.MkdirAll(outputPathAbs, 0755)
	if err != nil {
		panic(err)
	}

	return d.dbHandle.ReadOnly(func() error {
		for _, tableName := range d.TableNames {
			filePath := filepath.Join(d.OutputDir, fmt.Sprintf("%s.csv", tableName))
			d.logger.Info(fmt.Sprintf("Dumping %s to %s", tableName, filePath))
			columns, rows, err := d.queryTable(tableName)
			if err != nil {
				return err
			}
			f, err := os.Create(filePath)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			csvData := d.convertToCsvData(columns, rows)
			csvWriter := csv.NewWriter(f)

			err = csvWriter.WriteAll(csvData)
			if err != nil {
				panic(err)
			}
		}

		return nil
	})
}

func (d *Dumper) queryTable(tableName string) (columns []string, rows []map[string]string, err error) {
	q := d.sqlBuilder.Select("*").
		From(d.sqlBuilder.TableName(tableName)).
		Where("app_id = ANY (?)", pq.Array(d.AppIDs))

	qresult, err := d.sqlExecutor.QueryWith(q)
	if err != nil {
		return
	}
	defer qresult.Close()

	columns, err = qresult.Columns()
	if err != nil {
		return
	}
	for qresult.Next() {
		values := []any{}
		for range columns {
			var nullstr sql.NullString
			values = append(values, &nullstr)
		}

		err = qresult.Scan(values...)

		if err != nil {
			return
		}
		row := map[string]string{}
		for idx, col := range columns {
			nullstr := values[idx].(*sql.NullString)
			if !nullstr.Valid {
				row[col] = NULL
			} else {
				row[col] = nullstr.String
			}
		}
		rows = append(rows, row)
	}
	return
}

func (d *Dumper) convertToCsvData(columns []string, rows []map[string]string) [][]string {
	data := [][]string{}

	headerRow := []string{}
	for _, col := range columns {
		headerRow = append(headerRow, col)
	}
	data = append(data, headerRow)

	for _, row := range rows {
		dataRow := []string{}
		for _, col := range columns {
			dataRow = append(dataRow, row[col])
		}
		data = append(data, dataRow)
	}

	return data
}
