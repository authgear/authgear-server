package e2e

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig"
)

func ParseSQLTemplate(name string, rawTemplateText string) (*template.Template, error) {
	tmpl := template.New(name)
	tmpl.Funcs(sprig.GenericFuncMap())
	sqlTmpl, err := tmpl.Parse(rawTemplateText)
	return sqlTmpl, err
}

// copied from https://stackoverflow.com/a/60386531/19287186
func ParseRows(rows *sql.Rows) (outputRows []map[string]interface{}, err error) {
	outputRows = []map[string]interface{}{}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return
	}

	nCol := len(columnTypes)

	for rows.Next() {
		scanArgs := make([]interface{}, nCol)
		for i, v := range columnTypes {
			switch v.DatabaseTypeName() {
			case "VARCHAR", "TEXT", "UUID", "JSONB":
				scanArgs[i] = new(sql.NullString)
				break
			case "BOOL":
				scanArgs[i] = new(sql.NullBool)
				break
			case "INT4":
				scanArgs[i] = new(sql.NullInt64)
				break
			case "TIMESTAMP":
				scanArgs[i] = new(sql.NullTime)
			default:
				scanArgs[i] = new(sql.NullString)
			}
		}

		err = rows.Scan(scanArgs...)

		if err != nil {
			return
		}

		row := map[string]interface{}{}

		for i, v := range columnTypes {

			if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
				if z.Valid { // not null
					row[v.Name()] = z.Bool
				} else { // null
					row[v.Name()] = nil
				}
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
				if z.Valid { // not null
					databaseTypeName := v.DatabaseTypeName()
					if databaseTypeName == "JSONB" {
						var value any
						err = json.Unmarshal([]byte(z.String), &value)
						if err != nil {
							return
						}
						row[v.Name()] = value
					} else {
						row[v.Name()] = z.String
					}
				} else { // null
					row[v.Name()] = nil
				}
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
				if z.Valid { // not null
					row[v.Name()] = z.Int64
				} else { // null
					row[v.Name()] = nil
				}
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
				if z.Valid { // not null
					row[v.Name()] = z.Float64
				} else { // null
					row[v.Name()] = nil
				}
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
				if z.Valid { // not null
					row[v.Name()] = z.Int32
				} else { // null
					row[v.Name()] = nil
				}
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullTime); ok {
				if z.Valid { // not null
					row[v.Name()] = z.Time
				} else { // null
					row[v.Name()] = nil
				}
				continue
			}

			panic(fmt.Errorf("unknown datatype: (colName=%v, value=%v)", v.Name(), scanArgs[i]))
		}

		outputRows = append(outputRows, row)
	}

	return
}
