package sms

import (
	"fmt"
	"strconv"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/custom"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/nexmo"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/smsapi"
	"github.com/authgear/authgear-server/pkg/lib/infra/sms/twilio"
)

func NewTwilioClientCredentialsFromSecrets(secret *config.TwilioCredentials) *TwilioClientCredentials {
	return &TwilioClientCredentials{
		CredentialType:      secret.GetCredentialType(),
		AccountSID:          secret.AccountSID,
		AuthToken:           secret.AuthToken,
		APIKeySID:           secret.APIKeySID,
		APIKeySecret:        secret.APIKeySecret,
		MessagingServiceSID: secret.MessagingServiceSID,
	}
}

type TwilioClientCredentials struct {
	CredentialType      config.TwilioCredentialType
	AccountSID          string
	AuthToken           string
	APIKeySID           string
	APIKeySecret        string
	MessagingServiceSID string
}

func (c *TwilioClientCredentials) toSecret() *config.TwilioCredentials {
	if c == nil {
		return nil
	}
	return &config.TwilioCredentials{
		CredentialType_WriteOnly: &c.CredentialType,
		AccountSID:               c.AccountSID,
		AuthToken:                c.AuthToken,
		APIKeySID:                c.APIKeySID,
		APIKeySecret:             c.APIKeySecret,
		MessagingServiceSID:      c.MessagingServiceSID,
	}
}

func (TwilioClientCredentials) smsClientCredentials() {}

type NexmoClientCredentials struct {
	APIKey    string
	APISecret string
}

func (NexmoClientCredentials) smsClientCredentials() {}

type CustomClientCredentials struct {
	URL     string
	Timeout *config.DurationSeconds
}

func (CustomClientCredentials) smsClientCredentials() {}

type ClientResolver struct {
	AuthgearYAMLSMSProvider config.SMSProvider
	AuthgearYAMLSMSGateway  *config.SMSGatewayConfig

	AuthgearSecretsYAMLNexmoCredentials        *config.NexmoCredentials
	AuthgearSecretsYAMLTwilioCredentials       *config.TwilioCredentials
	AuthgearSecretsYAMLCustomSMSProviderConfig *config.CustomSMSProviderConfig

	EnvironmentDefaultProvider      config.SMSGatewayEnvironmentDefaultProvider
	EnvironmentDefaultUseConfigFrom config.SMSGatewayEnvironmentDefaultUseConfigFrom

	EnvironmentNexmoCredentials        config.SMSGatewayEnvironmentNexmoCredentials
	EnvironmentTwilioCredentials       config.SMSGatewayEnvironmentTwilioCredentials
	EnvironmentCustomSMSProviderConfig config.SMSGatewayEnvironmentCustomSMSProviderConfig

	SMSDenoHook custom.SMSDenoHook
	SMSWebHook  custom.SMSWebHook
}

func (r *ClientResolver) ResolveClient() (smsapi.Client, SMSClientCredentials, error) {
	nexmoClient, nexmoClientCredentials, twilioClient, twilioClientCredentials, customClient, customClientCredentials := r.resolveRawClients()
	provider := r.resolveProvider()

	var client smsapi.Client
	var smsClientCredentials SMSClientCredentials
	switch provider {
	case config.SMSProviderNexmo:
		if nexmoClient == nil {
			return nil, nil, smsapi.ErrNoAvailableClient
		}
		client = nexmoClient
		smsClientCredentials = nexmoClientCredentials
	case config.SMSProviderTwilio:
		if twilioClient == nil {
			return nil, nil, smsapi.ErrNoAvailableClient
		}
		client = twilioClient
		smsClientCredentials = twilioClientCredentials
	case config.SMSProviderCustom:
		if customClient == nil {
			return nil, nil, smsapi.ErrNoAvailableClient
		}
		client = customClient
		smsClientCredentials = customClientCredentials
	default:
		var availableClients []struct {
			RawClient            smsapi.Client
			SMSClientCredentials SMSClientCredentials
		} = []struct {
			RawClient            smsapi.Client
			SMSClientCredentials SMSClientCredentials
		}{}

		if nexmoClient != nil {
			availableClients = append(availableClients, struct {
				RawClient            smsapi.Client
				SMSClientCredentials SMSClientCredentials
			}{
				RawClient:            nexmoClient,
				SMSClientCredentials: nexmoClientCredentials,
			})
		}
		if twilioClient != nil {
			availableClients = append(availableClients, struct {
				RawClient            smsapi.Client
				SMSClientCredentials SMSClientCredentials
			}{
				RawClient:            twilioClient,
				SMSClientCredentials: twilioClientCredentials,
			})
		}
		if customClient != nil {
			availableClients = append(availableClients, struct {
				RawClient            smsapi.Client
				SMSClientCredentials SMSClientCredentials
			}{
				RawClient:            customClient,
				SMSClientCredentials: customClientCredentials,
			})
		}
		if len(availableClients) == 0 {
			return nil, nil, smsapi.ErrNoAvailableClient
		}
		if len(availableClients) > 1 {
			return nil, nil, smsapi.ErrAmbiguousClient
		}
		client = availableClients[0].RawClient
		smsClientCredentials = availableClients[0].SMSClientCredentials
	}
	return client, smsClientCredentials, nil
}

