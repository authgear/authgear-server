package main

import (
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/cmd/authgear/background"
)

var cmdBackground = &cobra.Command{
	Use:   "background",
	Short: "Start the background job runner",
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &background.Controller{}
		ctrl.Start()
	},
}
