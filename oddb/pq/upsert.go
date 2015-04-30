package pq

import (
	"bytes"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

// Let table = 'schema.device_table',
//     pks = ['id1', 'id2'],
//     data = {'id1': '1', 'id2': '2', type': 'ios', 'token': 'sometoken', 'userid': 'someuserid'}
//
// upsertQuery generates a query for upsert in the following format:
//
//	WITH updated AS (
//		UPDATE schema.device_table
//			SET ("id1", "id2", "type", "token", "user_id") =
//			($3, $4, $5, $6, $7)
//			WHERE "id1" = $1 AND "id2" = $2
//			RETURNING *
//		)
//	INSERT schema.device_table
//		("id1", "id2", "type", "token", "user_id")
//		SELECT $3, $4, $5, $6, $7
//		WHERE NOT EXISTS (SELECT * FROM updated);
//
// And args = ['1', '2', '1', '2', 'ios', 'sometoken', 'someuserid']
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
func upsertQuery(table string, pks []string, data map[string]interface{}) (sql string, args []interface{}) {
	// derive primary keys' data
	pkArgs := make([]interface{}, len(pks), len(pks))
	for i := range pks {
		value, ok := data[pks[i]]
		if !ok {
			log.WithFields(log.Fields{
				"table": table,
				"pks":   pks,
				"pk":    pks[i],
				"data":  data,
			}).Panicln("oddb/pq: primary key not found in data map")
		}
		pkArgs[i] = value
	}

	// extract columns values pair
	columns, args := extractKeyAndValue(data)

	b := bytes.Buffer{}
	b.Write([]byte(`WITH updated AS (UPDATE `))
	b.WriteString(table)
	b.Write([]byte(` SET(`))

	for _, column := range columns {
		b.WriteByte('"')
		b.WriteString(column)
		b.Write([]byte(`",`))
	}
	b.Truncate(b.Len() - 1)

	b.Write([]byte(`)=(`))

	for i := len(pks); i < len(pks)+len(columns); i++ {
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

	b.Write([]byte(` RETURNING *) INSERT INTO `))
	b.WriteString(table)
	b.WriteByte('(')

	for _, column := range columns {
		b.WriteByte('"')
		b.WriteString(column)
		b.Write([]byte(`",`))
	}
	b.Truncate(b.Len() - 1)

	b.Write([]byte(`) SELECT `))

	for i := len(pks); i < len(pks)+len(columns); i++ {
		b.WriteByte('$')
		b.WriteString(strconv.Itoa(i + 1))
		b.WriteByte(',')
	}
	b.Truncate(b.Len() - 1)

	b.Write([]byte(` WHERE NOT EXISTS (SELECT * FROM updated);`))

	return b.String(), append(pkArgs, args...)
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
