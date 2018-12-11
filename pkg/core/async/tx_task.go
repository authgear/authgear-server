package async

import "github.com/skygeario/skygear-server/pkg/core/db"

type TxTask interface {
	Task
	WithTx() bool
}

func TxTaskToTask(t TxTask, txContext db.TxContext) Task {
	return TaskFunc(func(param interface{}) error {
		if t.WithTx() {
			// assume txContext != nil if apiHandler.WithTx() is true
			if err := txContext.BeginTx(); err != nil {
				panic(err)
			}

			defer func() {
				if txContext.HasTx() {
					txContext.RollbackTx()
				}
			}()
		}

		return t.Run(param)
	})
}
