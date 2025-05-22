package oteldatabasesql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv/v1.30.0"

	"github.com/authgear/authgear-server/pkg/util/databasesqlwrapper"
	"github.com/authgear/authgear-server/pkg/util/debug"
)

var meter = otel.Meter("github.com/authgear/authgear-server/pkg/util/otelutil/oteldatabasesql")

func mustFloat64Histogram(name string, options ...metric.Float64HistogramOption) metric.Float64Histogram {
	histogram, err := meter.Float64Histogram(name, options...)
	if err != nil {
		panic(err)
	}
	return histogram
}

func mustInt64ObservableGauge(name string, options ...metric.Int64ObservableGaugeOption) metric.Int64ObservableGauge {
	counter, err := meter.Int64ObservableGauge(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

// DBClientConnectionCount is https://opentelemetry.io/docs/specs/semconv/database/database-metrics/#metric-dbclientconnectioncount
var DBClientConnectionCount = mustInt64ObservableGauge(
	semconv.DBClientConnectionCountName,
	metric.WithDescription(semconv.DBClientConnectionCountDescription),
	metric.WithUnit(semconv.DBClientConnectionCountUnit),
)

// DBClientConnectionIdleMax is https://opentelemetry.io/docs/specs/semconv/database/database-metrics/#metric-dbclientconnectionidlemax
var DBClientConnectionIdleMax = mustInt64ObservableGauge(
	semconv.DBClientConnectionIdleMaxName,
	metric.WithDescription(semconv.DBClientConnectionIdleMaxDescription),
	metric.WithUnit(semconv.DBClientConnectionIdleMaxUnit),
)

// DBClientConnectionMax is https://opentelemetry.io/docs/specs/semconv/database/database-metrics/#metric-dbclientconnectionmax
var DBClientConnectionMax = mustInt64ObservableGauge(
	semconv.DBClientConnectionMaxName,
	metric.WithDescription(semconv.DBClientConnectionMaxDescription),
	metric.WithUnit(semconv.DBClientConnectionMaxUnit),
)

// DBClientConnectionCreateTimeHistogram is https://opentelemetry.io/docs/specs/semconv/database/database-metrics/#metric-dbclientconnectioncreate_time
var DBClientConnectionCreateTimeHistogram = mustFloat64Histogram(
	semconv.DBClientConnectionCreateTimeName,
	metric.WithDescription(semconv.DBClientConnectionCreateTimeDescription),
	metric.WithUnit(semconv.DBClientConnectionCreateTimeUnit),
	// The spec does not specify an explicit boundary.
	// We borrow the boundary from http request.
	metric.WithExplicitBucketBoundaries(
		0.005,
		0.01,
		0.025,
		0.05,
		0.075,
		0.1,
		0.25,
		0.5,
		0.75,
		1,
		2.5,
		5,
		7.5,
		10,
	),
)

// DBClientConnectionWaitTimeHistogram is https://opentelemetry.io/docs/specs/semconv/database/database-metrics/#metric-dbclientconnectionwait_time
var DBClientConnectionWaitTimeHistogram = mustFloat64Histogram(
	semconv.DBClientConnectionWaitTimeName,
	metric.WithDescription(semconv.DBClientConnectionWaitTimeDescription),
	metric.WithUnit(semconv.DBClientConnectionWaitTimeUnit),
	// The spec does not specify an explicit boundary.
	// We borrow the boundary from http request.
	metric.WithExplicitBucketBoundaries(
		0.005,
		0.01,
		0.025,
		0.05,
		0.075,
		0.1,
		0.25,
		0.5,
		0.75,
		1,
		2.5,
		5,
		7.5,
		10,
	),
)

// DBClientConnectionUseTimeHistogram is https://opentelemetry.io/docs/specs/semconv/database/database-metrics/#metric-dbclientconnectionuse_time
var DBClientConnectionUseTimeHistogram = mustFloat64Histogram(
	semconv.DBClientConnectionUseTimeName,
	metric.WithDescription(semconv.DBClientConnectionUseTimeDescription),
	metric.WithUnit(semconv.DBClientConnectionUseTimeUnit),
	// The spec does not specify an explicit boundary.
	// We borrow the boundary from http request.
	metric.WithExplicitBucketBoundaries(
		0.005,
		0.01,
		0.025,
		0.05,
		0.075,
		0.1,
		0.25,
		0.5,
		0.75,
		1,
		2.5,
		5,
		7.5,
		10,
	),
)

// DBClientOperationDurationHistogram is https://opentelemetry.io/docs/specs/semconv/database/database-metrics/#metric-dbclientoperationduration
var DBClientOperationDurationHistogram = mustFloat64Histogram(
	semconv.DBClientOperationDurationName,
	metric.WithDescription(semconv.DBClientOperationDurationDescription),
	metric.WithUnit(semconv.DBClientOperationDurationUnit),
	metric.WithExplicitBucketBoundaries(
		0.001,
		0.005,
		0.01,
		0.05,
		0.1,
		0.5,
		1,
		5,
		10,
	),
)

type Conn_ interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Close() error
}

var _ Conn_ = (*Conn)(nil)

// Conn is a wrapper around *sql.Conn.
// Conn is used to track connection use time.
type Conn struct {
	ctx         context.Context
	conn        *sql.Conn
	startTime   time.Time
	commonAttrs []attribute.KeyValue
	poolAttrs   []attribute.KeyValue
}

func (c *Conn) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return c.conn.BeginTx(ctx, opts)
}

