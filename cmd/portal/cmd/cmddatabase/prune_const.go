package cmddatabase

// pruneTableNames lists every table that `authgear-portal database prune`
// deletes wholesale for the given app ids.
//
// _portal_config_source is deliberately excluded from this list: instead of
// deleting the row, prune scrubs its `data` column down to the minimum
// authgear.yaml/authgear.secrets.yaml needed to keep the app id from ever
// being re-registered (see pruneConfigSource in prune.go), so the row must
// survive.
//
// _portal_pending_domain holds transient, unverified custom-domain claims
// (see pkg/portal/service/domain.go) with no external system to desync, so
// unlike the billing tables in tablesExcludedFromPrune (prune_const_test.go)
// it is safe to bulk-delete here alongside its verified sibling
// _portal_domain, even though dump/restore doesn't support it (const.go).
var pruneTableNames []string = []string{
	"_portal_app_collaborator",
	"_portal_app_collaborator_invitation",
	"_portal_domain",
	"_portal_pending_domain",
	"_portal_tutorial_progress",
	"_portal_usage_record",
}
