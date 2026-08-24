package cmddatabase

import (
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/sqlmigrate"
)

// tablesExcludedFromPrune mirrors the comment in const.go: these tables have
// an app_id column but dumping/restoring/pruning them is deliberately not
// supported (they mirror Stripe billing state, so bulk-deleting rows here
// would desync it). A table only belongs in this list if that reasoning
// still applies to it - do not add a table here just to silence this test.
var tablesExcludedFromPrune = map[string]bool{
	"_portal_historical_subscription": true,
	"_portal_pending_domain":          true,
	"_portal_subscription":            true,
	"_portal_subscription_checkout":   true,
}

// TestTableNamesUpToDate fails when a migration adds or removes an app_id
// column without tableNames (used by both dump/restore and
// `authgear-portal database prune`) being updated to match, so app data
// cannot silently be left un-pruned (or pruned from a dropped table).
func TestTableNamesUpToDate(t *testing.T) {
	Convey("tableNames plus the documented exclusions cover every table with an app_id column", t, func() {
		actual, err := sqlmigrate.TablesWithColumn(portalMigrationFS, "migrations/portal", "app_id")
		So(err, ShouldBeNil)

		declared := map[string]bool{}
		for _, name := range tableNames {
			declared[name] = true
		}
		for name := range tablesExcludedFromPrune {
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
