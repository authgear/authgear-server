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
	"fmt"
	"strconv"
	"text/template"

	sq "github.com/lann/squirrel"
	"github.com/lib/pq"
)

const upsertTemplateText = `
{{define "commaSeparatedList"}}{{range $i, $_ := .}}{{if $i}}, {{end}}{{quoted .}}{{end}}{{end}}
WITH updated AS (
	{{if .UpdateCols }}
		UPDATE {{.Table}}
		SET ({{template "commaSeparatedList" .UpdateCols}}) = ({{placeholderList (len .Keys) (len .UpdateCols) .Wrappers}})
		WHERE {{range $i, $_ := .Keys}}{{if $i}} AND {{end}}{{quoted .}} = ${{addOne $i}}{{end}}
		RETURNING *
	{{else}}
		SELECT {{template "commaSeparatedList" .Keys}}
		FROM {{.Table}}
		WHERE {{range $i, $_ := .Keys}}{{if $i}} AND {{end}}{{quoted .}} = ${{addOne $i}}{{end}}
	{{end}}
), inserted AS (
	INSERT INTO {{.Table}}
		({{template "commaSeparatedList" .InsertCols}})
	SELECT {{placeholderList 0 (len .InsertCols) .Wrappers}}
	WHERE NOT EXISTS (SELECT * FROM updated)
	RETURNING *
)
SELECT * FROM updated
UNION ALL
SELECT * FROM inserted;
`

var funcMap = template.FuncMap{
	"addOne": func(n int) int { return n + 1 },
	"quoted": pq.QuoteIdentifier,
	"placeholderList": func(i, n int, wrappers map[int]func(string) string) string {
		b := bytes.Buffer{}
		from := i + 1
		to := from + n - 1
		for j := from; j <= to; j++ {
			if wrappers[j] != nil {
				b.WriteString(wrappers[j](fmt.Sprintf("$%d", j)))
			} else {
				b.WriteByte('$')
				b.WriteString(strconv.Itoa(j))
			}

			if j != to {
				b.WriteByte(',')
			}
		}
		return b.String()
	},
}

var upsertTemplate = template.Must(template.New("upsert").Funcs(funcMap).Parse(upsertTemplateText))

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
	updateIngnores map[string]struct{}
	wrappers       map[string]func(string) string
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
	return &upsertQueryBuilder{table, pkData, data, map[string]struct{}{}, map[string]func(string) string{}}
}

func upsertQueryWithWrappers(table string, pkData, data map[string]interface{}, wrappers map[string]func(string) string) *upsertQueryBuilder {
	return &upsertQueryBuilder{table, pkData, data, map[string]struct{}{}, wrappers}
}

func (upsert *upsertQueryBuilder) IgnoreKeyOnUpdate(col string) *upsertQueryBuilder {
	upsert.updateIngnores[col] = struct{}{}
	return upsert
}

// err always returns nil
func (upsert *upsertQueryBuilder) ToSql() (sql string, args []interface{}, err error) {
	// extract columns values pair
	pks, pkArgs := extractKeyAndValue(upsert.pkData)
	cols, args := extractKeyAndValue(upsert.data)

	cols, args, ignored := sortColsArgs(cols, args, upsert.updateIngnores)
	updateCols := cols[:len(cols)-ignored]

	b := bytes.Buffer{}

	insertCols := append(pks, cols...)
	wrappers := map[int]func(string) string{}

	for i, col := range insertCols {
		if upsert.wrappers[col] != nil {
			wrappers[i+1] = upsert.wrappers[col]
		}
	}

	err = upsertTemplate.Execute(&b, struct {
		Table      string
		Keys       []string
		UpdateCols []string
		InsertCols []string
		Wrappers   map[int]func(string) string
	}{
		Table:      upsert.table,
		Keys:       pks,
		UpdateCols: updateCols,
		InsertCols: insertCols,
		Wrappers:   wrappers,
	})
	if err != nil {
		panic(err)
	}

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

func sortColsArgs(cols []string, args []interface{}, ignoreCols map[string]struct{}) (sortedCols []string, sortedArgs []interface{}, ignore int) {
	var (
		ignored int
		c, ic   []string
		a, ia   []interface{}
	)

	for i, col := range cols {
		if _, ok := ignoreCols[col]; ok {
			ic = append(ic, col)
			ia = append(ia, args[i])
			ignored++
		} else {
			c = append(c, col)
			a = append(a, args[i])
		}
	}

	return append(c, ic...), append(a, ia...), ignored
}

var _ sq.Sqlizer = &upsertQueryBuilder{}
