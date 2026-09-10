package cmddatabase

import (
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/sqlmigrate"
)

// TestPruneTableNamesUpToDate fails when a migration adds or removes an
// app_id column without pruneTableNames being updated to match, so that
// `authgear database prune` cannot silently leave orphaned rows behind (or
// silently try to delete from a table that no longer exists).
func TestPruneTableNamesUpToDate(t *testing.T) {
	Convey("pruneTableNames covers every table with an app_id column", t, func() {
		actual, err := sqlmigrate.TablesWithColumn(mainMigrationFS, "migrations/authgear", "app_id")
		So(err, ShouldBeNil)

		declared := map[string]bool{}
		for _, name := range pruneTableNames {
			declared[name] = true
		}

		So(sortedKeysNotIn(actual, declared), ShouldBeEmpty)
		So(sortedKeysNotIn(declared, actual), ShouldBeEmpty)
	})
}

// sortedKeysNotIn returns the sorted keys of a that are absent from b.
func sortedKeysNotIn(a map[string]bool, b map[string]bool) []string {
	var out []string
	for name := range a {
		if !b[name] {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}
