package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// PrintError prints err to stderr and returns err.
func PrintError(err error) error {
	fatalError := FatalError{Err: err}
	fmt.Fprintf(os.Stderr, "%v", fatalError.View())
	return err
}

func GetLicenseServerEndpoint(cmd *cobra.Command) string {
	if !LicenseServerEndpointOverridable {
		return LicenseServerEndpoint
	}
	endpoint, err := cmd.Flags().GetString("license-server-endpoint")
	if err != nil {
		panic(err)
	}
	if endpoint != "" {
		return endpoint
	}
	return LicenseServerEndpoint
}
