package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/skygeario/skygear-server/cmd/authgear/server"
)

var cmdStart = &cobra.Command{
	Use:   "start [public|internal]...",
	Short: "Start public / internal server",
	Long:  `Start public / internal server`,
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &server.Controller{}

		serverTypes := args
		if len(serverTypes) == 0 {
			// Default to start both server
			serverTypes = []string{"public", "internal"}
		}
		for _, typ := range serverTypes {
			switch typ {
			case "public":
				ctrl.ServePublic = true
			case "internal":
				ctrl.ServeInternal = true
			default:
				log.Fatalf("unknown server type: %s", typ)
			}
		}

		ctrl.Start()
	},
}
