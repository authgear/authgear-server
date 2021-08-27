package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/authgear/authgear-server/cmd/portal/internal"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

var migrateCookieDomainAppHostSuffix string

var cmdInternalMigrateCookieDomain = &cobra.Command{
	Use:   "migrate-cookie-domain",
	Short: "Set cookie domain for apps which are using custom domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		binder := getBinder()

		dbURL, err := binder.GetRequiredString(cmd, ArgDatabaseURL)
		if err != nil {
			return err
		}

		dbSchema, err := binder.GetRequiredString(cmd, ArgDatabaseSchema)
		if err != nil {
			return err
		}

		migrateCookieDomainAppHostSuffix, err = binder.GetRequiredString(cmd, ArgAppHostSuffix)
		if err != nil {
			return err
		}

		internal.MigrateResources(&internal.MigrateResourcesOptions{
			DatabaseURL:            dbURL,
			DatabaseSchema:         dbSchema,
			UpdateConfigSourceFunc: migrateCookieDomain,
			DryRun:                 &MigrateResourcesDryRun,
		})

		return nil
	},
}

func migrateCookieDomain(appID string, configSourceData map[string]string, dryRun bool) error {
	encodedData := configSourceData["authgear.yaml"]
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed decode authgear.yaml: %w", err)
	}

	if dryRun {
		log.Printf("Converting app (%s)", appID)
		log.Printf("Before updated:")
		log.Printf("\n%s\n", string(decoded))
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(decoded, &m)
	if err != nil {
		return fmt.Errorf("failed unmarshal yaml: %w", err)
	}

	httpConfig, ok := m["http"].(map[string]interface{})
	if !ok {
		return nil
	}

	publicOrigin, ok := httpConfig["public_origin"].(string)
	if !ok {
		return fmt.Errorf("cannot read public origin from authgear.yaml: %s", appID)
	}

	if strings.HasSuffix(publicOrigin, migrateCookieDomainAppHostSuffix) {
		// skip default domain
		log.Printf("skip default domain...")
		return nil
	}

	_, ok = httpConfig["cookie_domain"].(string)
	if ok {
		// skip the config that has cookie_domain
		log.Printf("skip config that has cookie_domain...")
		return nil
	}

	u, err := url.Parse(publicOrigin)
	if err != nil {
		return fmt.Errorf("failed to parse public origin: %w", err)
	}

	cookieDomain := httputil.CookieDomainWithoutPort(u.Host)
	httpConfig["cookie_domain"] = cookieDomain

	migrated, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed marshal yaml: %w", err)
	}

	if dryRun {
		log.Printf("After updated:")
		log.Printf("\n%s\n", string(migrated))
	}

	configSourceData["authgear.yaml"] = base64.StdEncoding.EncodeToString(migrated)
	return nil
}

func init() {
	binder := getBinder()
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateCookieDomain)
	binder.BindString(cmdInternalMigrateCookieDomain.Flags(), ArgAppHostSuffix)
}
