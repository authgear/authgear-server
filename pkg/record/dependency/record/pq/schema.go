// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pq

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/record/dependency/record"
	"github.com/skygeario/skygear-server/pkg/server/logging"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func (s *RecordStore) Extend(recordType string, recordSchema record.Schema) (extended bool, err error) {
	remoteRecordSchema, err := s.RemoteColumnTypes(recordType)
	if err != nil {
		return
	}

	if len(remoteRecordSchema) > 0 && remoteRecordSchema.DefinitionCompatibleTo(recordSchema) {
		// The current record schema is superset of requested record
		// schema. There is no need to extend the schema.
		return
	}

	if !s.canMigrate {
		// The record schemas are different, but the database connection
		// does not allow migration.
		err = skyerr.NewError(
			skyerr.IncompatibleSchema,
			"Record schema requires migration but migration is disabled.",
		)
		return
	}

	if len(remoteRecordSchema) == 0 {
		if err := createTable(s.logger, s.sqlExecutor, s.sqlBuilder.FullTableName(recordType)); err != nil {
			return false, fmt.Errorf("failed to create table: %s", err)
		}
		extended = true
	}

	// Find new columns
	updatingSchema := record.Schema{}
	for key, fieldType := range recordSchema {
		if remoteFieldType, ok := remoteRecordSchema[key]; ok {
			if !remoteFieldType.DefinitionCompatibleTo(fieldType) {
				return false, skyerr.NewError(
					skyerr.IncompatibleSchema,
					fmt.Sprintf("conflicting schema %v => %v", remoteFieldType, fieldType),
				)
			}
		} else {
			updatingSchema[key] = fieldType
		}
	}

	if len(updatingSchema) > 0 {
		stmt := s.addColumnStmt(recordType, updatingSchema)

		s.logger.WithField("stmt", stmt).Debugln("Adding columns to table")
		if _, err := s.sqlExecutor.Exec(stmt); err != nil {
			return false, fmt.Errorf("failed to alter table: %s", err)
		}

		extended = true
	}

	// delete(db.c.RecordSchema, recordType)

	return
}

func (s *RecordStore) RenameSchema(recordType, oldName, newName string) error {
	if !s.canMigrate {
		// The record schemas are different, but the database connection
		// does not allow migration.
		return skyerr.NewError(skyerr.IncompatibleSchema, "Record schema requires migration but migration is disabled.")
	}

	tableName := s.sqlBuilder.FullTableName(recordType)
	oldName = pq.QuoteIdentifier(oldName)
	newName = pq.QuoteIdentifier(newName)

	stmt := fmt.Sprintf("ALTER TABLE %s RENAME %s TO %s", tableName, oldName, newName)
	if _, err := s.sqlExecutor.Exec(stmt); err != nil {
		return fmt.Errorf("failed to alter table: %s", err)
	}
	return nil
}

func (s *RecordStore) DeleteSchema(recordType, columnName string) error {
	if !s.canMigrate {
		// The record schemas are different, but the database connection
		// does not allow migration.
		return skyerr.NewError(skyerr.IncompatibleSchema, "Record schema requires migration but migration is disabled.")
	}

	tableName := s.sqlBuilder.FullTableName(recordType)
	columnName = pq.QuoteIdentifier(columnName)

	stmt := fmt.Sprintf("ALTER TABLE %s DROP %s", tableName, columnName)
	if _, err := s.sqlExecutor.Exec(stmt); err != nil {
		return fmt.Errorf("failed to alter table: %s", err)
	}
	return nil
}

func (s *RecordStore) GetSchema(recordType string) (record.Schema, error) {
	remoteRecordSchema, err := s.RemoteColumnTypes(recordType)
	if err != nil {
		return nil, err
	}
	return remoteRecordSchema, nil
}

func (s *RecordStore) GetRecordSchemas() (map[string]record.Schema, error) {
	schemaName := s.sqlBuilder.SchemaName()

	rows, err := s.sqlExecutor.Queryx(`
	SELECT table_name
	FROM information_schema.tables
	WHERE (table_name NOT LIKE '\_%') AND (table_schema=$1)
	`, schemaName)
	if err != nil {
		return nil, err
	}

	result := map[string]record.Schema{}
	for rows.Next() {
		var recordType string
		if err := rows.Scan(&recordType); err != nil {
			return nil, err
		}

		s.logger.Debugf("%s\n", recordType)
		schema, err := s.GetSchema(recordType)
		if err != nil {
			return nil, err
		}

		result[recordType] = schema
	}
	s.logger.Debugf("GetRecordSchemas Success")

	return result, nil
}

