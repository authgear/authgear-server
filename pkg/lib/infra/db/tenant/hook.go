package tenant

type TransactionHook interface {
	WillCommitTx() error
	DidCommitTx()
}
