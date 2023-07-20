package db

type TransactionHook interface {
	WillCommitTx() error
	DidCommitTx()
}
