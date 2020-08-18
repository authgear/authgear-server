package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/authgear/server"
)

var cmdStart = &cobra.Command{
	Use:   "start [main|resolver|admin]...",
	Short: "Start specified servers",
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &server.Controller{}

		serverTypes := args
		if len(serverTypes) == 0 {
			// Default to start both main & resolver servers
			serverTypes = []string{"main", "resolver"}
		}
		for _, typ := range serverTypes {
			switch typ {
			case "main":
				ctrl.ServeMain = true
			case "resolver":
				ctrl.ServeResolver = true
			case "admin":
				ctrl.ServeAdmin = true
			default:
				log.Fatalf("unknown server type: %s", typ)
			}
		}

		ctrl.Start()
	},
}
