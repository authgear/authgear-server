package accountdeletion

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/db/globaldb"
)

type Store struct {
	Handle *globaldb.Handle
}
