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
	"database/sql"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

func (db *database) Extend(recordType string, recordSchema skydb.RecordSchema) (extended bool, err error) {
	remoteRecordSchema, err := db.RemoteColumnTypes(recordType)
	if err != nil {
		return
	}

	if len(remoteRecordSchema) > 0 && remoteRecordSchema.DefinitionCompatibleTo(recordSchema) {
		// The current record schema is superset of requested record
		// schema. There is no need to extend the schema.
		return
	}

	if !db.c.canMigrate {
		// The record schemas are different, but the database connection
		// does not allow migration.
		err = skyerr.NewError(
			skyerr.IncompatibleSchema,
			"Record schema requires migration but migration is disabled.",
		)
		return
	}

	// Begin transaction for schema migration
	tx, err := db.c.db.Beginx()
	if err != nil {
		return
	}
	defer tx.Rollback()

	if len(remoteRecordSchema) == 0 {
		if err := createTable(tx, db.TableName(recordType)); err != nil {
			return false, fmt.Errorf("failed to create table: %s", err)
		}
		extended = true
	}

	// Find new columns
	updatingSchema := skydb.RecordSchema{}
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
		stmt := db.addColumnStmt(recordType, updatingSchema)

		log.WithField("stmt", stmt).Debugln("Adding columns to table")
		if _, err := tx.Exec(stmt); err != nil {
			return false, fmt.Errorf("failed to alter table: %s", err)
		}

		extended = true
	}

	if err = tx.Commit(); err != nil {
		return false, fmt.Errorf("unable to commit transaction for Extend: %s", err)
	}

	delete(db.c.RecordSchema, recordType)

	return
}

func (db *database) RenameSchema(recordType, oldName, newName string) error {
	if !db.c.canMigrate {
		// The record schemas are different, but the database connection
		// does not allow migration.
		return skyerr.NewError(skyerr.IncompatibleSchema, "Record schema requires migration but migration is disabled.")
	}

	tableName := db.TableName(recordType)
	oldName = pq.QuoteIdentifier(oldName)
	newName = pq.QuoteIdentifier(newName)

	stmt := fmt.Sprintf("ALTER TABLE %s RENAME %s TO %s", tableName, oldName, newName)
	if _, err := db.c.Exec(stmt); err != nil {
		return fmt.Errorf("failed to alter table: %s", err)
	}
	return nil
}

func (db *database) DeleteSchema(recordType, columnName string) error {
	if !db.c.canMigrate {
		// The record schemas are different, but the database connection
		// does not allow migration.
		return skyerr.NewError(skyerr.IncompatibleSchema, "Record schema requires migration but migration is disabled.")
	}

	tableName := db.TableName(recordType)
	columnName = pq.QuoteIdentifier(columnName)

	stmt := fmt.Sprintf("ALTER TABLE %s DROP %s", tableName, columnName)
	if _, err := db.c.Exec(stmt); err != nil {
		return fmt.Errorf("failed to alter table: %s", err)
	}
	return nil
}

func (db *database) GetSchema(recordType string) (skydb.RecordSchema, error) {
	remoteRecordSchema, err := db.RemoteColumnTypes(recordType)
	if err != nil {
		return nil, err
	}
	return remoteRecordSchema, nil
}

func (db *database) GetRecordSchemas() (map[string]skydb.RecordSchema, error) {
	schemaName := db.schemaName()

	rows, err := db.c.Queryx(`
	SELECT table_name
	FROM information_schema.tables
	WHERE (table_name NOT LIKE '\_%') AND (table_schema=$1)
	`, schemaName)
	if err != nil {
		return nil, err
	}

	result := map[string]skydb.RecordSchema{}
	for rows.Next() {
		var recordType string
		if err := rows.Scan(&recordType); err != nil {
			return nil, err
		}

		log.Debugf("%s\n", recordType)
		schema, err := db.GetSchema(recordType)
		if err != nil {
			return nil, err
		}

		result[recordType] = schema
	}
	log.Debugf("GetRecordSchemas Success")

	return result, nil
}

func createTable(tx *sqlx.Tx, tableName string) error {
	stmt := createTableStmt(tableName)
	log.WithField("stmt", stmt).Debugln("Creating table")
	if _, err := tx.Exec(stmt); err != nil {
		return err
	}

	stmt = fmt.Sprintf(`
		CREATE TRIGGER trigger_notify_record_change
		AFTER INSERT OR UPDATE OR DELETE ON %s FOR EACH ROW
		EXECUTE PROCEDURE public.notify_record_change();
	`, tableName)
	log.WithField("stmt", stmt).Debugln("Creating trigger")
	if _, err := tx.Exec(stmt); err != nil {
		return err
	}

	return nil
}

