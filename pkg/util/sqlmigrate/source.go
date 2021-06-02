package sqlmigrate

import (
	"strings"
	"text/template"

	"github.com/rubenv/sql-migrate"
)

type TemplateMigrationSource struct {
	OriginSource migrate.MigrationSource
	Data         interface{}
}

func (s TemplateMigrationSource) FindMigrations() (migrations []*migrate.Migration, err error) {
	migrations, err = s.OriginSource.FindMigrations()
	if err != nil {
		return
	}

	for _, migration := range migrations {
		var ups []string
		var downs []string
		for _, up := range migration.Up {
			var out string
			out, err = s.ExecuteTemplate(up)
			if err != nil {
				return
			}
			ups = append(ups, out)
		}
		for _, down := range migration.Down {
			var out string
			out, err = s.ExecuteTemplate(down)
			if err != nil {
				return
			}
			downs = append(downs, out)
		}
		migration.Up = ups
		migration.Down = downs
	}

	return
}

func (s TemplateMigrationSource) ExecuteTemplate(input string) (out string, err error) {
	tmpl, err := template.New("sql").Parse(input)
	if err != nil {
		return
	}

	buf := strings.Builder{}
	err = tmpl.Execute(&buf, s.Data)
	if err != nil {
		return
	}

	out = buf.String()
	return
}
