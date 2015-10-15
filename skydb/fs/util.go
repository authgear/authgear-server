package fs

import (
	"os"

	"github.com/oursky/skygear/skydb"
)

func recordEventByPath(path string) skydb.RecordHookEvent {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return skydb.RecordCreated
	}

	return skydb.RecordUpdated
}
