package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/phires/go-guerrilla"
	"github.com/phires/go-guerrilla/backends"
	guerrillalog "github.com/phires/go-guerrilla/log"

	"github.com/authgear/authgear-server/pkg/util/debug"
)

func main() {
	debug.TrapSIGQUIT()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	cfg := &guerrilla.AppConfig{
		LogFile: guerrillalog.OutputStdout.String(),
		AllowedHosts: []string{
			"*",
		},
		// The defaults of go-guerrilla, plus MailLog.
		BackendConfig: backends.BackendConfig{
			"log_received_mails": true,
			"save_workers_size":  1,
			"save_process":       "HeadersParser|Header|Debugger|MailLog",
		},
		Servers: []guerrilla.ServerConfig{
			{
				IsEnabled:       true,
				ListenInterface: "127.0.0.1:2525",
			},
		},
	}

	d := guerrilla.Daemon{Config: cfg}
	d.AddProcessor("MailLog", MailLogProcessor)

	err := d.Start()
	if err != nil {
		log.Fatalf("Failed to start smtp: %v", err)
	}
	defer d.Shutdown()

	<-ctx.Done()
	log.Println("Shutting down...")
}
