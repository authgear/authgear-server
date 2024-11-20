package databasesqlutil

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
)

func Register(name string, driver driver.Driver) {
	sql.Register(name, &Driver{
		Driver: driver,
	})
}

// Driver is a driver that caches driver.Stmt per driver.Conn.
type Driver struct {
	driver.Driver
}

var _ driver.Driver = (*Driver)(nil)

func (d *Driver) Open(name string) (driver.Conn, error) {
	driverConn, err := d.Driver.Open(name)
	if err != nil {
		return nil, err
	}

	return &conn{
		SuperConn:   toSuperConn(driverConn),
		queryToStmt: make(map[string]*stmt),
		stmtToQuery: make(map[*stmt]string),
	}, nil
}

type SuperConn interface {
	driver.Conn
	driver.ConnBeginTx
	driver.ConnPrepareContext
	driver.QueryerContext
	driver.ExecerContext
	driver.Validator
	// lib/pq@1.10.9 was released on 2023-04-26
	// While the support for NamedValueChecker was merged on 2023-05-04
	// See https://github.com/lib/pq/pull/1125
	//driver.NamedValueChecker
}

func toSuperConn(conn driver.Conn) SuperConn {
	if _, ok := conn.(driver.ConnBeginTx); !ok {
		panic(fmt.Errorf("driver.Conn %T does not implement ConnBeginTx", conn))
	}
	if _, ok := conn.(driver.ConnPrepareContext); !ok {
		panic(fmt.Errorf("driver.Conn %T does not implement ConnPrepareContext", conn))
	}
	if _, ok := conn.(driver.QueryerContext); !ok {
		panic(fmt.Errorf("driver.Conn %T does not implement QueryerContext", conn))
	}
	if _, ok := conn.(driver.ExecerContext); !ok {
		panic(fmt.Errorf("driver.Conn %T does not implement ExecerContext", conn))
	}
	if _, ok := conn.(driver.Validator); !ok {
		panic(fmt.Errorf("driver.Conn %T does not implement Validator", conn))
	}
	superConn, ok := conn.(SuperConn)
	if !ok {
		panic(fmt.Errorf("driver.Conn %T does not implement all Conn interfaces", conn))
	}
	return superConn
}

// conn is a driver.Conn with a prepared statement cache.
// No lock is imposed because driver.Conn is supposed to be used by a single goroutine.
type conn struct {
	SuperConn
	queryToStmt map[string]*stmt
	stmtToQuery map[*stmt]string
}

var _ driver.Conn = (*conn)(nil)
var _ driver.ConnBeginTx = (*conn)(nil)
var _ driver.ConnPrepareContext = (*conn)(nil)
var _ driver.QueryerContext = (*conn)(nil)
var _ driver.ExecerContext = (*conn)(nil)
var _ driver.Validator = (*conn)(nil)

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	s, ok := c.queryToStmt[query]
	if ok {
		return s, nil
	}

	driverStmt, err := c.SuperConn.Prepare(query)
	if err != nil {
		return nil, err
	}

	s = &stmt{
		SuperStmt: toSuperStmt(driverStmt),
		owner:     c,
	}
	c.queryToStmt[query] = s
	c.stmtToQuery[s] = query

	return s, nil
}

func (c *conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	s, ok := c.queryToStmt[query]
	if ok {
		return s, nil
	}

	driverStmt, err := c.SuperConn.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}

	s = &stmt{
		SuperStmt: toSuperStmt(driverStmt),
		owner:     c,
	}
	c.queryToStmt[query] = s
	c.stmtToQuery[s] = query

	return s, nil
}

func (c *conn) Close() error {
	var err error

	for s := range c.stmtToQuery {
		stmtErr := s.Close()
		if stmtErr != nil {
			err = errors.Join(err, stmtErr)
		}
	}

	closeErr := c.SuperConn.Close()
	if closeErr != nil {
		err = errors.Join(err, closeErr)
	}

	return err
}

func (c *conn) closeStmt(s *stmt) error {
	query, ok := c.stmtToQuery[s]
	if ok {
		delete(c.stmtToQuery, s)
		delete(c.queryToStmt, query)
	}

	return s.SuperStmt.Close()
}

type SuperStmt interface {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
	// lib/pq.stmt does not implement ColumnConverter.
	//driver.ColumnConverter
	// lib/pq.stmt does not implement NamedValueChecker.
	//driver.NamedValueChecker
}

func toSuperStmt(s driver.Stmt) SuperStmt {
	if _, ok := s.(driver.StmtExecContext); !ok {
		panic(fmt.Errorf("driver.Stmt %T does not implement StmtExecContext", s))
	}
	if _, ok := s.(driver.StmtExecContext); !ok {
		panic(fmt.Errorf("driver.Stmt %T does not implement StmtExecContext", s))
	}
	superStmt, ok := s.(SuperStmt)
	if !ok {
		panic(fmt.Errorf("driver.Stmt %T does not implement all Stmt interfaces", s))
	}
	return superStmt
}

// stmt is a driver.Stmt with a reference to its driver.Conn.
// When close, it asks its driver.Conn to evict itself from the cache.
type stmt struct {
	SuperStmt
	owner *conn
}

var _ driver.Stmt = (*stmt)(nil)
var _ driver.StmtExecContext = (*stmt)(nil)
var _ driver.StmtQueryContext = (*stmt)(nil)

func (s *stmt) Close() error {
	return s.owner.closeStmt(s)
}
