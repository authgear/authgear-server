package main

import (
	"github.com/authgear/authgear-server/cmd/authgear/images/server"
	"github.com/spf13/cobra"
)

func init() {
	cmdImages.AddCommand(cmdImagesStart)
}

var cmdImages = &cobra.Command{
	Use:   "images",
	Short: " commands",
}

var cmdImagesStart = &cobra.Command{
	Use:   "start",
	Short: "Start images server",
	Run: func(cmd *cobra.Command, args []string) {
		ctrl := &server.Controller{}
		ctrl.Start()
	},
}
