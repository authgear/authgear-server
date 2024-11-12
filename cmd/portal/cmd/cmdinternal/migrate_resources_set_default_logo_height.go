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

const lightThemeSelector = ":root"
const darkThemeSelector = ":root.dark"

var defaultLogoHeightDeclaration = theme.Declaration{
	Property: "--brand-logo__height",
	Value:    "40px",
}

const LightThemeCSSPath = "static/authgear-authflowv2-light-theme.css"
const DarkThemeCSSPath = "static/authgear-authflowv2-dark-theme.css"

var lightLogoPathRegex = regexp.MustCompile(`^static/([a-zA-Z-]+)/app_logo\.(png|jpe|jpeg|jpg|gif)$`)
var darkLogoPathRegex = regexp.MustCompile(`^static/([a-zA-Z-]+)/app_logo_dark\.(png|jpe|jpeg|jpg|gif)$`)
var LightThemeCSSPathRegex = regexp.MustCompile(fmt.Sprintf(`^%v$`, LightThemeCSSPath))
var DarkThemeCSSPathRegex = regexp.MustCompile(fmt.Sprintf(`^%v$`, DarkThemeCSSPath))

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

		internal.MigrateResources(cmd.Context(), &internal.MigrateResourcesOptions{
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

	if dryRun {
		log.Printf("\n================== Converting app (%s) ==================", appID)
	}

	// invariant check
	switch {
	case !hasLightLogo && !hasDarkLogo:
		log.Printf("Skipping app (%s) because it does not have logo set", appID)
		return nil
	}

	// handle light theme
	switch {
	case hasLightLogo && hasLightThemeCSS:
		err = handleExistingLightThemeCSS(appID, configSourceData, cfg.LightThemeCSS, dryRun)
		if err != nil {
			return err
		}
	case hasLightLogo && !hasLightThemeCSS:
		err = handleMissingLightThemeCSS(appID, configSourceData, dryRun)
		if err != nil {
			return err
		}
	case !hasLightLogo:
		log.Printf("Skipping light theme css creation of app (%s) because it does not have light theme logo set", appID)
	}

	// handle dark theme
	switch {
	case hasDarkLogo && hasDarkThemeCSS:
		err = handleExistingDarkThemeCSS(appID, configSourceData, cfg.DarkThemeCSS, dryRun)
		if err != nil {
			return err
		}
	case hasDarkLogo && !hasDarkThemeCSS:
		err = handleMissingDarkThemeCSS(appID, configSourceData, dryRun)
		if err != nil {
			return err
		}
	case !hasDarkLogo:
		log.Printf("Skipping dark theme css creation of app (%s) because it does not have dark theme logo set", appID)
	}

	return nil
}

func handleExistingLightThemeCSS(appID string, configSourceData map[string]string, cssResource *ResourceConfigDecoded, dryRun bool) (err error) {
	migratedCSS, added, err := theme.AddDeclarationInSelectorIfNotPresentAlready(string(cssResource.DecodedData), lightThemeSelector, defaultLogoHeightDeclaration)
	if err != nil {
		return fmt.Errorf("failed to migrate %v: %w", cssResource.OriginalPath, err)
	}
	if !added {
		log.Printf("Skipping %s of app (%s) because it already has logo height set", cssResource.OriginalPath, appID)
		return nil
	}
	configSourceData[cssResource.EscapedPath] = base64.StdEncoding.EncodeToString([]byte(migratedCSS))
	if dryRun {
		log.Printf("Before %s updated:", cssResource.OriginalPath)
		log.Printf("\n%s\n", string(cssResource.DecodedData))
		log.Printf("After %s updated:", cssResource.OriginalPath)
		log.Printf("\n%s\n", string(migratedCSS))
	}

	return nil
}

func handleExistingDarkThemeCSS(appID string, configSourceData map[string]string, cssResource *ResourceConfigDecoded, dryRun bool) (err error) {
	migratedCSS, added, err := theme.AddDeclarationInSelectorIfNotPresentAlready(string(cssResource.DecodedData), darkThemeSelector, defaultLogoHeightDeclaration)
	if err != nil {
		return fmt.Errorf("failed to migrate %v: %w", cssResource.OriginalPath, err)
	}
	if !added {
		log.Printf("Skipping %s of app (%s) because it already has logo height set", cssResource.OriginalPath, appID)
		return nil
	}
	configSourceData[cssResource.EscapedPath] = base64.StdEncoding.EncodeToString([]byte(migratedCSS))
	if dryRun {
		log.Printf("Before %s updated:", cssResource.OriginalPath)
		log.Printf("\n%s\n", string(cssResource.DecodedData))
		log.Printf("After %s updated:", cssResource.OriginalPath)
		log.Printf("\n%s\n", string(migratedCSS))
	}

	return nil
}

func handleMissingLightThemeCSS(appID string, configSourceData map[string]string, dryRun bool) (err error) {
	log.Printf("Creating light theme css at %s for app (%s) because it was not customized before", LightThemeCSSPath, appID)
	newLightThemeCSS := theme.Ruleset{
		Selector: lightThemeSelector,
		Declarations: []*theme.Declaration{
			&defaultLogoHeightDeclaration,
		},
	}
	var migratedCSSBuf bytes.Buffer
	newLightThemeCSS.Stringify(&migratedCSSBuf, theme.Indentation(""))
	migratedCSS := migratedCSSBuf.Bytes()

	escapedLightThemeCSSPath := filepathutil.EscapePath(LightThemeCSSPath)

	configSourceData[escapedLightThemeCSSPath] = base64.StdEncoding.EncodeToString(migratedCSS)
	if dryRun {
		log.Println("Before updated: no file")
		log.Printf("After %s updated:", LightThemeCSSPath)
		log.Printf("\n%s\n", string(migratedCSS))
	}

	return nil
}

func handleMissingDarkThemeCSS(appID string, configSourceData map[string]string, dryRun bool) (err error) {
	log.Printf("Creating dark theme css at %s for app (%s) because it was not customized before", LightThemeCSSPath, appID)
	newDarkThemeCSS := theme.Ruleset{
		Selector: darkThemeSelector,
		Declarations: []*theme.Declaration{
			&defaultLogoHeightDeclaration,
		},
	}
	var migratedCSSBuf bytes.Buffer
	newDarkThemeCSS.Stringify(&migratedCSSBuf, theme.Indentation(""))
	migratedCSS := migratedCSSBuf.Bytes()

	escapedDarkThemeCSSPath := filepathutil.EscapePath(DarkThemeCSSPath)

	configSourceData[escapedDarkThemeCSSPath] = base64.StdEncoding.EncodeToString(migratedCSS)
	if dryRun {
		log.Println("Before updated: no file")
		log.Printf("After %s updated:", DarkThemeCSSPath)
		log.Printf("\n%s\n", string(migratedCSS))
	}

	return nil
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

func parseLogoHeightConfigSource(configSourceData map[string]string) (*MigrateLogoHeightConfig, error) {
	out := &MigrateLogoHeightConfig{}

	if lightLogoMatched, ok := getMatchingConfigSourcePaths(lightLogoPathRegex, configSourceData); ok {
		firstLogo := lightLogoMatched[0]
		unescapedPath, err := filepathutil.UnescapePath(firstLogo)
		if err != nil {
			return nil, err
		}
		out.LightLogo = newResourceConfig(unescapedPath, firstLogo, configSourceData[firstLogo])
	}

	if darkLogoMatched, ok := getMatchingConfigSourcePaths(darkLogoPathRegex, configSourceData); ok {
		firstLogo := darkLogoMatched[0]
		unescapedPath, err := filepathutil.UnescapePath(firstLogo)
		if err != nil {
			return nil, err
		}
		out.DarkLogo = newResourceConfig(unescapedPath, firstLogo, configSourceData[firstLogo])
	}

	if lightThemeCSSMatch, ok := getMatchingConfigSourcePaths(LightThemeCSSPathRegex, configSourceData); ok {
		firstThemeCSS := lightThemeCSSMatch[0]
		unescapedPath, err := filepathutil.UnescapePath(firstThemeCSS)
		if err != nil {
			return nil, err
		}
		rcd, err := newResourceConfigDecoded(unescapedPath, firstThemeCSS, configSourceData[firstThemeCSS])
		if err != nil {
			return nil, err
		}
		out.LightThemeCSS = rcd
	}

	if darkThemeCSSMatch, ok := getMatchingConfigSourcePaths(DarkThemeCSSPathRegex, configSourceData); ok {
		firstThemeCSS := darkThemeCSSMatch[0]
		unescapedPath, err := filepathutil.UnescapePath(firstThemeCSS)
		if err != nil {
			return nil, err
		}
		rcd, err := newResourceConfigDecoded(unescapedPath, firstThemeCSS, configSourceData[firstThemeCSS])
		if err != nil {
			return nil, err
		}
		out.DarkThemeCSS = rcd
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

// getMatchingConfigSourcePaths get all paths in configSourceData that match the input pattern
//
// It returns a string-slice of path that matches the regexp. If the slice is non-empty, then 2nd return value is true, otherwise it is false.
func getMatchingConfigSourcePaths(pattern *regexp.Regexp, configSourceData map[string]string) (matched []string, ok bool) {
	for p := range configSourceData {
		ep, err := filepathutil.UnescapePath(p)
		if err != nil {
			log.Fatalf("failed to unescape path: %s", err)
			return nil, false
		}
		if ok, _ := regexp.MatchString(pattern.String(), ep); ok {
			matched = append(matched, p) // output escaped path
		}
	}

	return matched, len(matched) > 0
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateSetDefaultLogoHeight)
}
