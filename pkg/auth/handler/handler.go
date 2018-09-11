package handler

import (
	"time"

	"github.com/skygeario/skygear-server/pkg/server/uuid"
)

var (
	uuidNew = uuid.New
	timeNow = func() time.Time { return time.Now().UTC() }
)
