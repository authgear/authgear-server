package config

import (
	"encoding/json"
	"fmt"
)

type SecretUpdateInstructionAction string

const (
	SecretUpdateInstructionActionSet   SecretUpdateInstructionAction = "set"
	SecretUpdateInstructionActionUnset SecretUpdateInstructionAction = "unset"
)

type OAuthSSOProviderCredentialsUpdateInstructionDataItem struct {
	Alias        string `json:"alias,omitempty"`
	ClientSecret string `json:"clientSecret,omitempty"`
}

type OAuthSSOProviderCredentialsUpdateInstruction struct {
	Action SecretUpdateInstructionAction                          `json:"action,omitempty"`
	Data   []OAuthSSOProviderCredentialsUpdateInstructionDataItem `json:"data,omitempty"`
}

func (i *OAuthSSOProviderCredentialsUpdateInstruction) ApplyTo(currentConfig *SecretConfig) (*SecretConfig, error) {
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

func (i *SMTPServerCredentialsUpdateInstruction) ApplyTo(currentConfig *SecretConfig) (*SecretConfig, error) {
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

type SecretConfigUpdateInstruction struct {
	OAuthSSOProviderCredentialsUpdateInstruction *OAuthSSOProviderCredentialsUpdateInstruction `json:"oauthSSOProviderClientSecrets,omitempty"`
	SMTPServerCredentialsUpdateInstruction       *SMTPServerCredentialsUpdateInstruction       `json:"smtpSecret,omitempty"`
}

func (i *SecretConfigUpdateInstruction) ApplyTo(currentConfig *SecretConfig) (*SecretConfig, error) {
	var err error
	newConfig := currentConfig

	if i.OAuthSSOProviderCredentialsUpdateInstruction != nil {
		newConfig, err = i.OAuthSSOProviderCredentialsUpdateInstruction.ApplyTo(newConfig)
		if err != nil {
			return nil, err
		}
	}

	if i.SMTPServerCredentialsUpdateInstruction != nil {
		newConfig, err = i.SMTPServerCredentialsUpdateInstruction.ApplyTo(newConfig)
		if err != nil {
			return nil, err
		}
	}

	return newConfig, nil
}
