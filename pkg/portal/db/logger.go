package db

import "github.com/authgear/authgear-server/pkg/util/log"

type Logger struct{ *log.Logger }

func NewLogger(lf *log.Factory) Logger { return Logger{lf.New("database")} }