func (c *Conn) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return c.conn.PrepareContext(ctx, query)
}

func (c *Conn) Close() error {
	defer func() {
		elapsed := time.Since(c.startTime)
		seconds := elapsed.Seconds()
		DBClientConnectionUseTimeHistogram.Record(
			c.ctx,
			seconds,
			metric.WithAttributes(c.commonAttrs...),
			metric.WithAttributes(c.poolAttrs...),
		)
	}()
	return c.conn.Close()
}

type ConnPool_ interface {
	Conn(ctx context.Context) (Conn_, error)
	Close() error
	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
}

var _ ConnPool_ = (*ConnPool)(nil)

// ConnPool is a wrapper around *sql.DB.
// ConnPool only supports Conn() which returns a wrapped *sql.Conn.
// ConnPool is used to track connection wait time.
type ConnPool struct {
	db          *sql.DB
	commonAttrs []attribute.KeyValue
	poolAttrs   []attribute.KeyValue
}

func (p *ConnPool) Conn(ctx context.Context) (Conn_, error) {

	// Intentionally not calling .UTC() to use monotonic clock.
	startTime := time.Now()
	defer func() {
		elapsed := time.Since(startTime)
		seconds := elapsed.Seconds()
		DBClientConnectionWaitTimeHistogram.Record(
			ctx,
			seconds,
			metric.WithAttributes(p.commonAttrs...),
			metric.WithAttributes(p.poolAttrs...),
		)
	}()

	Conn_blocking_ended := make(chan struct{}, 1)
	debug_connectionTimeout := once_get_AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS()
	if debug_connectionTimeout > 0 {
		go func() {
			select {
			case <-Conn_blocking_ended:
				break
			case <-time.After(debug_connectionTimeout):
				stack := debug.Stack()
				fmt.Fprintf(os.Stderr, "AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS: %v\n", base64.StdEncoding.EncodeToString(stack))
			}
		}()
	}

	sqlConn, err := p.db.Conn(ctx)
	if debug_connectionTimeout > 0 {
		Conn_blocking_ended <- struct{}{}
	}
	if err != nil {
		return nil, err
	}

	return &Conn{
		ctx: ctx,
		// Intentionally not calling .UTC() to use monotonic clock.
		startTime:   time.Now(),
		conn:        sqlConn,
		commonAttrs: p.commonAttrs,
		poolAttrs:   p.poolAttrs,
	}, nil
}

func (p *ConnPool) Close() error {
	return p.db.Close()
}

func (p *ConnPool) SetConnMaxIdleTime(d time.Duration) {
	p.db.SetConnMaxIdleTime(d)
}

func (p *ConnPool) SetConnMaxLifetime(d time.Duration) {
	p.db.SetConnMaxLifetime(d)
}

func (p *ConnPool) SetMaxIdleConns(n int) {
	p.db.SetMaxIdleConns(n)
}

func (p *ConnPool) SetMaxOpenConns(n int) {
	p.db.SetMaxOpenConns(n)
}

type OpenOptions struct {
	DriverName string
	DSN        string
	PoolName   string
	IdleMax    int
}

