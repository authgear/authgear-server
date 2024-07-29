package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwk"

	corerand "github.com/authgear/authgear-server/pkg/util/rand"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

type SecretUpdateInstructionAction string

const (
	SecretUpdateInstructionActionSet      SecretUpdateInstructionAction = "set"
	SecretUpdateInstructionActionUnset    SecretUpdateInstructionAction = "unset"
	SecretUpdateInstructionActionGenerate SecretUpdateInstructionAction = "generate"
	SecretUpdateInstructionActionCleanup  SecretUpdateInstructionAction = "cleanup"
	SecretUpdateInstructionActionDelete   SecretUpdateInstructionAction = "delete"
)

type SecretConfigUpdateInstruction struct {
	OAuthSSOProviderCredentialsUpdateInstruction      *OAuthSSOProviderCredentialsUpdateInstruction      `json:"oauthSSOProviderClientSecrets,omitempty"`
	SMTPServerCredentialsUpdateInstruction            *SMTPServerCredentialsUpdateInstruction            `json:"smtpSecret,omitempty"`
	OAuthClientSecretsUpdateInstruction               *OAuthClientSecretsUpdateInstruction               `json:"oauthClientSecrets,omitempty"`
	AdminAPIAuthKeyUpdateInstruction                  *AdminAPIAuthKeyUpdateInstruction                  `json:"adminAPIAuthKey,omitempty"`
	BotProtectionProviderCredentialsUpdateInstruction *BotProtectionProviderCredentialsUpdateInstruction `json:"botProtectionProviderSecret,omitempty"`
}

func (i *SecretConfigUpdateInstruction) ApplyTo(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error) {
	var err error
	newConfig := currentConfig

	if i.OAuthSSOProviderCredentialsUpdateInstruction != nil {
		newConfig, err = i.OAuthSSOProviderCredentialsUpdateInstruction.ApplyTo(ctx, newConfig)
		if err != nil {
			return nil, err
		}
	}

	if i.SMTPServerCredentialsUpdateInstruction != nil {
		newConfig, err = i.SMTPServerCredentialsUpdateInstruction.ApplyTo(ctx, newConfig)
		if err != nil {
			return nil, err
		}
	}

	if i.OAuthClientSecretsUpdateInstruction != nil {
		newConfig, err = i.OAuthClientSecretsUpdateInstruction.ApplyTo(ctx, newConfig)
		if err != nil {
			return nil, err
		}
	}

	if i.AdminAPIAuthKeyUpdateInstruction != nil {
		newConfig, err = i.AdminAPIAuthKeyUpdateInstruction.ApplyTo(ctx, newConfig)
		if err != nil {
			return nil, err
		}
	}

	if i.BotProtectionProviderCredentialsUpdateInstruction != nil {
		newConfig, err = i.BotProtectionProviderCredentialsUpdateInstruction.ApplyTo(ctx, newConfig)
		if err != nil {
			return nil, err
		}
	}

	return newConfig, nil
}

type SecretConfigUpdateInstructionInterface interface {
	ApplyTo(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error)
}

type OAuthSSOProviderCredentialsUpdateInstructionDataItem struct {
	OriginalAlias   *string `json:"originalAlias,omitempty"`
	NewAlias        string  `json:"newAlias,omitempty"`
	NewClientSecret *string `json:"newClientSecret,omitempty"`
}

type OAuthSSOProviderCredentialsUpdateInstruction struct {
	Action SecretUpdateInstructionAction                          `json:"action,omitempty"`
	Data   []OAuthSSOProviderCredentialsUpdateInstructionDataItem `json:"data,omitempty"`
}

func (i *OAuthSSOProviderCredentialsUpdateInstruction) ApplyTo(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error) {
	switch i.Action {
	case SecretUpdateInstructionActionSet:
		return i.set(currentConfig)
	default:
		return nil, fmt.Errorf("config: unexpected action for OAuthSSOProviderCredentialsUpdateInstruction: %s", i.Action)
	}
}

