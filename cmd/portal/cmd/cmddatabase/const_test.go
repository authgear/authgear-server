package cmddatabase

import (
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/sqlmigrate"
)

// tablesExcludedFromDumpRestore mirrors the comment in const.go: these
// tables have an app_id column but dumping/restoring them is deliberately
// not supported. A table only belongs in this list if that reasoning still
// applies to it - do not add a table here just to silence this test.
//
// _portal_pending_domain is here (dump/restore doesn't support it) but is
// NOT excluded from prune - see tablesExcludedFromPrune in
// prune_const_test.go.
var tablesExcludedFromDumpRestore = map[string]bool{
	"_portal_historical_subscription": true,
	"_portal_pending_domain":          true,
	"_portal_subscription":            true,
	"_portal_subscription_checkout":   true,
}

// TestTableNamesUpToDate fails when a migration adds or removes an app_id
// column without tableNames (used by dump/restore) being updated to match.
func TestTableNamesUpToDate(t *testing.T) {
	Convey("tableNames plus the documented exclusions cover every table with an app_id column", t, func() {
		actual, err := sqlmigrate.TablesWithColumn(portalMigrationFS, "migrations/portal", "app_id")
		So(err, ShouldBeNil)

		declared := map[string]bool{}
		for _, name := range tableNames {
			declared[name] = true
		}
		for name := range tablesExcludedFromDumpRestore {
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
