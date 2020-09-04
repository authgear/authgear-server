package model

import (
	"io/ioutil"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type App struct {
	ID      string
	Context *config.AppContext
}

func (a *App) LoadFile(path string) ([]byte, error) {
	file, err := a.Context.Fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return data, nil
}
