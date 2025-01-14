import { useLocation, useParams } from "react-router-dom";
import { AppSecretKey } from "./globalTypes.generated";
import React, { useState } from "react";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useAppSecretConfigForm } from "../../hook/useAppSecretConfigForm";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import FormContainer from "../../FormContainer";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
  SMSProvider,
  getHookKind,
} from "../../types";
import { produce } from "immer";

const SECRETS = [AppSecretKey.SmsProviderSecrets];

interface LocationState {
  isRevealSecrets: boolean;
}

function isLocationState(raw: unknown): raw is LocationState {
  return (
    raw != null &&
    typeof raw === "object" &&
    (raw as Partial<LocationState>).isRevealSecrets != null
  );
}

enum SMSProviderType {
  Twilio = "twilio",
  Webhook = "webhook",
  Deno = "deno",
}

interface FormState {
  enabled: boolean;
  providerType: SMSProviderType;

  // twilio
  twilioSID: string;
  twilioAuthToken: string | null;
  twilioMessagingServiceSID: string;

  // webhook
  webhookURL: string;
  webhookTimeout: number;

  // deno
  denoHookURL: string;
  denoHookTimeout: number;
}

function constructFormState(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  let enabled: boolean;
  let providerType: SMSProviderType;

  // This implementation only handles the new sms_gateway config and ignores the old sms_provider config
  const isSMSGatewayIsTwilio =
    config.messaging?.sms_gateway != null &&
    config.messaging.sms_gateway.provider === "twilio";
  const hasCustomTwilioCredentials =
    secrets.smsProviderSecrets?.twilioCredentials != null;

  const isSMSGatewayIsCustom =
    config.messaging?.sms_gateway != null &&
    config.messaging.sms_gateway.provider === "custom";
  const hasCustomProviderSecrets =
    secrets.smsProviderSecrets?.customSmsProvider != null;

  if (isSMSGatewayIsTwilio && hasCustomTwilioCredentials) {
    enabled = true;
    providerType = SMSProviderType.Twilio;
  } else if (isSMSGatewayIsCustom && hasCustomProviderSecrets) {
    enabled = true;
    if (
      getHookKind(secrets.smsProviderSecrets!.customSmsProvider!.url) ===
      "denohook"
    ) {
    }
    providerType =
      getHookKind(secrets.smsProviderSecrets!.customSmsProvider!.url) ===
      "denohook"
        ? SMSProviderType.Deno
        : SMSProviderType.Webhook;
  } else {
    enabled = false;
    providerType = SMSProviderType.Twilio;
  }

  const twilioSID =
    secrets.smsProviderSecrets?.twilioCredentials?.accountSid ?? "";
  const twilioAuthToken =
    secrets.smsProviderSecrets?.twilioCredentials?.authToken ?? null;
  const twilioMessagingServiceSID =
    secrets.smsProviderSecrets?.twilioCredentials?.messageServiceSid ?? "";

  let webhookURL = "";
  let webhookTimeout = 30;

  let denoHookURL = "";
  let denoHookTimeout = 30;

  if (secrets.smsProviderSecrets?.customSmsProvider != null) {
    if (
      getHookKind(secrets.smsProviderSecrets.customSmsProvider.url) ===
      "denohook"
    ) {
      denoHookURL = secrets.smsProviderSecrets.customSmsProvider.url;
    } else {
      webhookURL = secrets.smsProviderSecrets.customSmsProvider.url;
    }
    if (secrets.smsProviderSecrets.customSmsProvider.timeout != null) {
      denoHookTimeout = secrets.smsProviderSecrets.customSmsProvider.timeout;
      webhookTimeout = secrets.smsProviderSecrets.customSmsProvider.timeout;
    }
  }
  return {
    enabled,
    providerType,

    twilioSID,
    twilioAuthToken,
    twilioMessagingServiceSID,

    webhookURL,
    webhookTimeout,

    denoHookURL,
    denoHookTimeout,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  const newConfig = produce(config, (config) => {
    config.messaging ??= {};
    if (!currentState.enabled) {
      config.messaging.sms_gateway = undefined;
      config.messaging.sms_provider = undefined;
    } else {
      config.messaging.sms_provider = undefined;

      let newProvider: SMSProvider;
      switch (currentState.providerType) {
        case SMSProviderType.Twilio:
          newProvider = "twilio";
          break;
        case SMSProviderType.Deno:
          newProvider = "custom";
          break;
        case SMSProviderType.Webhook:
          newProvider = "custom";
          break;
      }

      config.messaging.sms_gateway = {
        provider: newProvider,
        use_config_from: "authgear.secrets.yaml",
      };
    }
  });

  const newSecrets = produce(secrets, (secrets) => {
    if (!currentState.enabled) {
      secrets.smsProviderSecrets = null;
    } else {
      switch (currentState.providerType) {
        case SMSProviderType.Twilio:
          secrets.smsProviderSecrets = {
            twilioCredentials: {
              accountSid: currentState.twilioSID,
              authToken: currentState.twilioAuthToken,
              messageServiceSid: currentState.twilioMessagingServiceSID,
            },
          };
          break;
        case SMSProviderType.Webhook:
          secrets.smsProviderSecrets = {
            customSmsProvider: {
              url: currentState.webhookURL,
              timeout: currentState.webhookTimeout,
            },
          };
          break;
        case SMSProviderType.Deno:
          secrets.smsProviderSecrets = {
            customSmsProvider: {
              url: currentState.denoHookURL,
              timeout: currentState.denoHookTimeout,
            },
          };
          break;
      }
    }
  });
  return [newConfig, newSecrets];
}

function constructSecretUpdateInstruction(
  _config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  currentState: FormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  if (!currentState.enabled || !secrets.smsProviderSecrets) {
    return undefined;
  }

  switch (currentState.providerType) {
    case SMSProviderType.Twilio:
      if (secrets.smsProviderSecrets.twilioCredentials == null) {
        console.error("unexpected null twilioCredentials");
        return undefined;
      }
      if (secrets.smsProviderSecrets.twilioCredentials.authToken == null) {
        console.error("unexpected masked twilioCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            twilioCredentials: {
              accountSid:
                secrets.smsProviderSecrets.twilioCredentials.accountSid,
              authToken: secrets.smsProviderSecrets.twilioCredentials.authToken,
              messageServiceSid:
                secrets.smsProviderSecrets.twilioCredentials.messageServiceSid,
            },
          },
        },
      };
    case SMSProviderType.Webhook:
      if (secrets.smsProviderSecrets.customSmsProvider == null) {
        console.error("unexpected null customSmsProvider");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            customSmsProvider: {
              url: secrets.smsProviderSecrets.customSmsProvider.url,
              timeout: secrets.smsProviderSecrets.customSmsProvider.timeout,
            },
          },
        },
      };
    case SMSProviderType.Deno:
      if (secrets.smsProviderSecrets.customSmsProvider == null) {
        console.error("unexpected null customSmsProvider");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            customSmsProvider: {
              url: secrets.smsProviderSecrets.customSmsProvider.url,
              timeout: secrets.smsProviderSecrets.customSmsProvider.timeout,
            },
          },
        },
      };
  }
}