func createTable(logger *logrus.Entry, sqlExecutor db.SQLExecutor, tableName string) error {
	stmt := createTableStmt(tableName)
	logger.WithField("stmt", stmt).Debugln("Creating table")
	if _, err := sqlExecutor.Exec(stmt); err != nil {
		return err
	}

	stmt = fmt.Sprintf(`
		CREATE TRIGGER trigger_notify_record_change
		AFTER INSERT OR UPDATE OR DELETE ON %s FOR EACH ROW
		EXECUTE PROCEDURE public.notify_record_change();
	`, tableName)
	logger.WithField("stmt", stmt).Debugln("Creating trigger")
	if _, err := sqlExecutor.Exec(stmt); err != nil {
		return err
	}

	return nil
}

func dropTable(ctx context.Context, tx *sqlx.Tx, tableName string) error {
	logger := logging.CreateLogger(ctx, "skydb")
	stmt := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS trigger_notify_record_change
		ON %s
		CASCADE
	`, tableName)
	logger.WithField("stmt", stmt).Debugln("Deleting trigger")
	if _, err := tx.Exec(stmt); err != nil {
		return err
	}

	stmt = fmt.Sprintf(`
		DROP TABLE IF EXISTS %s
		CASCADE
	`, tableName)
	logger.WithField("stmt", stmt).Debugln("Deleting table")
	if _, err := tx.Exec(stmt); err != nil {
		return err
	}

	return nil
}

func createTableStmt(tableName string) string {
	return fmt.Sprintf(`
CREATE TABLE %s (
    _id text,
    _database_id text,
    _owner_id text,
    _access jsonb,
    _created_at timestamp without time zone NOT NULL,
    _created_by text,
    _updated_at timestamp without time zone NOT NULL,
    _updated_by text,
    PRIMARY KEY(_id, _database_id, _owner_id),
    UNIQUE (_id)
);
`, tableName)
}

func (s *RecordStore) getSequences(recordType string) ([]string, error) {
	const queryString = `
		SELECT c.relname
		FROM pg_catalog.pg_class c
			LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relname LIKE $1 AND n.nspname = $2;
	`

	rows, err := s.sqlExecutor.Queryx(
		queryString,
		fmt.Sprintf("%s\\_%%\\_seq", recordType),
		s.sqlBuilder.SchemaName(),
	)
	if err != nil {
		return []string{}, err
	}

	seqList := []string{}

	for rows.Next() {
		var relname string
		if err = rows.Scan(&relname); err != nil {
			return []string{}, err
		}

		relname = strings.TrimPrefix(relname, fmt.Sprintf("%s_", recordType))
		relname = strings.TrimSuffix(relname, "_seq")

		seqList = append(seqList, relname)
	}

	return seqList, nil
}

// RemoteColumnTypes returns a typemap of a database table.
// STEP 1 & 2 are obtained by reverse engineering psql \d with -E option
//
// STEP 3: example of getting foreign keys
// SELECT
//     tc.table_name, kcu.column_name,
//     ccu.table_name AS foreign_table_name,
//     ccu.column_name AS foreign_column_name
// FROM
//     information_schema.table_constraints AS tc
//     JOIN information_schema.key_column_usage
//         AS kcu ON tc.constraint_name = kcu.constraint_name
//     JOIN information_schema.constraint_column_usage
//         AS ccu ON ccu.constraint_name = tc.constraint_name
// WHERE constraint_type = 'FOREIGN KEY'
// AND tc.table_schema = 'app__'
// AND tc.table_name = 'note';
// nolint: gocyclo
func (s *RecordStore) RemoteColumnTypes(recordType string) (record.Schema, error) {
	typemap := record.Schema{}
	var err error
	// STEP 0: Return the cached ColumnType
	// if schema, ok := db.c.RecordSchema[recordType]; ok {
	// 	logger.Debugf("Using cached remoteColumnTypes %s", recordType)
	// 	return schema, nil
	// }
	s.logger.Debugf("Querying remoteColumnTypes %s", recordType)
	// STEP 1: Get the oid of the current table
	var oid int
	err = s.sqlExecutor.QueryRowx(`
SELECT c.oid
FROM pg_catalog.pg_class c
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE c.relname = $1
  AND n.nspname = $2`,
		recordType, s.sqlBuilder.SchemaName()).Scan(&oid)

	if err == sql.ErrNoRows {
		// db.c.RecordSchema[recordType] = nil
		s.logger.Debugf("Cache remoteColumnTypes %s (no table)", recordType)
		return nil, nil
	}
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"schemaName": s.sqlBuilder.SchemaName(),
			"recordType": recordType,
			"err":        err,
		}).Errorln("Failed to query oid of table")
		return nil, err
	}

	// STEP 2: Get column name and data type
	rows, err := s.sqlExecutor.Queryx(`
SELECT a.attname,
  pg_catalog.format_type(a.atttypid, a.atttypmod)
FROM pg_catalog.pg_attribute a
WHERE a.attrelid = $1 AND a.attnum > 0 AND NOT a.attisdropped`,
		oid)

	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"schemaName": s.sqlBuilder.SchemaName(),
			"recordType": recordType,
			"oid":        oid,
			"err":        err,
		}).Errorln("Failed to query column and data type")
		return nil, err
	}

	var columnName, pqType string
	var integerColumns = []string{}
	for rows.Next() {
		if err := rows.Scan(&columnName, &pqType); err != nil {
			return nil, err
		}

		schema := record.FieldType{
			UnderlyingType: pqType,
		}
		switch pqType {
		case TypeCaseInsensitiveString:
			fallthrough
		case TypeString:
			schema.Type = record.TypeString
		case TypeNumber:
			schema.Type = record.TypeNumber
		case TypeTimestamp:
			schema.Type = record.TypeDateTime
		case TypeBoolean:
			schema.Type = record.TypeBoolean
		case TypeJSON:
			if columnName == "_access" {
				schema.Type = record.TypeACL
			} else {
				schema.Type = record.TypeJSON
			}
		case TypeLocation:
			schema.Type = record.TypeLocation
		case TypeBigInteger:
			fallthrough
		case TypeInteger:
			schema.Type = record.TypeInteger
			integerColumns = append(integerColumns, columnName)
		case TypeGeometry:
			schema.Type = record.TypeGeometry
		default:
			schema.Type = record.TypeUnknown
		}

		typemap[columnName] = schema
	}

	// STEP 2.1: Convert integer column to sequence column if applicable
	if len(integerColumns) > 0 {
		sequenceList, err := s.getSequences(recordType)
		if err != nil {
			return nil, err
		}

		sequenceMap := map[string]bool{}
		for _, perSeq := range sequenceList {
			sequenceMap[perSeq] = true
		}

		for _, perIntColumn := range integerColumns {
			if _, ok := sequenceMap[perIntColumn]; ok {
				schema := typemap[perIntColumn]
				schema.Type = record.TypeSequence

				typemap[perIntColumn] = schema
			}
		}
	}

	// STEP 3: FOREIGN KEY, assumeing we can only reference _id i.e. "ccu.column_name" = _id
	builder := s.sqlBuilder.Select("kcu.column_name", "ccu.table_name").
		From("information_schema.table_constraints AS tc").
		Join("information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name").
		Join("information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name").
		Where("constraint_type = 'FOREIGN KEY' AND tc.table_schema = ? AND tc.table_name = ?", s.sqlBuilder.SchemaName(), recordType)

	refs, err := s.sqlExecutor.QueryWith(builder)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"schemaName": s.sqlBuilder.SchemaName(),
			"recordType": recordType,
			"err":        err,
		}).Errorln("Failed to query foreign key information schema")

		return nil, err
	}

	for refs.Next() {
		ft := record.FieldType{}
		var primaryColumn, referencedTable string
		if err := refs.Scan(&primaryColumn, &referencedTable); err != nil {
			s.logger.Debugf("err %v", err)
			return nil, err
		}
		switch referencedTable {
		case "_asset":
			ft.Type = record.TypeAsset
		default:
			ft.Type = record.TypeReference
			ft.ReferenceType = referencedTable
		}
		typemap[primaryColumn] = ft
	}

	// db.c.RecordSchema[recordType] = typemap
	s.logger.Debugf("Cache remoteColumnTypes %s", recordType)
	return typemap, nil
}

// ALTER TABLE app__.note add collection text;
// ALTER TABLE app__.note
// ADD CONSTRAINT fk_note_collection_collection
// FOREIGN KEY (collection)
// REFERENCES app__.collection(_id);
func (s *RecordStore) addColumnStmt(recordType string, recordSchema record.Schema) string {
	buf := bytes.Buffer{}
	buf.Write([]byte("ALTER TABLE "))
	buf.WriteString(s.sqlBuilder.FullTableName(recordType))
	buf.WriteByte(' ')
	for column, schema := range recordSchema {
		buf.Write([]byte("ADD "))
		buf.WriteString(pq.QuoteIdentifier(column))
		buf.WriteByte(' ')
		buf.WriteString(pqDataType(schema.Type))
		buf.WriteByte(',')
		switch schema.Type {
		case record.TypeAsset:
			s.writeForeignKeyConstraint(&buf, column, "_asset", "id")
		case record.TypeReference:
			s.writeForeignKeyConstraint(&buf, column, schema.ReferenceType, "_id")
		}
	}

	// remote the last ','
	buf.Truncate(buf.Len() - 1)

	return buf.String()
}

func (s *RecordStore) writeForeignKeyConstraint(buf *bytes.Buffer, localCol, referent, remoteCol string) {
	buf.Write([]byte(`ADD CONSTRAINT `))
	buf.WriteString(pq.QuoteIdentifier(fmt.Sprintf(`fk_%s_%s_%s`, localCol, referent, remoteCol)))
	buf.Write([]byte(` FOREIGN KEY (`))
	buf.WriteString(pq.QuoteIdentifier(localCol))
	buf.Write([]byte(`) REFERENCES `))
	buf.WriteString(s.sqlBuilder.FullTableName(referent))
	buf.Write([]byte(` (`))
	buf.WriteString(pq.QuoteIdentifier(remoteCol))
	buf.Write([]byte(`),`))
}

func (s *RecordStore) GetIndexesByRecordType(recordType string) (indexes map[string]record.Index, err error) {
	schemaName := s.sqlBuilder.SchemaName()
	rows, err := s.sqlExecutor.Queryx(`
SELECT
    t.relname AS table_name,
    i.relname AS index_name,
    array_to_string(array_agg(a.attname), ',') AS column_names
FROM
    pg_class t,
    pg_class i,
    pg_index ix,
    pg_attribute a,
    pg_namespace ns
WHERE
    t.oid = ix.indrelid
    AND i.oid = ix.indexrelid
    AND ns.oid = t.relnamespace
    AND ns.oid = i.relnamespace
    AND a.attrelid = t.oid
    AND a.attnum = ANY(ix.indkey)
    AND t.relkind = 'r'
    AND ix.indisunique = TRUE
    AND ns.nspname = $1
    AND t.relname = $2
GROUP BY
    ns.nspname,
    t.relname,
    i.relname;`,
		schemaName, recordType)

	if err != nil {
		return
	}

	indexes = map[string]record.Index{}
	for rows.Next() {
		var table string
		var name string
		var columnNames string
		if err = rows.Scan(&table, &name, &columnNames); err != nil {
			return
		}

		indexes[name] = record.Index{
			Fields: strings.Split(columnNames, ","),
		}
	}

	return
}

func (s *RecordStore) SaveIndex(recordType, indexName string, index record.Index) error {
	quotedColumns := []string{}
	for _, col := range index.Fields {
		quotedColumns = append(quotedColumns, fmt.Sprintf("%s", col))
	}

	stmt := fmt.Sprintf(`
		ALTER TABLE "%s"."%s" ADD CONSTRAINT %s UNIQUE (%s);
	`, s.sqlBuilder.SchemaName(), recordType, indexName, strings.Join(quotedColumns, ","))
	fmt.Println("Save Index", stmt)
	s.logger.WithField("stmt", stmt).Debugln("Creating unique constraint")
	if _, err := s.sqlExecutor.Exec(stmt); err != nil {
		return err
	}

	return nil
}

func (s *RecordStore) DeleteIndex(recordType string, indexName string) error {
	stmt := fmt.Sprintf(`
		ALTER TABLE "%s"."%s" DROP CONSTRAINT %s;
	`, s.sqlBuilder.SchemaName(), recordType, indexName)
	s.logger.WithField("stmt", stmt).Debugln("Dropping unique constraint")
	if _, err := s.sqlExecutor.Exec(stmt); err != nil {
		return err
	}

	return nil
}
