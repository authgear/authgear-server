package db

// MockTxContext implements and record db.TxContext methods
// FIXME: It assumes that the TxContext does not get reuse
type MockTxContext struct {
	DidBegin, DidCommit, DidRollback bool
}

var _ TxContext = &MockTxContext{}

func NewMockTxContext() *MockTxContext {
	return &MockTxContext{}
}

func (c *MockTxContext) UseHook(h TransactionHook) {
}

func (c *MockTxContext) HasTx() bool {
	return c.DidBegin == true && c.DidCommit == false && c.DidRollback == false
}

func (c *MockTxContext) beginTx() error {
	c.DidBegin = true
	return nil
}

func (c *MockTxContext) commitTx() error {
	c.DidCommit = true
	return nil
}

func (c *MockTxContext) rollbackTx() error {
	c.DidRollback = true
	return nil
}
