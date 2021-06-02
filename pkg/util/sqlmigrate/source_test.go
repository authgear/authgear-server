package sqlmigrate

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/rubenv/sql-migrate"
)

func TestTemplateMigrationSource(t *testing.T) {
	Convey("TemplateMigrationSource", t, func() {
		source := &TemplateMigrationSource{
			OriginSource: migrate.MemoryMigrationSource{
				Migrations: []*migrate.Migration{
					&migrate.Migration{
						Id:   "1",
						Up:   []string{"CREATE TABLE {{ .Schema }}.people (id int)"},
						Down: []string{"DROP TABLE {{ .Schema }}.people"},
					},
				},
			},
			Data: map[string]interface{}{
				"Schema": "myapp",
			},
		}
		migrations, err := source.FindMigrations()
		So(err, ShouldBeNil)
		So(migrations, ShouldResemble, []*migrate.Migration{
			&migrate.Migration{
				Id:   "1",
				Up:   []string{"CREATE TABLE myapp.people (id int)"},
				Down: []string{"DROP TABLE myapp.people"},
			},
		})
	})
}
