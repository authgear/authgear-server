package cmddatabase

// The order of the list is important because it defines the restoration order
// The referenced table must precede the referencing table
var tableNames []string = []string{
	"_portal_config_source",
	"_portal_app_collaborator",
	"_portal_app_collaborator_invitation",
	"_portal_domain",
	"_portal_tutorial_progress",
	"_portal_usage_record",
	// Dumping / Restoring of the following tables are not supportted
	// "_portal_historical_subscription",
	// "_portal_pending_domain",
	// "_portal_subscription",
	// "_portal_subscription_checkout",

}
