package e2e

import (
	"database/sql"
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
func ParseRows(rows *sql.Rows) (outputRows []interface{}, err error) {
	outputRows = []interface{}{}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return
	}

	nCol := len(columnTypes)

	for rows.Next() {
		scanArgs := make([]interface{}, nCol)
		for i, v := range columnTypes {
			switch v.DatabaseTypeName() {
			case "VARCHAR", "TEXT", "UUID", "TIMESTAMP":
				scanArgs[i] = new(sql.NullString)
				break
			case "BOOL":
				scanArgs[i] = new(sql.NullBool)
				break
			case "INT4":
				scanArgs[i] = new(sql.NullInt64)
				break
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
				row[v.Name()] = z.Bool
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullString); ok {
				row[v.Name()] = z.String
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
				row[v.Name()] = z.Int64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
				row[v.Name()] = z.Float64
				continue
			}

			if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
				row[v.Name()] = z.Int32
				continue
			}

			row[v.Name()] = scanArgs[i]
		}

		outputRows = append(outputRows, row)
	}

	return
}