func (r *ClientResolver) resolveProvider() config.SMSProvider {
	if r.AuthgearYAMLSMSGateway != nil {
		// Use sms gateway config. See Table 3
		return r.resolveProviderFromAuthgearYAMLAndAuthgearSecretsYAML()
	}
	if r.AuthgearYAMLSMSProvider != "" {
		// Use `messaging.sms_provider` from `authgear.yaml`. Read config from `sms.{messaging.sms_provider}` from `authgear.secrets.yaml`
		return r.AuthgearYAMLSMSProvider
	}
	// sms_provider == "" and sms_gateway == nil
	// See table 2
	return r.resolveProviderFromEnv()
}

// Table 2
func (r *ClientResolver) resolveProviderFromEnv() config.SMSProvider {
	if r.EnvironmentDefaultUseConfigFrom == "" {
		// `provider` will be determined from application logic. Read config from `sms.{provider}` from `authgear.secrets.yaml`
		return ""
	}
	switch r.EnvironmentDefaultUseConfigFrom {
	case config.SMSGatewayEnvironmentDefaultUseConfigFromEnvironmentVariable:
		if r.EnvironmentDefaultProvider == "" {
			// `provider` will be determined from application logic. Read config from `SMS_GATEWAY_{provider}_*` from environment variables
			return ""
		}
		// Use `SMS_GATEWAY_DEFAULT_PROVIDER` as provider. Will read config from `SMS_GATEWAY_{SMS_GATEWAY_DEFAULT_PROVIDER}_*` environment variables
		return config.SMSProvider(r.EnvironmentDefaultProvider)
	case config.SMSGatewayEnvironmentDefaultUseConfigFromAuthgearSecretsYAML:
		// `provider` will be determined from application logic. Read config from `sms.{provider}` from `authgear.secrets.yaml`
		return ""
	default:
		panic(fmt.Errorf("Invalid DEFAULT_USE_CONFIG_FROM %v", r.EnvironmentDefaultUseConfigFrom))
	}
}

// Table 3
func (r *ClientResolver) resolveProviderFromAuthgearYAMLAndAuthgearSecretsYAML() config.SMSProvider {
	AuthgearYAMLUseConfigFrom := r.AuthgearYAMLSMSGateway.UseConfigFrom
	switch AuthgearYAMLUseConfigFrom {
	case config.SMSGatewayUseConfigFromEnvironmentVariable:
		if r.AuthgearYAMLSMSGateway.Provider == "" {
			if r.EnvironmentDefaultProvider == "" {
				// provider` will be determined from application logic. Read config from `SMS_GATEWAY_{provider}_*` from environment variables
				return ""
			}
			// Use `SMS_GATEWAY_DEFAULT_PROVIDER` as provider. Will read config from `SMS_GATEWAY_{SMS_GATEWAY_DEFAULT_PROVIDER}_*` environment variables
			return config.SMSProvider(r.EnvironmentDefaultProvider)
		}
		// Use `sms_gateway.provider` as provider. Will read config from `SMS_GATEWAY_{sms_gateway.provider}_*` environment variables
		return r.AuthgearYAMLSMSGateway.Provider
	case config.SMSGatewayUseConfigFromAuthgearSecretsYAML:
		// `sms_gateway.provider` is required
		// Use provider configs from `authgear.yaml`. Will read config from `sms.{sms_gateway.provider}` from `authgear.secrets.yaml`
		return r.AuthgearYAMLSMSGateway.Provider
	default:
		panic(fmt.Errorf("Invalid sms_gateway.use_config_from %v", AuthgearYAMLUseConfigFrom))
	}
}

