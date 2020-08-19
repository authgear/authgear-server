package main

import (
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/portal/server"
)

var cmdStart = &cobra.Command{
	Use:   "start",
	Short: "Start server",
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &server.Controller{}
		ctrl.Start()
	},
}