func dropTable(tx *sqlx.Tx, tableName string) error {
	stmt := fmt.Sprintf(`
		DROP TRIGGER IF EXISTS trigger_notify_record_change
		ON %s
		CASCADE
	`, tableName)
	log.WithField("stmt", stmt).Debugln("Deleting trigger")
	if _, err := tx.Exec(stmt); err != nil {
		return err
	}

	stmt = fmt.Sprintf(`
		DROP TABLE IF EXISTS %s
		CASCADE
	`, tableName)
	log.WithField("stmt", stmt).Debugln("Deleting table")
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

func (db *database) getSequences(recordType string) ([]string, error) {
	const queryString = `
		SELECT c.relname
		FROM pg_catalog.pg_class c
			LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
		WHERE c.relname LIKE $1 AND n.nspname = $2;
	`

	rows, err := db.c.Queryx(
		queryString,
		fmt.Sprintf("%s\\_%%\\_seq", recordType),
		db.schemaName(),
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
func (db *database) RemoteColumnTypes(recordType string) (skydb.RecordSchema, error) {
	typemap := skydb.RecordSchema{}
	var err error
	// STEP 0: Return the cached ColumnType
	if schema, ok := db.c.RecordSchema[recordType]; ok {
		log.Debugf("Using cached remoteColumnTypes %s", recordType)
		return schema, nil
	}
	log.Debugf("Querying remoteColumnTypes %s", recordType)
	// STEP 1: Get the oid of the current table
	var oid int
	err = db.c.QueryRowx(`
SELECT c.oid
FROM pg_catalog.pg_class c
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE c.relname = $1
  AND n.nspname = $2`,
		recordType, db.schemaName()).Scan(&oid)

	if err == sql.ErrNoRows {
		db.c.RecordSchema[recordType] = nil
		log.Debugf("Cache remoteColumnTypes %s (no table)", recordType)
		return nil, nil
	}
	if err != nil {
		log.WithFields(logrus.Fields{
			"schemaName": db.schemaName(),
			"recordType": recordType,
			"err":        err,
		}).Errorln("Failed to query oid of table")
		return nil, err
	}

	// STEP 2: Get column name and data type
	rows, err := db.c.Queryx(`
SELECT a.attname,
  pg_catalog.format_type(a.atttypid, a.atttypmod)
FROM pg_catalog.pg_attribute a
WHERE a.attrelid = $1 AND a.attnum > 0 AND NOT a.attisdropped`,
		oid)

	if err != nil {
		log.WithFields(logrus.Fields{
			"schemaName": db.schemaName(),
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

		schema := skydb.FieldType{
			UnderlyingType: pqType,
		}
		switch pqType {
		case TypeCaseInsensitiveString:
			fallthrough
		case TypeString:
			schema.Type = skydb.TypeString
		case TypeNumber:
			schema.Type = skydb.TypeNumber
		case TypeTimestamp:
			schema.Type = skydb.TypeDateTime
		case TypeBoolean:
			schema.Type = skydb.TypeBoolean
		case TypeJSON:
			if columnName == "_access" {
				schema.Type = skydb.TypeACL
			} else {
				schema.Type = skydb.TypeJSON
			}
		case TypeLocation:
			schema.Type = skydb.TypeLocation
		case TypeBigInteger:
			fallthrough
		case TypeInteger:
			schema.Type = skydb.TypeInteger
			integerColumns = append(integerColumns, columnName)
		case TypeGeometry:
			schema.Type = skydb.TypeGeometry
		default:
			schema.Type = skydb.TypeUnknown
		}

		typemap[columnName] = schema
	}

	// STEP 2.1: Convert integer column to sequence column if applicable
	if len(integerColumns) > 0 {
		sequenceList, err := db.getSequences(recordType)
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
				schema.Type = skydb.TypeSequence

				typemap[perIntColumn] = schema
			}
		}
	}

	// STEP 3: FOREIGN KEY, assumeing we can only reference _id i.e. "ccu.column_name" = _id
	builder := psql.Select("kcu.column_name", "ccu.table_name").
		From("information_schema.table_constraints AS tc").
		Join("information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name").
		Join("information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name").
		Where("constraint_type = 'FOREIGN KEY' AND tc.table_schema = ? AND tc.table_name = ?", db.schemaName(), recordType)

	refs, err := db.c.QueryWith(builder)
	if err != nil {
		log.WithFields(logrus.Fields{
			"schemaName": db.schemaName(),
			"recordType": recordType,
			"err":        err,
		}).Errorln("Failed to query foreign key information schema")

		return nil, err
	}

	for refs.Next() {
		s := skydb.FieldType{}
		var primaryColumn, referencedTable string
		if err := refs.Scan(&primaryColumn, &referencedTable); err != nil {
			log.Debugf("err %v", err)
			return nil, err
		}
		switch referencedTable {
		case "_asset":
			s.Type = skydb.TypeAsset
		default:
			s.Type = skydb.TypeReference
			s.ReferenceType = referencedTable
		}
		typemap[primaryColumn] = s
	}

	db.c.RecordSchema[recordType] = typemap
	log.Debugf("Cache remoteColumnTypes %s", recordType)
	return typemap, nil
}

// ALTER TABLE app__.note add collection text;
// ALTER TABLE app__.note
// ADD CONSTRAINT fk_note_collection_collection
// FOREIGN KEY (collection)
// REFERENCES app__.collection(_id);
func (db *database) addColumnStmt(recordType string, recordSchema skydb.RecordSchema) string {
	buf := bytes.Buffer{}
	buf.Write([]byte("ALTER TABLE "))
	buf.WriteString(db.TableName(recordType))
	buf.WriteByte(' ')
	for column, schema := range recordSchema {
		buf.Write([]byte("ADD "))
		buf.WriteString(pq.QuoteIdentifier(column))
		buf.WriteByte(' ')
		buf.WriteString(pqDataType(schema.Type))
		buf.WriteByte(',')
		switch schema.Type {
		case skydb.TypeAsset:
			db.writeForeignKeyConstraint(&buf, column, "_asset", "id")
		case skydb.TypeReference:
			db.writeForeignKeyConstraint(&buf, column, schema.ReferenceType, "_id")
		}
	}

	// remote the last ','
	buf.Truncate(buf.Len() - 1)

	return buf.String()
}

func (db *database) writeForeignKeyConstraint(buf *bytes.Buffer, localCol, referent, remoteCol string) {
	buf.Write([]byte(`ADD CONSTRAINT `))
	buf.WriteString(pq.QuoteIdentifier(fmt.Sprintf(`fk_%s_%s_%s`, localCol, referent, remoteCol)))
	buf.Write([]byte(` FOREIGN KEY (`))
	buf.WriteString(pq.QuoteIdentifier(localCol))
	buf.Write([]byte(`) REFERENCES `))
	buf.WriteString(db.TableName(referent))
	buf.Write([]byte(` (`))
	buf.WriteString(pq.QuoteIdentifier(remoteCol))
	buf.Write([]byte(`),`))
}

func (db *database) GetIndexesByRecordType(recordType string) (indexes map[string]skydb.Index, err error) {
	schemaName := db.schemaName()
	rows, err := db.c.Queryx(`
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

	indexes = map[string]skydb.Index{}
	for rows.Next() {
		var table string
		var name string
		var columnNames string
		if err = rows.Scan(&table, &name, &columnNames); err != nil {
			return
		}

		indexes[name] = skydb.Index{
			Fields: strings.Split(columnNames, ","),
		}
	}

	return
}

func (db *database) SaveIndex(recordType, indexName string, index skydb.Index) error {
	quotedColumns := []string{}
	for _, col := range index.Fields {
		quotedColumns = append(quotedColumns, fmt.Sprintf("%s", col))
	}

	stmt := fmt.Sprintf(`
		ALTER TABLE "%s"."%s" ADD CONSTRAINT %s UNIQUE (%s);
	`, db.schemaName(), recordType, indexName, strings.Join(quotedColumns, ","))
	fmt.Println("Save Index", stmt)
	log.WithField("stmt", stmt).Debugln("Creating unique constraint")
	if _, err := db.c.Exec(stmt); err != nil {
		return err
	}

	return nil
}

func (db *database) DeleteIndex(recordType string, indexName string) error {
	stmt := fmt.Sprintf(`
		ALTER TABLE "%s"."%s" DROP CONSTRAINT %s;
	`, db.schemaName(), recordType, indexName)
	log.WithField("stmt", stmt).Debugln("Dropping unique constraint")
	if _, err := db.c.Exec(stmt); err != nil {
		return err
	}

	return nil
}
