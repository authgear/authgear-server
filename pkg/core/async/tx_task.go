package async

import (
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type TxTask interface {
	Task
	WithTx() bool
}

func TxTaskToTask(t TxTask, txContext db.TxContext) Task {
	return TaskFunc(func(param interface{}) (err error) {
		if t.WithTx() {
			err = db.WithTx(txContext, func() error { return t.Run(param) })
		} else {
			err = t.Run(param)
		}
		return
	})
}
