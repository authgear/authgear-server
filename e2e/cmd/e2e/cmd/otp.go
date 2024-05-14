package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	e2e "github.com/authgear/authgear-server/e2e/cmd/e2e/pkg"
)

func init() {
	binder := GetBinder()

	Root.AddCommand(cmdInternalE2EGetLinkOTPCode)
	binder.BindString(cmdInternalE2EGetLinkOTPCode.PersistentFlags(), ArgAppID)
}

var cmdInternalE2EGetLinkOTPCode = &cobra.Command{
	Use:   "link-otp-code [claim-name] [claim-value]",
	Short: "Get Link OTP Code by claim",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := GetBinder()

		appID := binder.GetString(cmd, ArgAppID)
		claimName := args[0]
		claimValue := args[1]

		instance := e2e.End2End{
			Context: cmd.Context(),
		}

		otpCode, err := instance.GetLinkOTPCode(appID, claimName, claimValue)
		if err != nil {
			return err
		}

		fmt.Print(otpCode)

		return nil
	},
}
