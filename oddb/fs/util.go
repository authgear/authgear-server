package fs

import (
	"os"

	"github.com/oursky/ourd/oddb"
)

func recordEventByPath(path string) oddb.RecordHookEvent {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return oddb.RecordCreated
	}

	return oddb.RecordUpdated
}
