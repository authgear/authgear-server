package accountdeletion

import (
	"context"
)

type Runnable struct {
	Store *Store
}

func (r *Runnable) Run(ctx context.Context) error {
	return nil
}
