package db

import (
	"errors"
	"time"

	"github.com/lib/pq"
)

type PQListener struct {
	DatabaseURL string
	Listener    *pq.Listener
}

func (p *PQListener) Listen(channels []string, done <-chan struct{}, onChange func(channel string, extra string), onError func(error)) {
	if p.Listener != nil {
		onError(errors.New("db: PQListener is started already"))
		return
	}
	listener := pq.NewListener(
		p.DatabaseURL,
		10*time.Second,
		1*time.Minute,
		nil,
	)
	for _, channel := range channels {
		err := listener.Listen(channel)
		if err != nil {
			onError(err)
			return
		}
	}

	p.Listener = listener
	for {
		select {
		case <-done:
			p.Listener = nil
			return
		case e := <-listener.Notify:
			if e != nil {
				onChange(e.Channel, e.Extra)
			}
		case <-time.After(time.Minute):
			err := listener.Ping()
			if err != nil {
				onError(err)
			}
		}
	}
}
