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

func (c *MockContext) WithTx(do func() error) error {
	return do()
}

func (c *MockContext) ReadOnly(do func() error) error {
	return do()
}
