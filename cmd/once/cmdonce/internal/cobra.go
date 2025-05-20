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

func GetDockerImage(cmd *cobra.Command) string {
	image, err := cmd.Flags().GetString("image")
	if err != nil {
		panic(err)
	}

	if image != "" {
		return image
	}

	return fmt.Sprintf("%v:%v", DefaultDockerName_NoTag, Version)
}

func FlagsGetBool(cmd *cobra.Command, name string) bool {
	b, err := cmd.Flags().GetBool(name)
	if err != nil {
		panic(err)
	}
	return b
}