func (i *OAuthSSOProviderCredentialsUpdateInstruction) set(currentConfig *SecretConfig) (*SecretConfig, error) {
	out := &SecretConfig{}
	for _, item := range currentConfig.Secrets {
		out.Secrets = append(out.Secrets, item)
	}

	idx, secretData, found := out.LookupDataWithIndex(OAuthSSOProviderCredentialsKey)
	if len(i.Data) == 0 {
		// remove the secret item
		if found {
			out.Secrets = append(out.Secrets[:idx], out.Secrets[idx+1:]...)
		}
		return out, nil
	}

	existingCredentialItems := []OAuthSSOProviderCredentialsItem{}
	if found {
		existingCredentialItems = secretData.(*OAuthSSOProviderCredentials).Items
	}

	newCredentialItems := []OAuthSSOProviderCredentialsItem{}

	for _, dataItem := range i.Data {
		if dataItem.OriginalAlias == nil {
			// This is a new secret
			if dataItem.NewClientSecret == nil {
				// New secret cannot have null client secret, return error
				return nil, fmt.Errorf("missing client secret for new item")
			}
			newCredentialItems = append(newCredentialItems, OAuthSSOProviderCredentialsItem{
				Alias:        dataItem.NewAlias,
				ClientSecret: *dataItem.NewClientSecret,
			})
		} else {
			// This is an update of exist secret
			var originalItem *OAuthSSOProviderCredentialsItem = nil
			for _, it := range existingCredentialItems {
				existingItem := it
				if existingItem.Alias == *dataItem.OriginalAlias {
					originalItem = &existingItem
					break
				}
			}
			if originalItem == nil {
				// Cannot find the original item, return error
				return nil, fmt.Errorf("original client secret item not found")
			}
			newClientSecret := originalItem.ClientSecret
			if dataItem.NewClientSecret != nil {
				newClientSecret = *dataItem.NewClientSecret
			}
			newCredentialItems = append(newCredentialItems, OAuthSSOProviderCredentialsItem{
				Alias:        dataItem.NewAlias,
				ClientSecret: newClientSecret,
			})
		}
	}

	newCredentials := &OAuthSSOProviderCredentials{
		Items: newCredentialItems,
	}

	var data []byte
	data, err := json.Marshal(newCredentials)
	if err != nil {
		return nil, err
	}
	newSecretItem := SecretItem{
		Key:     OAuthSSOProviderCredentialsKey,
		RawData: json.RawMessage(data),
	}

	if found {
		out.Secrets[idx] = newSecretItem
	} else {
		out.Secrets = append(out.Secrets, newSecretItem)
	}

	return out, nil
}

type SMTPServerCredentialsUpdateInstructionData struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type SMTPServerCredentialsUpdateInstruction struct {
	Action SecretUpdateInstructionAction               `json:"action,omitempty"`
	Data   *SMTPServerCredentialsUpdateInstructionData `json:"data,omitempty"`
}

func (i *SMTPServerCredentialsUpdateInstruction) ApplyTo(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error) {
	switch i.Action {
	case SecretUpdateInstructionActionSet:
		return i.set(currentConfig)
	case SecretUpdateInstructionActionUnset:
		return i.unset(currentConfig)
	default:
		return nil, fmt.Errorf("config: unexpected action for SMTPServerCredentialsUpdateInstruction: %s", i.Action)
	}
}

func (i *SMTPServerCredentialsUpdateInstruction) set(currentConfig *SecretConfig) (*SecretConfig, error) {
	out := &SecretConfig{}
	for _, item := range currentConfig.Secrets {
		out.Secrets = append(out.Secrets, item)
	}

	if i.Data == nil {
		return nil, fmt.Errorf("missing data for SMTPServerCredentialsUpdateInstruction")
	}

	credentials := &SMTPServerCredentials{
		Host:     i.Data.Host,
		Port:     i.Data.Port,
		Username: i.Data.Username,
		Password: i.Data.Password,
	}

	var data []byte
	data, err := json.Marshal(credentials)
	if err != nil {
		return nil, err
	}
	newSecretItem := SecretItem{
		Key:     SMTPServerCredentialsKey,
		RawData: json.RawMessage(data),
	}

	idx, _, found := out.LookupDataWithIndex(SMTPServerCredentialsKey)
	if found {
		out.Secrets[idx] = newSecretItem
	} else {
		out.Secrets = append(out.Secrets, newSecretItem)
	}
	return out, nil
}

