package pq

import (
	"bytes"
	"strconv"
)

// Let table = 'schema.device_table',
//     pks = ['id1', 'id2'],
//     data = {'type': 'ios', 'token': 'sometoken', 'userid': 'someuserid'}
//
// upsertQuery generates a query for upsert in the following format:
//
//	WITH updated AS (
//		UPDATE schema.device_table
//			SET ("type", "token", "user_id") =
//			($3, $4, $5)
//			WHERE "id1" = $1 AND "id2" = $2
//			RETURNING *
//		)
//	INSERT INTO schema.device_table
//		("id1", "id2", "type", "token", "user_id")
//		SELECT $1, $2, $3, $4, $5
//		WHERE NOT EXISTS (SELECT * FROM updated);
//
// And args = ['1', '2', 'ios', 'sometoken', 'someuserid']
//
// For empty data, following will be generated
//	WITH updated AS (
//		SELECT "id1", "id2" FROM schema.device_table
//		WHERE "id1" = $1 AND "id2" = $2
//	)
//	INSERT INTO schema.device_table
//		("id1", "id2")
//		SELECT $1, $2
//		WHERE NOT EXISTS (SELECT * FROM updated);
//
// And args = ['1', '2', 'ios', 'sometoken', 'someuserid']
//
// This approach uses CTE to do an INSERT after UPDATE in one query,
// hoping that the time gap between the UPDATE and INSERT is short
// enough that chance of a concurrent insert is rare.
//
// A complete upsert example is included in postgresql documentation [1],
// but that implementation contains a loop that does not guarantee
// exit. Adding that to poor performance, that implementation is not
// adopted.
//
// More on UPSERT: https://wiki.postgresql.org/wiki/UPSERT#PostgreSQL_.28today.29
//
// [1]: http://www.postgresql.org/docs/9.4/static/plpgsql-control-structures.html#PLPGSQL-UPSERT-EXAMPLE
type upsertQueryBuilder struct {
	table          string
	pkData         map[string]interface{}
	data           map[string]interface{}
	updateIngnores []string
}

// TODO(limouren): we can support a better fluent builder like this
//
//	upsert := upsertQuery(tableName).
//		WithKey("composite0", "0").WithKey("composite1", "1").
//		Set("string", "s").
//		Set("int", 1).
//		OnUpdate(func(upsert *upsertBuilder) {
//			upsert.Unset("deleteme")
//		})
//
func upsertQuery(table string, pkData, data map[string]interface{}) *upsertQueryBuilder {
	return &upsertQueryBuilder{table, pkData, data, nil}
}

func (upsert *upsertQueryBuilder) IgnoreKeyOnUpdate(ignore string) *upsertQueryBuilder {
	upsert.updateIngnores = append(upsert.updateIngnores, ignore)
	return upsert
}

// err always returns nil
func (upsert *upsertQueryBuilder) ToSql() (sql string, args []interface{}, err error) {
	// extract columns values pair
	pks, pkArgs := extractKeyAndValue(upsert.pkData)
	columns, args := extractKeyAndValue(upsert.data)
	ignoreIndex := findIgnoreIndex(columns, upsert.updateIngnores)

	b := bytes.Buffer{}
	if len(columns) > 0 {
		// Generate with UPDATE
		b.Write([]byte(`WITH updated AS (UPDATE `))
		b.WriteString(upsert.table)
		b.Write([]byte(` SET(`))

		for i, column := range columns {
			if ignoreIndex[i] {
				continue
			}
			b.WriteByte('"')
			b.WriteString(column)
			b.Write([]byte(`",`))
		}
		b.Truncate(b.Len() - 1)

		b.Write([]byte(`)=(`))

		for i := len(pks); i < len(pks)+len(columns); i++ {
			if ignoreIndex[i-len(pks)] {
				continue
			}
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(i + 1))
			b.WriteByte(',')
		}
		b.Truncate(b.Len() - 1)

		b.Write([]byte(`) WHERE `))

		for i, pk := range pks {
			b.WriteByte('"')
			b.WriteString(pk)
			b.Write([]byte(`" = $`))
			b.WriteString(strconv.Itoa(i + 1))
			b.Write([]byte(` AND `))
		}
		b.Truncate(b.Len() - 5)
		b.Write([]byte(` RETURNING *) `))
	} else {
		// Generate with SELECT
		b.Write([]byte(`WITH updated AS (SELECT `))
		for _, pk := range pks {
			b.WriteByte('"')
			b.WriteString(pk)
			b.Write([]byte(`",`))
		}
		b.Truncate(b.Len() - 1)
		b.Write([]byte(` FROM `))
		b.WriteString(upsert.table)
		b.Write([]byte(` WHERE `))
		for i, pk := range pks {
			b.WriteByte('"')
			b.WriteString(pk)
			b.Write([]byte(`" = $`))
			b.WriteString(strconv.Itoa(i + 1))
			b.Write([]byte(` AND `))
		}
		b.Truncate(b.Len() - 5)
		b.Write([]byte(`) `))
	}

	// generate INSERT
	b.Write([]byte(`INSERT INTO `))
	b.WriteString(upsert.table)
	b.WriteByte('(')

	for _, column := range append(pks, columns...) {
		b.WriteByte('"')
		b.WriteString(column)
		b.Write([]byte(`",`))
	}
	b.Truncate(b.Len() - 1)

	b.Write([]byte(`) SELECT `))

	for i := 0; i < len(pks)+len(columns); i++ {
		b.WriteByte('$')
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteByte(',')
	}
	b.Truncate(b.Len() - 1)

	b.Write([]byte(` WHERE NOT EXISTS (SELECT * FROM updated);`))

	return b.String(), append(pkArgs, args...), nil
}

func extractKeyAndValue(data map[string]interface{}) (keys []string, values []interface{}) {
	keys = make([]string, len(data), len(data))
	values = make([]interface{}, len(data), len(data))

	i := 0
	for key, value := range data {
		keys[i] = key
		values[i] = value
		i++
	}

	return
}

func findIgnoreIndex(columns []string, ignoreColumns []string) (ignoreIndex []bool) {
	ignoreIndex = make([]bool, len(columns), len(columns))

	for i, column := range columns {
		for _, ignored := range ignoreColumns {
			ignoreIndex[i] = (column == ignored)
			break
		}
	}

	return
}

var _ sqlizer = &upsertQueryBuilder{}
