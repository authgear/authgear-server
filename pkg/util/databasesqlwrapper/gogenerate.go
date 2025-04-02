package databasesqlwrapper

//go:generate go run github.com/authgear/authgear-server/pkg/util/databasesqlwrapper/codegen -package databasesqlwrapper -variant Driver
//go:generate go run github.com/authgear/authgear-server/pkg/util/databasesqlwrapper/codegen -package databasesqlwrapper -variant Connector
//go:generate go run github.com/authgear/authgear-server/pkg/util/databasesqlwrapper/codegen -package databasesqlwrapper -variant Conn
//go:generate go run github.com/authgear/authgear-server/pkg/util/databasesqlwrapper/codegen -package databasesqlwrapper -variant Stmt
//go:generate go run github.com/authgear/authgear-server/pkg/util/databasesqlwrapper/codegen -package databasesqlwrapper -variant Rows
//go:generate go run github.com/authgear/authgear-server/pkg/util/databasesqlwrapper/codegen -package databasesqlwrapper -variant Tx
