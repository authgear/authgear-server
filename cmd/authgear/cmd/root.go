package cmd

import (
	"github.com/spf13/cobra"

	"github.com/authgear/authgear-server/pkg/version"
)

var Root = &cobra.Command{
	Use:     "authgear",
	Version: version.Version,
}