func (i *SMTPServerCredentialsUpdateInstruction) unset(currentConfig *SecretConfig) (*SecretConfig, error) {
	out := &SecretConfig{}
	for _, item := range currentConfig.Secrets {
		if item.Key == SMTPServerCredentialsKey {
			continue
		}
		out.Secrets = append(out.Secrets, item)
	}
	return out, nil
}

type OAuthClientSecretsUpdateInstructionGenerateData struct {
	ClientID string `json:"clientID,omitempty"`
}

type OAuthClientSecretsUpdateInstructionCleanupData struct {
	KeepClientIDs []string `json:"keepClientIDs,omitempty"`
}

type OAuthClientSecretsUpdateInstruction struct {
	Action SecretUpdateInstructionAction `json:"action,omitempty"`

	GenerateData *OAuthClientSecretsUpdateInstructionGenerateData `json:"generateData,omitempty"`
	CleanupData  *OAuthClientSecretsUpdateInstructionCleanupData  `json:"cleanupData,omitempty"`
}

func (i *OAuthClientSecretsUpdateInstruction) ApplyTo(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error) {
	switch i.Action {
	case SecretUpdateInstructionActionGenerate:
		return i.generate(ctx, currentConfig)
	case SecretUpdateInstructionActionCleanup:
		return i.cleanup(currentConfig)
	default:
		return nil, fmt.Errorf("config: unexpected action for OAuthClientSecretsUpdateInstruction: %s", i.Action)
	}
}

func (i *OAuthClientSecretsUpdateInstruction) decodeOAuthClientCredentials(rawData json.RawMessage) (*OAuthClientCredentials, error) {
	decoder := json.NewDecoder(bytes.NewReader(rawData))
	data := &OAuthClientCredentials{}
	err := decoder.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("config: failed to decode OAuthClientCredentials in authgear.secrets.yaml: %w", err)
	}
	return data, nil
}

func (i *OAuthClientSecretsUpdateInstruction) generate(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error) {
	out := &SecretConfig{}
	for _, item := range currentConfig.Secrets {
		out.Secrets = append(out.Secrets, item)
	}

	if i.GenerateData == nil || i.GenerateData.ClientID == "" {
		return nil, fmt.Errorf("config: missing client id for OAuthClientSecretsUpdateInstruction")
	}

	clientID := i.GenerateData.ClientID
	jwkKey := ctx.GenerateClientSecretOctetKeyFunc(ctx.Clock.NowUTC(), corerand.SecureRand)
	keySet := jwk.NewSet()
	_ = keySet.AddKey(jwkKey)
	newCredentialsItem := OAuthClientCredentialsItem{
		ClientID:                     clientID,
		OAuthClientCredentialsKeySet: OAuthClientCredentialsKeySet{Set: keySet},
	}

	newOAuthClientCredentials := &OAuthClientCredentials{}
	idx, item, found := out.Lookup(OAuthClientCredentialsKey)
	if found {
		oauth, err := i.decodeOAuthClientCredentials(item.RawData)
		if err != nil {
			return nil, err
		}
		_, ok := oauth.Lookup(clientID)
		if ok {
			return nil, fmt.Errorf("config: client secret already exist")
		}
		// copy oauth client secret items from the current config to new config
		newOAuthClientCredentials.Items = make([]OAuthClientCredentialsItem, len(oauth.Items))
		copy(newOAuthClientCredentials.Items, oauth.Items)
	}

	// Add new credentials item to the OAuthClientCredentials
	newOAuthClientCredentials.Items = append(newOAuthClientCredentials.Items, newCredentialsItem)
	var jsonData []byte
	jsonData, err := json.Marshal(newOAuthClientCredentials)
	if err != nil {
		return nil, err
	}
	newSecretItem := SecretItem{
		Key:     OAuthClientCredentialsKey,
		RawData: json.RawMessage(jsonData),
	}

	if found {
		out.Secrets[idx] = newSecretItem
	} else {
		out.Secrets = append(out.Secrets, newSecretItem)
	}

	return out, nil
}

