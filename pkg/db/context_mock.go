package db

import "context"

type MockContext struct {
	context.Context
	DidBegin, DidCommit, DidRollback bool
}

var _ Context = &MockContext{}

func NewMockTxContext() *MockContext {
	return &MockContext{}
}

func (c *MockContext) UseHook(h TransactionHook) {
}

func (c *MockContext) DB() (ExtContext, error) {
	return nil, nil
}

func (c *MockContext) HasTx() bool {
	return c.DidBegin == true && c.DidCommit == false && c.DidRollback == false
}

func (c *MockContext) beginTx() error {
	c.DidBegin = true
	return nil
}

func (c *MockContext) commitTx() error {
	c.DidCommit = true
	return nil
}

func (c *MockContext) rollbackTx() error {
	c.DidRollback = true
	return nil
}
