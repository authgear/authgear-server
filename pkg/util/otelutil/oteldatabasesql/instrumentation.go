package oteldatabasesql

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/authgear/authgear-server/pkg/util/databasesqlwrapper"
)

func Open(driverName string, dsn string) (*sql.DB, error) {
	var wrapDriver func(d driver.Driver) driver.Driver
	var wrapConnector func(c driver.Connector) driver.Connector
	var wrapConn func(c driver.Conn) driver.Conn
	var wrapStmt func(s driver.Stmt) driver.Stmt
	var wrapTx func(t driver.Tx) driver.Tx
	var wrapRows func(r driver.Rows) driver.Rows

	wrapDriver = func(d driver.Driver) driver.Driver {
		return databasesqlwrapper.WrapDriver(d, databasesqlwrapper.DriverInterceptor{
			Open: func(original databasesqlwrapper.Driver_Open) databasesqlwrapper.Driver_Open {
				return func(name string) (driver.Conn, error) {
					conn, err := original(name)
					if err != nil {
						return nil, err
					}
					return wrapConn(conn), nil
				}
			},
			OpenConnector: func(original databasesqlwrapper.Driver_OpenConnector) databasesqlwrapper.Driver_OpenConnector {
				return func(name string) (driver.Connector, error) {
					connector, err := original(name)
					if err != nil {
						return nil, err
					}
					return wrapConnector(connector), nil
				}
			},
		})
	}

	wrapConnector = func(c driver.Connector) driver.Connector {
		return databasesqlwrapper.WrapConnector(c, databasesqlwrapper.ConnectorInterceptor{
			Connect: func(original databasesqlwrapper.Connector_Connect) databasesqlwrapper.Connector_Connect {
				return func(ctx context.Context) (driver.Conn, error) {
					conn, err := original(ctx)
					if err != nil {
						return nil, err
					}
					return wrapConn(conn), nil
				}
			},
			Driver: func(original databasesqlwrapper.Connector_Driver) databasesqlwrapper.Connector_Driver {
				return func() driver.Driver {
					driver := original()
					return wrapDriver(driver)
				}
			},
		})
	}

	wrapConn = func(c driver.Conn) driver.Conn {
		return databasesqlwrapper.WrapConn(c, databasesqlwrapper.ConnInterceptor{
			Begin: func(original databasesqlwrapper.Conn_Begin) databasesqlwrapper.Conn_Begin {
				return func() (driver.Tx, error) {
					tx, err := original()
					if err != nil {
						return nil, err
					}
					return wrapTx(tx), nil
				}
			},
			BeginTx: func(original databasesqlwrapper.Conn_BeginTx) databasesqlwrapper.Conn_BeginTx {
				return func(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
					tx, err := original(ctx, opts)
					if err != nil {
						return nil, err
					}
					return wrapTx(tx), nil
				}
			},
			Prepare: func(original databasesqlwrapper.Conn_Prepare) databasesqlwrapper.Conn_Prepare {
				return func(query string) (driver.Stmt, error) {
					stmt, err := original(query)
					if err != nil {
						return nil, err
					}
					return wrapStmt(stmt), nil
				}
			},
			PrepareContext: func(original databasesqlwrapper.Conn_PrepareContext) databasesqlwrapper.Conn_PrepareContext {
				return func(ctx context.Context, query string) (driver.Stmt, error) {
					stmt, err := original(ctx, query)
					if err != nil {
						return nil, err
					}
					return wrapStmt(stmt), nil
				}
			},
			Exec: func(original databasesqlwrapper.Conn_Exec) databasesqlwrapper.Conn_Exec {
				return func(query string, args []driver.Value) (driver.Result, error) {
					result, err := original(query, args)
					if err != nil {
						return nil, err
					}
					return result, nil
				}
			},
			ExecContext: func(original databasesqlwrapper.Conn_ExecContext) databasesqlwrapper.Conn_ExecContext {
				return func(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
					result, err := original(ctx, query, args)
					if err != nil {
						return nil, err
					}
					return result, nil
				}
			},
			Query: func(original databasesqlwrapper.Conn_Query) databasesqlwrapper.Conn_Query {
				return func(query string, args []driver.Value) (driver.Rows, error) {
					rows, err := original(query, args)
					if err != nil {
						return nil, err
					}
					return wrapRows(rows), nil
				}
			},
			QueryContext: func(original databasesqlwrapper.Conn_QueryContext) databasesqlwrapper.Conn_QueryContext {
				return func(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
					rows, err := original(ctx, query, args)
					if err != nil {
						return nil, err
					}
					return wrapRows(rows), nil
				}
			},
		})
	}

	wrapStmt = func(s driver.Stmt) driver.Stmt {
		return databasesqlwrapper.WrapStmt(s, databasesqlwrapper.StmtInterceptor{
			Exec: func(original databasesqlwrapper.Stmt_Exec) databasesqlwrapper.Stmt_Exec {
				return func(args []driver.Value) (driver.Result, error) {
					result, err := original(args)
					if err != nil {
						return nil, err
					}
					return result, nil
				}
			},
			ExecContext: func(original databasesqlwrapper.Stmt_ExecContext) databasesqlwrapper.Stmt_ExecContext {
				return func(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
					result, err := original(ctx, args)
					if err != nil {
						return nil, err
					}
					return result, nil
				}
			},
			Query: func(original databasesqlwrapper.Stmt_Query) databasesqlwrapper.Stmt_Query {
				return func(args []driver.Value) (driver.Rows, error) {
					rows, err := original(args)
					if err != nil {
						return nil, err
					}
					return wrapRows(rows), nil
				}
			},
			QueryContext: func(original databasesqlwrapper.Stmt_QueryContext) databasesqlwrapper.Stmt_QueryContext {
				return func(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
					rows, err := original(ctx, args)
					if err != nil {
						return nil, err
					}
					return wrapRows(rows), nil
				}
			},
		})
	}

	wrapTx = func(t driver.Tx) driver.Tx {
		return databasesqlwrapper.WrapTx(t, databasesqlwrapper.TxInterceptor{
			Commit: func(original databasesqlwrapper.Tx_Commit) databasesqlwrapper.Tx_Commit {
				return func() error {
					return original()
				}
			},
			Rollback: func(original databasesqlwrapper.Tx_Rollback) databasesqlwrapper.Tx_Rollback {
				return func() error {
					return original()
				}
			},
		})
	}

	wrapRows = func(r driver.Rows) driver.Rows {
		return databasesqlwrapper.WrapRows(r, databasesqlwrapper.RowsInterceptor{
			Next: func(original databasesqlwrapper.Rows_Next) databasesqlwrapper.Rows_Next {
				return func(dest []driver.Value) error {
					return original(dest)
				}
			},
		})
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}
	originalDriver := db.Driver()
	err = db.Close()
	if err != nil {
		return nil, err
	}
	wrappedDriver := wrapDriver(originalDriver)

	if driverContext, ok := wrappedDriver.(driver.DriverContext); ok {
		connector, err := driverContext.OpenConnector(dsn)
		if err != nil {
			return nil, err
		}
		return sql.OpenDB(connector), nil
	}

	return sql.OpenDB(contextIgnoringConnector{
		driver: wrappedDriver,
		dsn:    dsn,
	}), nil
}

type contextIgnoringConnector struct {
	driver driver.Driver
	dsn    string
}

var _ driver.Connector = contextIgnoringConnector{}

func (c contextIgnoringConnector) Connect(_ context.Context) (driver.Conn, error) {
	return c.driver.Open(c.dsn)
}

func (c contextIgnoringConnector) Driver() driver.Driver {
	return c.driver
}