func (r *ClientResolver) clientsFromAuthgearSecretsYAML() (*nexmo.NexmoClient, *NexmoClientCredentials, *twilio.TwilioClient, *TwilioClientCredentials, *custom.CustomClient, *CustomClientCredentials) {
	var nexmoClientCredentials *NexmoClientCredentials
	var twilioClientCredentials *TwilioClientCredentials
	var customClientCredentials *CustomClientCredentials

	if r.AuthgearSecretsYAMLNexmoCredentials != nil {
		nexmoClientCredentials = &NexmoClientCredentials{
			APIKey:    r.AuthgearSecretsYAMLNexmoCredentials.APIKey,
			APISecret: r.AuthgearSecretsYAMLNexmoCredentials.APISecret,
		}
	}

	if r.AuthgearSecretsYAMLTwilioCredentials != nil {
		credtyp := r.AuthgearSecretsYAMLTwilioCredentials.GetCredentialType()
		twilioClientCredentials = &TwilioClientCredentials{
			CredentialType:      credtyp,
			AccountSID:          r.AuthgearSecretsYAMLTwilioCredentials.AccountSID,
			AuthToken:           r.AuthgearSecretsYAMLTwilioCredentials.AuthToken,
			APIKeySID:           r.AuthgearSecretsYAMLTwilioCredentials.APIKeySID,
			APIKeySecret:        r.AuthgearSecretsYAMLTwilioCredentials.APIKeySecret,
			MessagingServiceSID: r.AuthgearSecretsYAMLTwilioCredentials.MessagingServiceSID,
		}
	}

	if r.AuthgearSecretsYAMLCustomSMSProviderConfig != nil {
		customClientCredentials = &CustomClientCredentials{
			URL:     r.AuthgearSecretsYAMLCustomSMSProviderConfig.URL,
			Timeout: r.AuthgearSecretsYAMLCustomSMSProviderConfig.Timeout,
		}
	}

	return nexmo.NewNexmoClient(r.AuthgearSecretsYAMLNexmoCredentials), nexmoClientCredentials, twilio.NewTwilioClient(r.AuthgearSecretsYAMLTwilioCredentials), twilioClientCredentials, custom.NewCustomClient(r.AuthgearSecretsYAMLCustomSMSProviderConfig, r.SMSDenoHook, r.SMSWebHook), customClientCredentials
}

func (r *ClientResolver) clientsFromEnv() (*nexmo.NexmoClient, *NexmoClientCredentials, *twilio.TwilioClient, *TwilioClientCredentials, *custom.CustomClient, *CustomClientCredentials) {
	var nexmoClientCredentials *NexmoClientCredentials
	var twilioClientCredentials *TwilioClientCredentials
	var customClientCredentials *CustomClientCredentials

	print(fmt.Sprintf("%v", r.EnvironmentNexmoCredentials))
	if r.EnvironmentNexmoCredentials != (config.SMSGatewayEnvironmentNexmoCredentials{}) {
		nexmoClientCredentials = &NexmoClientCredentials{
			APIKey:    r.EnvironmentNexmoCredentials.APIKey,
			APISecret: r.EnvironmentNexmoCredentials.APISecret,
		}
	}

	if r.EnvironmentTwilioCredentials != (config.SMSGatewayEnvironmentTwilioCredentials{}) {
		credtyp := config.TwilioCredentialTypeAuthToken
		twilioClientCredentials = &TwilioClientCredentials{
			CredentialType:      credtyp,
			AccountSID:          r.EnvironmentTwilioCredentials.AccountSID,
			AuthToken:           r.EnvironmentTwilioCredentials.AuthToken,
			APIKeySID:           "",
			APIKeySecret:        "",
			MessagingServiceSID: r.EnvironmentTwilioCredentials.MessagingServiceSID,
		}
	}

	if r.EnvironmentCustomSMSProviderConfig != (config.SMSGatewayEnvironmentCustomSMSProviderConfig{}) {
		timeoutInt, _ := strconv.Atoi(r.EnvironmentCustomSMSProviderConfig.Timeout)
		var timeout *config.DurationSeconds
		timeout = new(config.DurationSeconds)
		*timeout = config.DurationSeconds(timeoutInt)
		customClientCredentials = &CustomClientCredentials{
			URL:     r.EnvironmentCustomSMSProviderConfig.URL,
			Timeout: timeout,
		}
	}

	return nexmo.NewNexmoClient((*config.NexmoCredentials)(nexmoClientCredentials)), nexmoClientCredentials, twilio.NewTwilioClient(twilioClientCredentials.toSecret()), twilioClientCredentials, custom.NewCustomClient((*config.CustomSMSProviderConfig)(customClientCredentials), r.SMSDenoHook, r.SMSWebHook), customClientCredentials
}

