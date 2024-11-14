package cmdinternal

import (
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
)

var cmdInternalDomain = &cobra.Command{
	Use:   "domain",
	Short: "Domain commands.",
}

var cmdInternalDomainCreateDefault = &cobra.Command{
	Use:   "create-default",
	Short: "Create default domain for all apps. It does NOT create duplicate records.",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := portalcmd.GetBinder()

		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		suffix, err := binder.GetRequiredString(cmd, portalcmd.ArgDefaultDomainSuffix)
		if err != nil {
			return err
		}

		err = validateDomainSuffix(suffix)
		if err != nil {
			return fmt.Errorf("%s: %w", portalcmd.ArgDefaultDomainSuffix.ArgumentName, err)
		}

		return internal.CreateDefaultDomain(cmd.Context(), internal.CreateDefaultDomainOptions{
			DatabaseURL:         dbURL,
			DatabaseSchema:      dbSchema,
			DefaultDomainSuffix: suffix,
		})
	},
}

var cmdInternalDomainCreateCustom = &cobra.Command{
	Use:   "create-custom <app-id>",
	Short: "Create custom domain for the app. It does NOT create duplicate records.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := portalcmd.GetBinder()

		dbURL, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, portalcmd.ArgDatabaseSchema)
		if err != nil {
			return err
		}

		domain, err := binder.GetRequiredString(cmd, portalcmd.ArgDomain)
		if err != nil {
			return err
		}

		apexDomain, err := binder.GetRequiredString(cmd, portalcmd.ArgApexDomain)
		if err != nil {
			return err
		}

		err = validateDomain(domain)
		if err != nil {
			return fmt.Errorf("%s: %w", portalcmd.ArgDomain.ArgumentName, err)
		}

		err = validateDomain(apexDomain)
		if err != nil {
			return fmt.Errorf("%s: %w", portalcmd.ArgApexDomain.ArgumentName, err)
		}

		appID := args[0]

		// FIXME(domain): Create the ingress resource?
		return internal.CreateCustomDomain(cmd.Context(), internal.CreateCustomDomainOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
			AppID:          appID,
			Domain:         domain,
			ApexDomain:     apexDomain,
		})
	},
}

func validateDomainSuffix(suffix string) error {
	if !strings.HasPrefix(suffix, ".") {
		return fmt.Errorf("domain suffix must start with a `.`")
	}

	// Trim the dot.
	domain := suffix[1:]

	host, _, err := net.SplitHostPort(domain)
	if err == nil {
		return fmt.Errorf("domain suffix must not contain a port")
	}

	ip := net.ParseIP(host)
	if ip != nil {
		return fmt.Errorf("domain suffix must not be an IP")
	}

	return nil
}

func validateDomain(domain string) error {
	host, _, err := net.SplitHostPort(domain)
	if err == nil {
		return fmt.Errorf("domain must not contain a port")
	}

	ip := net.ParseIP(host)
	if ip != nil {
		return fmt.Errorf("domain must not be an IP")
	}

	return nil
}
