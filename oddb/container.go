package oddb

// A Container represents the set of data that is accessible by an app's
// user, which is currently provided by one public database and one
// private database.
type Container interface {
	PublicDB() Database
	PrivateDB(userKey string) Database
}