func (r *ClientResolver) resolveRawClients() (*nexmo.NexmoClient, *NexmoClientCredentials, *twilio.TwilioClient, *TwilioClientCredentials, *custom.CustomClient, *CustomClientCredentials) {
	if r.AuthgearYAMLSMSGateway != nil {
		// Use sms gateway config. See Table 3
		return r.resolveConfigFromAuthgearYAMLAndAuthgearSecretsYAML()
	}
	if r.AuthgearYAMLSMSProvider != "" {
		// Use `messaging.sms_provider` from `authgear.yaml`. Read config from `sms.{messaging.sms_provider}` from `authgear.secrets.yaml`
		return r.clientsFromAuthgearSecretsYAML()
	}
	// sms_provider == "" and sms_gateway == nil
	// See table 2
	return r.resolveConfigFromEnv()
}

// Table 2
func (r *ClientResolver) resolveConfigFromEnv() (*nexmo.NexmoClient, *NexmoClientCredentials, *twilio.TwilioClient, *TwilioClientCredentials, *custom.CustomClient, *CustomClientCredentials) {
	if r.EnvironmentDefaultUseConfigFrom == "" {
		// `provider` will be determined from application logic. Read config from `sms.{provider}` from `authgear.secrets.yaml`
		return r.clientsFromAuthgearSecretsYAML()
	}
	switch r.EnvironmentDefaultUseConfigFrom {
	case config.SMSGatewayEnvironmentDefaultUseConfigFromEnvironmentVariable:
		if r.EnvironmentDefaultProvider == "" {
			// `provider` will be determined from application logic. Read config from `SMS_GATEWAY_{provider}_*` from environment variables
			return r.clientsFromEnv()
		}
		// Use `SMS_GATEWAY_DEFAULT_PROVIDER` as provider. Will read config from `SMS_GATEWAY_{SMS_GATEWAY_DEFAULT_PROVIDER}_*` environment variables
		return r.clientsFromEnv()
	case config.SMSGatewayEnvironmentDefaultUseConfigFromAuthgearSecretsYAML:
		// `provider` will be determined from application logic. Read config from `sms.{provider}` from `authgear.secrets.yaml`
		return r.clientsFromAuthgearSecretsYAML()
	default:
		panic(fmt.Errorf("Invalid DEFAULT_USE_CONFIG_FROM %v", r.EnvironmentDefaultUseConfigFrom))
	}
}

// Table 3
func (r *ClientResolver) resolveConfigFromAuthgearYAMLAndAuthgearSecretsYAML() (*nexmo.NexmoClient, *NexmoClientCredentials, *twilio.TwilioClient, *TwilioClientCredentials, *custom.CustomClient, *CustomClientCredentials) {
	switch r.AuthgearYAMLSMSGateway.UseConfigFrom {
	case config.SMSGatewayUseConfigFromEnvironmentVariable:
		if r.AuthgearYAMLSMSGateway.Provider == "" {
			if r.EnvironmentDefaultProvider == "" {
				// provider` will be determined from application logic. Read config from `SMS_GATEWAY_{provider}_*` from environment variables
				return r.clientsFromEnv()
			}
			// Use `SMS_GATEWAY_DEFAULT_PROVIDER` as provider. Will read config from `SMS_GATEWAY_{SMS_GATEWAY_DEFAULT_PROVIDER}_*` environment variables
			return r.clientsFromEnv()
		}
		// Use `sms_gateway.provider` as provider. Will read config from `SMS_GATEWAY_{sms_gateway.provider}_*` environment variables
		return r.clientsFromEnv()
	case config.SMSGatewayUseConfigFromAuthgearSecretsYAML:
		// `sms_gateway.provider` is required
		// Use provider configs from `authgear.yaml`. Will read config from `sms.{sms_gateway.provider}` from `authgear.secrets.yaml`
		return r.clientsFromAuthgearSecretsYAML()
	default:
		panic(fmt.Errorf("Invalid sms_gateway.use_config_from %v", r.AuthgearYAMLSMSGateway.UseConfigFrom))
	}
}
