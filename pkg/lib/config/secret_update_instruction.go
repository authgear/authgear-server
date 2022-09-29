package config

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/lestrrat-go/jwx/jwk"

	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

type SecretUpdateInstructionAction string

const (
	SecretUpdateInstructionActionSet      SecretUpdateInstructionAction = "set"
	SecretUpdateInstructionActionUnset    SecretUpdateInstructionAction = "unset"
	SecretUpdateInstructionActionGenerate SecretUpdateInstructionAction = "generate"
)

type SecretConfigUpdateInstruction struct {
	OAuthSSOProviderCredentialsUpdateInstruction *OAuthSSOProviderCredentialsUpdateInstruction `json:"oauthSSOProviderClientSecrets,omitempty"`
	SMTPServerCredentialsUpdateInstruction       *SMTPServerCredentialsUpdateInstruction       `json:"smtpSecret,omitempty"`
	OAuthClientSecretsUpdateInstruction          *OAuthClientSecretsUpdateInstruction          `json:"oauthClientSecrets,omitempty"`
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

	return newConfig, nil
}

type SecretConfigUpdateInstructionInterface interface {
	ApplyTo(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error)
}

type OAuthSSOProviderCredentialsUpdateInstructionDataItem struct {
	Alias        string `json:"alias,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
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

	idx, _, found := out.LookupDataWithIndex(OAuthSSOProviderCredentialsKey)
	if len(i.Data) == 0 {
		// remove the secret item
		if found {
			out.Secrets = append(out.Secrets[:idx], out.Secrets[idx+1:]...)
		}
		return out, nil
	}

	credentials := &OAuthSSOProviderCredentials{}
	for _, i := range i.Data {
		credentials.Items = append(credentials.Items, OAuthSSOProviderCredentialsItem{
			Alias:        i.Alias,
			ClientSecret: i.ClientSecret,
		})
	}

	var data []byte
	data, err := json.Marshal(credentials)
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

type OAuthClientSecretsUpdateInstruction struct {
	Action SecretUpdateInstructionAction `json:"action,omitempty"`

	GenerateData *OAuthClientSecretsUpdateInstructionGenerateData `json:"generateData,omitempty"`
}

func (i *OAuthClientSecretsUpdateInstruction) ApplyTo(ctx *SecretConfigUpdateInstructionContext, currentConfig *SecretConfig) (*SecretConfig, error) {
	switch i.Action {
	case SecretUpdateInstructionActionGenerate:
		return i.generate(ctx, currentConfig)
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
	_ = keySet.Add(jwkKey)
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

var _ SecretConfigUpdateInstructionInterface = &SecretConfigUpdateInstruction{}
var _ SecretConfigUpdateInstructionInterface = &OAuthSSOProviderCredentialsUpdateInstruction{}
var _ SecretConfigUpdateInstructionInterface = &SMTPServerCredentialsUpdateInstruction{}
var _ SecretConfigUpdateInstructionInterface = &OAuthClientSecretsUpdateInstruction{}
