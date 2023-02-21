package accountmigration

type HookRequest struct {
	MigrationToken string `json:"migration_token"`
}
