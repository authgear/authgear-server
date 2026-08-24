package sqlmigrate

import (
	"embed"
	"fmt"
	"regexp"

	migrate "github.com/rubenv/sql-migrate"
)

var reCreateTable = regexp.MustCompile(`(?is)^\s*CREATE TABLE\s+(?:IF NOT EXISTS\s+)?"?([A-Za-z0-9_]+)"?\s*\((.*)\)\s*;?\s*$`)
var reDropTable = regexp.MustCompile(`(?is)^\s*DROP TABLE\s+(?:IF EXISTS\s+)?"?([A-Za-z0-9_]+)"?\s*;?\s*$`)

// TablesWithColumn statically replays every migration's Up statements, in
// migration order, and returns the set of tables that currently exist (i.e.
// have been CREATE TABLE'd and not since DROP TABLE'd) and were created
// with a column named columnName. It requires no database connection.
//
// It only understands the plain `CREATE TABLE name (...)` and
// `DROP TABLE name` statement forms used by this repo's migrations today
// (no ALTER TABLE ADD/DROP COLUMN, no CREATE TABLE ... AS, no multi-table
// DROP TABLE a, b). If a future migration needs one of those forms for a
// table this function is used to track, this function needs updating too.
func TablesWithColumn(fsys embed.FS, root string, columnName string) (map[string]bool, error) {
	source := migrate.EmbedFileSystemMigrationSource{
		FileSystem: fsys,
		Root:       root,
	}
	migrations, err := source.FindMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to parse migrations: %w", err)
	}

	columnPattern := regexp.MustCompile(`(?im)^\s*"?` + regexp.QuoteMeta(columnName) + `"?\b`)

	exists := map[string]bool{}
	hasColumn := map[string]bool{}
	for _, m := range migrations {
		for _, stmt := range m.Up {
			if match := reCreateTable.FindStringSubmatch(stmt); match != nil {
				name, body := match[1], match[2]
				exists[name] = true
				if columnPattern.MatchString(body) {
					hasColumn[name] = true
				} else {
					delete(hasColumn, name)
				}
				continue
			}
			if match := reDropTable.FindStringSubmatch(stmt); match != nil {
				name := match[1]
				delete(exists, name)
				delete(hasColumn, name)
				continue
			}
		}
	}

	result := map[string]bool{}
	for name := range hasColumn {
		if exists[name] {
			result[name] = true
		}
	}
	return result, nil
}
