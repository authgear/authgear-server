package cmdinternal

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"regexp"

	"github.com/spf13/cobra"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
	"github.com/authgear/authgear-server/pkg/lib/theme"
	"github.com/authgear/authgear-server/pkg/util/filepathutil"
)

var cmdInternalMigrateSetDefaultLogoHeight = &cobra.Command{
	Use:   "migrate-set-default-logo-height",
	Short: "Set default logo height for apps which have logos and do not have logo height customized", // more context in DEV-2126
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

		internal.MigrateResources(&internal.MigrateResourcesOptions{
			DatabaseURL:            dbURL,
			DatabaseSchema:         dbSchema,
			UpdateConfigSourceFunc: migrateSetDefaultLogoHeight,
			DryRun:                 &MigrateResourcesDryRun,
		})

		return nil
	},
}

func migrateSetDefaultLogoHeight(appID string, configSourceData map[string]string, dryRun bool) error {
	cfg, err := parseLogoHeightConfigSource(configSourceData)
	if err != nil {
		return err
	}

	hasLightLogo := cfg.LightLogo != nil
	hasDarkLogo := cfg.DarkLogo != nil
	hasLightThemeCSS := cfg.LightThemeCSS != nil
	hasDarkThemeCSS := cfg.DarkThemeCSS != nil

	if !hasLightLogo && !hasDarkLogo {
		log.Printf("Skipping app (%s) because it does not have logo set", appID)
		return nil
	}

	if dryRun {
		log.Printf("Converting app (%s)", appID)
	}

	var lightThemeCSSMigrated []byte
	var darkThemeCSSMigrated []byte
	var lightThemeCSSAlreadySet bool
	var darkThemeCSSAlreadySet bool
	if hasLightLogo {
		if hasLightThemeCSS {
			lightThemeCSSMigrated, lightThemeCSSAlreadySet, err = handleExistingCSS(cfg.LightThemeCSS)
			if lightThemeCSSAlreadySet {
				log.Printf("Skipping light theme css of app (%s) because it already has logo height set", appID)
			}
			if err != nil {
				return err
			}
		}
	}

	if hasDarkLogo {
		if hasDarkThemeCSS {
			darkThemeCSSMigrated, darkThemeCSSAlreadySet, err = handleExistingCSS(cfg.DarkThemeCSS)
			if darkThemeCSSAlreadySet {
				log.Printf("Skipping dark theme css of app (%s) because it already has logo height set", appID)
			}
			if err != nil {
				return err
			}
		}

	}

	if !lightThemeCSSAlreadySet {
		configSourceData[cfg.LightThemeCSS.OriginalPath] = base64.StdEncoding.EncodeToString(lightThemeCSSMigrated)
		if dryRun {
			log.Printf("Before light-theme.css updated:")
			log.Printf("%s:\n%s\n", cfg.LightThemeCSS.OriginalPath, string(cfg.LightThemeCSS.DecodedData))
			log.Printf("After light-theme.css updated:")
			log.Printf("%s:\n%s\n", cfg.LightThemeCSS.OriginalPath, string(lightThemeCSSMigrated))
		}
	}

	if !darkThemeCSSAlreadySet {
		configSourceData[cfg.DarkThemeCSS.OriginalPath] = base64.StdEncoding.EncodeToString(darkThemeCSSMigrated)
		if dryRun {
			log.Printf("Before dark-theme.css updated:")
			log.Printf("%s:\n%s\n", cfg.DarkThemeCSS.OriginalPath, string(cfg.DarkThemeCSS.DecodedData))
			log.Printf("After dark-theme.css updated:")
			log.Printf("%s:\n%s\n", cfg.DarkThemeCSS.OriginalPath, string(darkThemeCSSMigrated))
		}
	}

	return nil
}

func handleExistingCSS(cssResource *ResourceConfigDecoded) (migratedCSS []byte, alreadySet bool, err error) {
	r := bytes.NewReader(cssResource.DecodedData)
	migratedCSS, alreadySet, err = theme.MigrateSetDefaultLogoHeight(r)
	richErr := fmt.Errorf("failed to migrate %v: %w", cssResource.OriginalPath, err)
	return migratedCSS, alreadySet, richErr
}

type ResourceConfig struct {
	OriginalPath string
	EscapedPath  string
	EncodedData  string
}

type ResourceConfigDecoded struct {
	OriginalPath string
	EscapedPath  string
	EncodedData  string
	DecodedData  []byte
}
type MigrateLogoHeightConfig struct {
	LightLogo     *ResourceConfig
	DarkLogo      *ResourceConfig
	LightThemeCSS *ResourceConfigDecoded
	DarkThemeCSS  *ResourceConfigDecoded
}

var LightLogoPathRegex = `^static/([a-zA-Z-]+)/app_logo\.([a-zA-Z-]+)$`
var DarkLogoPathRegex = `^static/([a-zA-Z-]+)/app_logo_dark\.([a-zA-Z-]+)$`
var LightThemeCSSPath = "static/authgear-authflowv2-light-theme.css"
var DarkThemeCSSPath = "static/authgear-authflowv2-dark-theme.css"

func parseLogoHeightConfigSource(configSourceData map[string]string) (*MigrateLogoHeightConfig, error) {
	out := &MigrateLogoHeightConfig{}

	for k, v := range configSourceData {
		p, err := filepathutil.UnescapePath(k)
		if err != nil {
			return nil, err
		}
		if ok, _ := regexp.MatchString(LightLogoPathRegex, p); ok {
			out.LightLogo = newResourceConfig(p, k, v)
		}
		if ok, _ := regexp.MatchString(DarkLogoPathRegex, p); ok {
			out.DarkLogo = newResourceConfig(p, k, v)
		}
		if p == LightThemeCSSPath {
			rcd, err := newResourceConfigDecoded(p, k, v)
			if err != nil {
				return nil, err
			}
			out.LightThemeCSS = rcd
		}
		if p == DarkThemeCSSPath {
			rcd, err := newResourceConfigDecoded(p, k, v)
			if err != nil {
				return nil, err
			}
			out.DarkThemeCSS = rcd
		}
	}

	return out, nil
}

func newResourceConfig(originalPath string, escapedPath string, encoded string) *ResourceConfig {
	return &ResourceConfig{
		OriginalPath: originalPath,
		EscapedPath:  escapedPath,
		EncodedData:  encoded,
	}
}

func newResourceConfigDecoded(originalPath string, escapedPath string, encoded string) (*ResourceConfigDecoded, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %v: %w", originalPath, err)
	}
	return &ResourceConfigDecoded{
		OriginalPath: originalPath,
		EscapedPath:  escapedPath,
		EncodedData:  encoded,
		DecodedData:  decoded,
	}, nil
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateSetDefaultLogoHeight)
}
