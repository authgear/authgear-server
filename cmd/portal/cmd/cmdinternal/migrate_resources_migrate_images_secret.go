package cmdinternal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	portalcmd "github.com/authgear/authgear-server/cmd/portal/cmd"
	"github.com/authgear/authgear-server/cmd/portal/internal"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/rand"
	utilsecrets "github.com/authgear/authgear-server/pkg/util/secrets"
)

var cmdInternalMigrateImagesSecret = &cobra.Command{
	Use:   "migrate-images-secret",
	Short: "Generate images secret to existing apps",
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
			UpdateConfigSourceFunc: migrateImagesSecret,
			DryRun:                 &MigrateResourcesDryRun,
		})

		return nil
	},
}

func migrateImagesSecret(appID string, configSourceData map[string]string, dryRun bool) error {
	encodedData := configSourceData["authgear.secrets.yaml"]
	decoded, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return fmt.Errorf("failed decode authgear.secrets.yaml: %w", err)
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

	secrets := m["secrets"].([]interface{})
	for _, secretItemIface := range secrets {
		secretItem := secretItemIface.(map[string]interface{})
		key := secretItem["key"].(string)
		if key == string(config.ImagesKeyMaterialsKey) {
			if dryRun {
				log.Printf("Images secret found, not need to generate")
			}
			return nil
		}
	}

	// generate new images key
	createdAt := time.Now().UTC()
	jwkKey := utilsecrets.GenerateOctetKeyForSig(createdAt, rand.SecureRand)
	keySet := jwk.NewSet()
	_ = keySet.AddKey(jwkKey)
	imagesKeySet := &config.ImagesKeyMaterials{Set: keySet}

	dataBytes, err := json.Marshal(imagesKeySet)
	if err != nil {
		return err
	}

	var dataJSON map[string]interface{}
	err = json.Unmarshal(dataBytes, &dataJSON)
	if err != nil {
		return err
	}

	secretItem := map[string]interface{}{}
	secretItem["key"] = config.ImagesKeyMaterialsKey
	secretItem["data"] = dataJSON
	secrets = append(secrets, secretItem)
	m["secrets"] = secrets

	migrated, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed marshal yaml: %w", err)
	}

	if dryRun {
		log.Printf("After updated:")
		log.Printf("\n%s\n", string(migrated))
	}

	configSourceData["authgear.secrets.yaml"] = base64.StdEncoding.EncodeToString(migrated)
	return nil
}

func init() {
	cmdInternalBreakingChangeMigrateResources.AddCommand(cmdInternalMigrateImagesSecret)
}