const SMSProviderConfigurationScreen: React.VFC =
  function SMSProviderConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const location = useLocation();
    const [shouldRefreshToken] = useState<boolean>(() => {
      const { state } = location;
      if (isLocationState(state) && state.isRevealSecrets) {
        return true;
      }
      return false;
    });
    useLocationEffect<LocationState>(() => {
      // Pop the location state if exist
    });
    const { token, loading, error, retry } = useAppSecretVisitToken(
      appID,
      SECRETS,
      shouldRefreshToken
    );

    if (error) {
      return <ShowError error={error} onRetry={retry} />;
    }

    if (loading || token === undefined) {
      return <ShowLoading />;
    }

    return (
      <SMSProviderConfigurationScreen1 appID={appID} secretToken={token} />
    );
  };

export default SMSProviderConfigurationScreen;

function SMSProviderConfigurationScreen1({
  appID,
  secretToken,
}: {
  appID: string;
  secretToken: string | null;
}) {
  const form = useAppSecretConfigForm({
    appID,
    secretVisitToken: secretToken,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });
  const featureConfig = useAppFeatureConfigQuery(appID);

  if (form.isLoading || featureConfig.loading) {
    return <ShowLoading />;
  }

  if (form.loadError ?? featureConfig.error) {
    return (
      <ShowError
        error={form.loadError ?? featureConfig.error}
        onRetry={() => {
          form.reload();
          featureConfig.refetch().finally(() => {});
        }}
      />
    );
  }

  return (
    <FormContainer form={form}>
      {/* FIXME */}
      {/* <SMSProviderConfigurationContent form={form} /> */}
    </FormContainer>
  );
}
