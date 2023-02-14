package db

type TransactionHook interface {
	WillCommitTx() error
	DidCommitTx()
	WillRollbackTx() error
	DidRollbackTx()
}
