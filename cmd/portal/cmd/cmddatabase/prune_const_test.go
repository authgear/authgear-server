package cmddatabase

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/sqlmigrate"
)

// tablesExcludedFromPrune are tables with an app_id column that prune
// deliberately never touches: they mirror Stripe billing state, so
// bulk-deleting rows here would desync it. A table only belongs in this
// list if that reasoning still applies to it - do not add a table here
// just to silence this test.
var tablesExcludedFromPrune = map[string]bool{
	"_portal_historical_subscription": true,
	"_portal_subscription":            true,
	"_portal_subscription_checkout":   true,
}

// TestPruneTableNamesUpToDate fails when a migration adds or removes an
// app_id column without pruneTableNames (the tables `authgear-portal
// database prune` deletes wholesale) being updated to match.
// _portal_config_source is handled specially (see pruneConfigSource in
// prune.go) rather than deleted, and tablesExcludedFromPrune documents
// tables that are deliberately never touched by prune.
func TestPruneTableNamesUpToDate(t *testing.T) {
	Convey("pruneTableNames plus _portal_config_source and the documented exclusions cover every table with an app_id column", t, func() {
		actual, err := sqlmigrate.TablesWithColumn(portalMigrationFS, "migrations/portal", "app_id")
		So(err, ShouldBeNil)

		declared := map[string]bool{
			"_portal_config_source": true,
		}
		for _, name := range pruneTableNames {
			declared[name] = true
		}
		for name := range tablesExcludedFromPrune {
			declared[name] = true
		}

		So(sortedKeysNotIn(actual, declared), ShouldBeEmpty)
		So(sortedKeysNotIn(declared, actual), ShouldBeEmpty)
	})
}