func (i *OAuthClientSecretsUpdateInstruction) cleanup(currentConfig *SecretConfig) (*SecretConfig, error) {
	out := &SecretConfig{}
	out.Secrets = make([]SecretItem, len(currentConfig.Secrets))
	copy(out.Secrets, currentConfig.Secrets)

	if i.CleanupData == nil || i.CleanupData.KeepClientIDs == nil {
		return nil, fmt.Errorf("config: missing keepClientIDs for OAuthClientSecretsUpdateInstruction")
	}

	idx, item, found := out.Lookup(OAuthClientCredentialsKey)
	if !found {
		return out, nil
	}
	oauth, err := i.decodeOAuthClientCredentials(item.RawData)
	if err != nil {
		return nil, err
	}

	keepClientIDSet := setutil.NewSetFromSlice(i.CleanupData.KeepClientIDs, setutil.Identity[string])
	newOAuthClientCredentials := &OAuthClientCredentials{}
	for _, item := range oauth.Items {
		if _, ok := keepClientIDSet[item.ClientID]; ok {
			newOAuthClientCredentials.Items = append(newOAuthClientCredentials.Items, item)
		}
	}

	if len(newOAuthClientCredentials.Items) == 0 {
		// remove oauth.client_secrets from secrets
		out.Secrets = append(out.Secrets[:idx], out.Secrets[idx+1:]...)
	} else {
		var jsonData []byte
		jsonData, err := json.Marshal(newOAuthClientCredentials)
		if err != nil {
			return nil, err
		}
		newSecretItem := SecretItem{
			Key:     OAuthClientCredentialsKey,
			RawData: json.RawMessage(jsonData),
		}
		out.Secrets[idx] = newSecretItem
	}

	return out, nil
}

type AdminAPIAuthKeyUpdateInstructionDeleteData struct {
	KeyID string `json:"keyID,omitempty"`
}

type AdminAPIAuthKeyUpdateInstruction struct {
	Action SecretUpdateInstructionAction `json:"action,omitempty"`

	DeleteData *AdminAPIAuthKeyUpdateInstructionDeleteData `json:"deleteData,omitempty"`
}

func (i *AdminAPIAuthKeyUpdateInstruction) ApplyTo(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error) {
	switch i.Action {
	case SecretUpdateInstructionActionGenerate:
		return i.generate(ctx, currentConfig)
	case SecretUpdateInstructionActionDelete:
		return i.delete(currentConfig)
	default:
		return nil, fmt.Errorf("config: unexpected action for AdminAPIAuthKeyUpdateInstruction: %s", i.Action)
	}
}

func (i *AdminAPIAuthKeyUpdateInstruction) decodeAdminAPIAuthKey(rawData json.RawMessage) (*AdminAPIAuthKey, error) {
	decoder := json.NewDecoder(bytes.NewReader(rawData))
	data := &AdminAPIAuthKey{}
	err := decoder.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("config: failed to decode AdminAPIAuthKey in authgear.secrets.yaml: %w", err)
	}
	return data, nil
}

func (i *AdminAPIAuthKeyUpdateInstruction) generate(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error) {
	out := &SecretConfig{}
	out.Secrets = make([]SecretItem, len(currentConfig.Secrets))
	copy(out.Secrets, currentConfig.Secrets)

	newAuthKey := ctx.GenerateAdminAPIAuthKeyFunc(ctx.Clock.NowUTC(), corerand.SecureRand)

	newAdminAPIAuthKey := &AdminAPIAuthKey{Set: jwk.NewSet()}
	idx, item, found := out.Lookup(AdminAPIAuthKeyKey)
	if found {
		authKey, err := i.decodeAdminAPIAuthKey(item.RawData)
		if err != nil {
			return nil, err
		}
		// copy auth key set from the current config to new config
		newAdminAPIAuthKey.Set, err = authKey.Clone()
		if err != nil {
			return nil, err
		}
	}

	// Add new key to the AdminAPIAuthKey
	_ = newAdminAPIAuthKey.AddKey(newAuthKey)
	if newAdminAPIAuthKey.Len() > 2 {
		return nil, fmt.Errorf("config: must have at most two Admin API auth keys")
	} else {
		var jsonData []byte
		jsonData, err := json.Marshal(newAdminAPIAuthKey)
		if err != nil {
			return nil, err
		}
		newSecretItem := SecretItem{
			Key:     AdminAPIAuthKeyKey,
			RawData: json.RawMessage(jsonData),
		}

		if found {
			out.Secrets[idx] = newSecretItem
		} else {
			out.Secrets = append(out.Secrets, newSecretItem)
		}
	}
	return out, nil
}

