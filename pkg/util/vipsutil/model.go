package vipsutil

import (
	"io"
)

type Input struct {
	Reader  io.Reader
	Options Options
}

type Output struct {
	FileExtension string
	Data          []byte
}
