package sms

import (
	"fmt"
	"strconv"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

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

	SMSDenoHook SMSDenoHook
	SMSWebHook  SMSWebHook
}

func (r *ClientResolver) ResolveClient() (RawClient, SMSClientCredentials, error) {
	nexmoClient, nexmoClientCredentials, twilioClient, twilioClientCredentials, customClient, customClientCredentials := r.resolveRawClients()
	provider := r.resolveProvider()

	var client RawClient
	var smsClientCredentials SMSClientCredentials
	switch provider {
	case config.SMSProviderNexmo:
		if nexmoClient == nil {
			return nil, nil, ErrNoAvailableClient
		}
		client = nexmoClient
		smsClientCredentials = nexmoClientCredentials
	case config.SMSProviderTwilio:
		if twilioClient == nil {
			return nil, nil, ErrNoAvailableClient
		}
		client = twilioClient
		smsClientCredentials = twilioClientCredentials
	case config.SMSProviderCustom:
		if customClient == nil {
			return nil, nil, ErrNoAvailableClient
		}
		client = customClient
		smsClientCredentials = customClientCredentials
	default:
		var availableClients []struct {
			RawClient            RawClient
			SMSClientCredentials SMSClientCredentials
		} = []struct {
			RawClient            RawClient
			SMSClientCredentials SMSClientCredentials
		}{}

		if nexmoClient != nil {
			availableClients = append(availableClients, struct {
				RawClient            RawClient
				SMSClientCredentials SMSClientCredentials
			}{
				RawClient:            nexmoClient,
				SMSClientCredentials: nexmoClientCredentials,
			})
		}
		if twilioClient != nil {
			availableClients = append(availableClients, struct {
				RawClient            RawClient
				SMSClientCredentials SMSClientCredentials
			}{
				RawClient:            twilioClient,
				SMSClientCredentials: twilioClientCredentials,
			})
		}
		if customClient != nil {
			availableClients = append(availableClients, struct {
				RawClient            RawClient
				SMSClientCredentials SMSClientCredentials
			}{
				RawClient:            customClient,
				SMSClientCredentials: customClientCredentials,
			})
		}
		if len(availableClients) == 0 {
			return nil, nil, ErrNoAvailableClient
		}
		if len(availableClients) > 1 {
			return nil, nil, ErrAmbiguousClient
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

func (r *ClientResolver) clientsFromAuthgearSecretsYAML() (*NexmoClient, *NexmoClientCredentials, *TwilioClient, *TwilioClientCredentials, *CustomClient, *CustomClientCredentials) {
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
		twilioClientCredentials = &TwilioClientCredentials{
			AccountSID:          r.AuthgearSecretsYAMLTwilioCredentials.AccountSID,
			AuthToken:           r.AuthgearSecretsYAMLTwilioCredentials.AuthToken,
			MessagingServiceSID: r.AuthgearSecretsYAMLTwilioCredentials.MessagingServiceSID,
		}
	}

	if r.AuthgearSecretsYAMLCustomSMSProviderConfig != nil {
		customClientCredentials = &CustomClientCredentials{
			URL:     r.AuthgearSecretsYAMLCustomSMSProviderConfig.URL,
			Timeout: r.AuthgearSecretsYAMLCustomSMSProviderConfig.Timeout,
		}
	}

	return NewNexmoClient(r.AuthgearSecretsYAMLNexmoCredentials), nexmoClientCredentials, NewTwilioClient(r.AuthgearSecretsYAMLTwilioCredentials), twilioClientCredentials, NewCustomClient(r.AuthgearSecretsYAMLCustomSMSProviderConfig, r.SMSDenoHook, r.SMSWebHook), customClientCredentials
}

func (r *ClientResolver) clientsFromEnv() (*NexmoClient, *NexmoClientCredentials, *TwilioClient, *TwilioClientCredentials, *CustomClient, *CustomClientCredentials) {
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
		twilioClientCredentials = &TwilioClientCredentials{
			AccountSID:          r.EnvironmentTwilioCredentials.AccountSID,
			AuthToken:           r.EnvironmentTwilioCredentials.AuthToken,
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

	return NewNexmoClient((*config.NexmoCredentials)(nexmoClientCredentials)), nexmoClientCredentials, NewTwilioClient((*config.TwilioCredentials)(twilioClientCredentials)), twilioClientCredentials, NewCustomClient((*config.CustomSMSProviderConfig)(customClientCredentials), r.SMSDenoHook, r.SMSWebHook), customClientCredentials
}

func (r *ClientResolver) resolveRawClients() (*NexmoClient, *NexmoClientCredentials, *TwilioClient, *TwilioClientCredentials, *CustomClient, *CustomClientCredentials) {
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
func (r *ClientResolver) resolveConfigFromEnv() (*NexmoClient, *NexmoClientCredentials, *TwilioClient, *TwilioClientCredentials, *CustomClient, *CustomClientCredentials) {
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
func (r *ClientResolver) resolveConfigFromAuthgearYAMLAndAuthgearSecretsYAML() (*NexmoClient, *NexmoClientCredentials, *TwilioClient, *TwilioClientCredentials, *CustomClient, *CustomClientCredentials) {
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
