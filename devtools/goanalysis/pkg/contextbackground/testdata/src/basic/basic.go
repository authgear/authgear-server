package basic

import (
	"context"
)

func UseContext() {
	_ = context.Background() // want `Unvetted usage of context.Background is forbidden.`
	_ = context.TODO()       // want `Unvetted usage of context.TODO is forbidden.`
}