func (i *AdminAPIAuthKeyUpdateInstruction) delete(currentConfig *SecretConfig) (*SecretConfig, error) {
	out := &SecretConfig{}
	out.Secrets = make([]SecretItem, len(currentConfig.Secrets))
	copy(out.Secrets, currentConfig.Secrets)

	if i.DeleteData == nil || i.DeleteData.KeyID == "" {
		return nil, fmt.Errorf("config: missing KeyID for AdminAPIAuthKeyUpdateInstruction")
	}

	idx, item, found := out.Lookup(AdminAPIAuthKeyKey)
	if !found {
		return out, nil
	}
	authKey, err := i.decodeAdminAPIAuthKey(item.RawData)
	if err != nil {
		return nil, err
	}

	newAdminAPIAuthKey := &AdminAPIAuthKey{Set: jwk.NewSet()}
	for it := authKey.Keys(context.Background()); it.Next(context.Background()); {
		if key, ok := it.Pair().Value.(jwk.Key); ok && key.KeyID() != i.DeleteData.KeyID {
			_ = newAdminAPIAuthKey.AddKey(key)
		}
	}

	if newAdminAPIAuthKey.Len() == 0 {
		return nil, fmt.Errorf("config: must have at least one Admin API auth key")
	} else {
		var jsonData []byte
		jsonData, err := json.Marshal(newAdminAPIAuthKey)
		if err != nil {
			return nil, err
		}
		newSecretItem := SecretItem{
			Key:     AdminAPIAuthKeyKey,
			RawData: json.RawMessage(jsonData),
		}
		out.Secrets[idx] = newSecretItem
	}

	return out, nil
}

type BotProtectionProviderCredentialsUpdateInstructionData struct {
	Type      string `json:"type,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`
}

type BotProtectionProviderCredentialsUpdateInstruction struct {
	Action SecretUpdateInstructionAction                          `json:"action,omitempty"`
	Data   *BotProtectionProviderCredentialsUpdateInstructionData `json:"data,omitempty"`
}

func (i *BotProtectionProviderCredentialsUpdateInstruction) ApplyTo(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error) {
	switch i.Action {
	case SecretUpdateInstructionActionSet:
		return i.set(currentConfig)
	default:
		return nil, fmt.Errorf("config: unexpected action for BotProtectionProviderCredentialsUpdateInstruction: %s", i.Action)
	}
}

func (i *BotProtectionProviderCredentialsUpdateInstruction) set(currentConfig *SecretConfig) (*SecretConfig, error) {
	out := &SecretConfig{}
	for _, item := range currentConfig.Secrets {
		out.Secrets = append(out.Secrets, item)
	}

	if i.Data == nil {
		return nil, fmt.Errorf("missing data for BotProtectionProviderCredentialsUpdateInstruction")
	}

	credentials := &BotProtectionProviderCredentials{
		Type:      BotProtectionProviderType(i.Data.Type),
		SecretKey: i.Data.SecretKey,
	}

	var data []byte
	data, err := json.Marshal(credentials)
	if err != nil {
		return nil, err
	}
	newSecretItem := SecretItem{
		Key:     BotProtectionProviderCredentialsKey,
		RawData: json.RawMessage(data),
	}

	idx, _, found := out.LookupDataWithIndex(BotProtectionProviderCredentialsKey)
	if found {
		out.Secrets[idx] = newSecretItem
	} else {
		out.Secrets = append(out.Secrets, newSecretItem)
	}
	return out, nil
}

var _ SecretConfigUpdateInstructionInterface = &SecretConfigUpdateInstruction{}
var _ SecretConfigUpdateInstructionInterface = &OAuthSSOProviderCredentialsUpdateInstruction{}
var _ SecretConfigUpdateInstructionInterface = &SMTPServerCredentialsUpdateInstruction{}
var _ SecretConfigUpdateInstructionInterface = &OAuthClientSecretsUpdateInstruction{}
var _ SecretConfigUpdateInstructionInterface = &AdminAPIAuthKeyUpdateInstruction{}
var _ SecretConfigUpdateInstructionInterface = &BotProtectionProviderCredentialsUpdateInstruction{}
