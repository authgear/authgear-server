package principal

import "github.com/skygeario/skygear-server/pkg/server/uuid"

type Principal struct {
	ID       string
	Provider string
	UserID   string
}

func New() Principal {
	return Principal{
		ID: uuid.New(),
	}
}
