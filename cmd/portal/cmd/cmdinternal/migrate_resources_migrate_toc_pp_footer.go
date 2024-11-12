package cmdinternal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
)

const defaultTermsOfServiceLink = "https://www.authgear.com/terms"
const defaultPrivacyPolicyLink = "https://www.authgear.com/data-privacy"

const translationFilePathFormat = "^templates_2f_([a-zA-Z0-9-]+)_2f_translation.json$"

var translationFilePathRegex = regexp.MustCompile(translationFilePathFormat)

var cmdInternalMigrateTOCPPFooter = &cobra.Command{
	Use:   "migrate-toc-pp-footer",
	Short: "Set default terms of service and privacy policy link in translation.json",
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
			UpdateConfigSourceFunc: migrateTOCPPFooter,
			DryRun:                 &MigrateResourcesDryRun,
		})

		return nil
	},
}

//nolint:gocognit
func migrateTOCPPFooter(appID string, configSourceData map[string]string, dryRun bool) error {
	encodedConfig := configSourceData["authgear.yaml"]
	decodedConfig, err := base64.StdEncoding.DecodeString(encodedConfig)
	if err != nil {
		return fmt.Errorf("failed decode authgear.yaml: %w", err)
	}

	m := make(map[string]interface{})
	err = yaml.Unmarshal(decodedConfig, &m)
	if err != nil {
		return fmt.Errorf("failed unmarshal yaml: %w", err)
	}

	var supportedLanguageTags []string

	emptyJSON, err := json.Marshal(map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed marshal empty json: %w", err)
	}

	// Get all supported languages tags from the config
	locale, ok := m["localization"].(map[string]interface{})
	if ok {
		sl, ok := locale["supported_languages"].([]interface{})
		if ok {
			for _, t := range sl {
				supportedLanguageTags = append(supportedLanguageTags, t.(string))
			}
		} else {
			supportedLanguageTags = append(supportedLanguageTags, "en")
		}
	} else {
		supportedLanguageTags = append(supportedLanguageTags, "en")
	}

	// Create translation.json file if it does not exist for all supported languages
	for _, t := range supportedLanguageTags {
		f := fmt.Sprintf("templates_2f_%s_2f_translation.json", t)
		if _, ok = configSourceData[f]; !ok {
			configSourceData[f] = base64.StdEncoding.EncodeToString(emptyJSON)
		}
	}

	// Rewrite the translation.json file
	for k := range configSourceData {
		if ok := translationFilePathRegex.MatchString(k); ok {
			encodedData := configSourceData[k]
			decoded, err := base64.StdEncoding.DecodeString(encodedData)
			if err != nil {
				return fmt.Errorf("failed to decode %v: %w", k, err)
			}

			if dryRun {
				log.Printf("Converting app (%s) with file (%s)", appID, k)
				log.Printf("Before updated:")
				log.Printf("\n%s\n", string(decoded))
			}

			m := make(map[string]interface{})
			err = json.Unmarshal(decoded, &m)
			if err != nil {
				return fmt.Errorf("failed unmarshal %s: %w", k, err)
			}

			if _, ok := m["terms-of-service-link"].(string); !ok {
				m["terms-of-service-link"] = defaultTermsOfServiceLink
			}

			if _, ok := m["privacy-policy-link"].(string); !ok {
				m["privacy-policy-link"] = defaultPrivacyPolicyLink
			}

			migrated, err := json.Marshal(m)
			if err != nil {
				return fmt.Errorf("failed to marshal %s: %w", k, err)
			}

			if dryRun {
				log.Printf("After updated:")
				log.Printf("\n%s\n", string(migrated))
			}

			configSourceData[k] = base64.StdEncoding.EncodeToString(migrated)
		}
	}
	return nil
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateTOCPPFooter)
}