//nolint:gocognit
func Open(opts OpenOptions) (ConnPool_, error) {
	var commonAttrs []attribute.KeyValue
	var poolAttrs []attribute.KeyValue

	switch opts.DriverName {
	case "postgres":
		commonAttrs = append(commonAttrs, semconv.DBSystemNamePostgreSQL)
	case "sqlite3":
		commonAttrs = append(commonAttrs, semconv.DBSystemNameSqlite)
	default:
		panic(fmt.Errorf("unknown driver: %v", opts.DriverName))
	}
	poolAttrs = append(poolAttrs, semconv.DBClientConnectionPoolName(opts.PoolName))

	var wrapDriver func(d driver.Driver) driver.Driver
	var wrapConnector func(c driver.Connector) driver.Connector
	var wrapConn func(ctx context.Context, c driver.Conn) driver.Conn
	var wrapStmt func(ctx context.Context, s driver.Stmt, query string) driver.Stmt
	var wrapTx func(ctx context.Context, t driver.Tx) driver.Tx
	var wrapRows func(r driver.Rows) driver.Rows

	wrapDriver = func(d driver.Driver) driver.Driver {
		return databasesqlwrapper.WrapDriver(d, databasesqlwrapper.DriverInterceptor{
			Open: func(original databasesqlwrapper.Driver_Open) databasesqlwrapper.Driver_Open {
				return func(name string) (driver.Conn, error) {
					// We have no context here.
					ctx := context.TODO()
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientConnectionCreateTimeHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(poolAttrs...),
						)
					}()
					conn, err := original(name)
					if err != nil {
						return nil, err
					}
					return wrapConn(ctx, conn), nil
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
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientConnectionCreateTimeHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(poolAttrs...),
						)
					}()
					conn, err := original(ctx)
					if err != nil {
						return nil, err
					}
					return wrapConn(ctx, conn), nil
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

	wrapConn = func(ctx context.Context, c driver.Conn) driver.Conn {
		return databasesqlwrapper.WrapConn(c, databasesqlwrapper.ConnInterceptor{
			Close: func(original databasesqlwrapper.Conn_Close) databasesqlwrapper.Conn_Close {
				return func() error {
					return original()
				}
			},
			Begin: func(original databasesqlwrapper.Conn_Begin) databasesqlwrapper.Conn_Begin {
				return func() (driver.Tx, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("BEGIN")),
							metric.WithAttributes(semconv.DBQueryText("BEGIN")),
						)
					}()
					tx, err := original()
					if err != nil {
						return nil, err
					}
					return wrapTx(ctx, tx), nil
				}
			},
			BeginTx: func(original databasesqlwrapper.Conn_BeginTx) databasesqlwrapper.Conn_BeginTx {
				return func(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("BEGIN")),
							metric.WithAttributes(semconv.DBQueryText("BEGIN")),
						)
					}()
					tx, err := original(ctx, opts)
					if err != nil {
						return nil, err
					}
					return wrapTx(ctx, tx), nil
				}
			},
			Prepare: func(original databasesqlwrapper.Conn_Prepare) databasesqlwrapper.Conn_Prepare {
				return func(query string) (driver.Stmt, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("PREPARE")),
							metric.WithAttributes(semconv.DBQueryText(query)),
						)
					}()
					stmt, err := original(query)
					if err != nil {
						return nil, err
					}
					return wrapStmt(ctx, stmt, query), nil
				}
			},
			PrepareContext: func(original databasesqlwrapper.Conn_PrepareContext) databasesqlwrapper.Conn_PrepareContext {
				return func(ctx context.Context, query string) (driver.Stmt, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("PREPARE")),
							metric.WithAttributes(semconv.DBQueryText(query)),
						)
					}()
					stmt, err := original(ctx, query)
					if err != nil {
						return nil, err
					}
					return wrapStmt(ctx, stmt, query), nil
				}
			},
			Exec: func(original databasesqlwrapper.Conn_Exec) databasesqlwrapper.Conn_Exec {
				return func(query string, args []driver.Value) (driver.Result, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("Exec")),
							metric.WithAttributes(semconv.DBQueryText(query)),
						)
					}()
					result, err := original(query, args)
					if err != nil {
						return nil, err
					}
					return result, nil
				}
			},
			ExecContext: func(original databasesqlwrapper.Conn_ExecContext) databasesqlwrapper.Conn_ExecContext {
				return func(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("Exec")),
							metric.WithAttributes(semconv.DBQueryText(query)),
						)
					}()
					result, err := original(ctx, query, args)
					if err != nil {
						return nil, err
					}
					return result, nil
				}
			},
			Query: func(original databasesqlwrapper.Conn_Query) databasesqlwrapper.Conn_Query {
				return func(query string, args []driver.Value) (driver.Rows, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("Query")),
							metric.WithAttributes(semconv.DBQueryText(query)),
						)
					}()
					rows, err := original(query, args)
					if err != nil {
						return nil, err
					}
					return wrapRows(rows), nil
				}
			},
			QueryContext: func(original databasesqlwrapper.Conn_QueryContext) databasesqlwrapper.Conn_QueryContext {
				return func(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("Query")),
							metric.WithAttributes(semconv.DBQueryText(query)),
						)
					}()
					rows, err := original(ctx, query, args)
					if err != nil {
						return nil, err
					}
					return wrapRows(rows), nil
				}
			},
		})
	}

	wrapStmt = func(ctx context.Context, s driver.Stmt, query string) driver.Stmt {
		return databasesqlwrapper.WrapStmt(s, databasesqlwrapper.StmtInterceptor{
			Exec: func(original databasesqlwrapper.Stmt_Exec) databasesqlwrapper.Stmt_Exec {
				return func(args []driver.Value) (driver.Result, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("Exec")),
							metric.WithAttributes(semconv.DBQueryText(query)),
						)
					}()
					result, err := original(args)
					if err != nil {
						return nil, err
					}
					return result, nil
				}
			},
			ExecContext: func(original databasesqlwrapper.Stmt_ExecContext) databasesqlwrapper.Stmt_ExecContext {
				return func(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("Exec")),
							metric.WithAttributes(semconv.DBQueryText(query)),
						)
					}()
					result, err := original(ctx, args)
					if err != nil {
						return nil, err
					}
					return result, nil
				}
			},
			Query: func(original databasesqlwrapper.Stmt_Query) databasesqlwrapper.Stmt_Query {
				return func(args []driver.Value) (driver.Rows, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("Query")),
							metric.WithAttributes(semconv.DBQueryText(query)),
						)
					}()
					rows, err := original(args)
					if err != nil {
						return nil, err
					}
					return wrapRows(rows), nil
				}
			},
			QueryContext: func(original databasesqlwrapper.Stmt_QueryContext) databasesqlwrapper.Stmt_QueryContext {
				return func(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("Query")),
							metric.WithAttributes(semconv.DBQueryText(query)),
						)
					}()
					rows, err := original(ctx, args)
					if err != nil {
						return nil, err
					}
					return wrapRows(rows), nil
				}
			},
		})
	}

	wrapTx = func(ctx context.Context, t driver.Tx) driver.Tx {
		return databasesqlwrapper.WrapTx(t, databasesqlwrapper.TxInterceptor{
			Commit: func(original databasesqlwrapper.Tx_Commit) databasesqlwrapper.Tx_Commit {
				return func() error {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("COMMIT")),
							metric.WithAttributes(semconv.DBQueryText("COMMIT")),
						)
					}()
					return original()
				}
			},
			Rollback: func(original databasesqlwrapper.Tx_Rollback) databasesqlwrapper.Tx_Rollback {
				return func() error {
					// Intentionally not calling .UTC() to use monotonic clock.
					startTime := time.Now()
					defer func() {
						elapsed := time.Since(startTime)
						seconds := elapsed.Seconds()
						DBClientOperationDurationHistogram.Record(
							ctx,
							seconds,
							metric.WithAttributes(commonAttrs...),
							metric.WithAttributes(semconv.DBOperationName("ROLLBACK")),
							metric.WithAttributes(semconv.DBQueryText("ROLLBACK")),
						)
					}()
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

	db, err := sql.Open(opts.DriverName, opts.DSN)
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
		connector, err := driverContext.OpenConnector(opts.DSN)
		if err != nil {
			return nil, err
		}
		db = sql.OpenDB(connector)
	} else {
		db = sql.OpenDB(contextIgnoringConnector{
			driver: wrappedDriver,
			dsn:    opts.DSN,
		})
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		stats := db.Stats()

		o.ObserveInt64(
			DBClientConnectionCount,
			int64(stats.InUse),
			metric.WithAttributes(commonAttrs...),
			metric.WithAttributes(poolAttrs...),
			metric.WithAttributes(semconv.DBClientConnectionStateUsed),
		)
		o.ObserveInt64(
			DBClientConnectionCount,
			int64(stats.Idle),
			metric.WithAttributes(commonAttrs...),
			metric.WithAttributes(poolAttrs...),
			metric.WithAttributes(semconv.DBClientConnectionStateIdle),
		)
		o.ObserveInt64(
			DBClientConnectionIdleMax,
			int64(opts.IdleMax),
			metric.WithAttributes(commonAttrs...),
			metric.WithAttributes(poolAttrs...),
		)
		o.ObserveInt64(
			DBClientConnectionMax,
			int64(stats.MaxOpenConnections),
			metric.WithAttributes(commonAttrs...),
			metric.WithAttributes(poolAttrs...),
		)

		return nil
	}, DBClientConnectionCount, DBClientConnectionIdleMax, DBClientConnectionMax)
	if err != nil {
		return nil, err
	}

	return &ConnPool{
		db:          db,
		commonAttrs: commonAttrs,
		poolAttrs:   poolAttrs,
	}, nil
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

func get_AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS() time.Duration {
	s := os.Getenv("AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS")
	if s == "" {
		return 0
	}
	millis, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	if millis <= 0 {
		return 0
	}
	return time.Duration(millis) * time.Millisecond
}

var once_get_AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS = sync.OnceValue(get_AUTHGEARDEBUG_DATABASE_CONNECTION_WAIT_TIME_TIMEOUT_MILLISECONDS)
