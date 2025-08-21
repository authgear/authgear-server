import cn from "classnames";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import {
  AppSecretKey,
  SmsProviderConfigurationInput,
  SmsProviderConfigurationTwilioInput,
  TwilioCredentialType,
} from "./globalTypes.generated";
import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import FormContainer, { FormSaveButton } from "../../FormContainer";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
  SMSProvider,
  SMSProviderTwilioCredentials,
  getHookKind,
} from "../../types";
import { produce } from "immer";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import {
  ChoiceGroup,
  IChoiceGroupOption,
  IChoiceGroupOptionProps,
  IChoiceGroupStyles,
  Text,
} from "@fluentui/react";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import styles from "./SMSProviderConfigurationScreen.module.css";
import Widget from "../../Widget";
import ScreenDescription from "../../ScreenDescription";
import Toggle from "../../Toggle";
import { ProviderCard } from "../../components/common/ProviderCard";
import logoTwilio from "../../images/twilio_logo.svg";
import logoWebhook from "../../images/webhook_logo.svg";
import logoDeno from "../../images/deno_logo.svg";
import FormTextField from "../../FormTextField";
import PrimaryButton from "../../PrimaryButton";
import { startReauthentication } from "./Authenticated";
import { CodeField } from "../../components/common/CodeField";
import TextField from "../../TextField";
import DefaultButton from "../../DefaultButton";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import CodeEditor from "../../CodeEditor";
import { useResourceForm } from "../../hook/useResourceForm";
import {
  Resource,
  ResourceSpecifier,
  ResourcesDiffResult,
  getDenoScriptPathFromURL,
  makeDenoScriptSpecifier,
} from "../../util/resource";
import { DENO_TYPES_URL } from "../../util/deno";
import { genRandomHexadecimalString } from "../../util/random";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import { useSendTestSMSMutation } from "./mutations/sendTestSMS";
import { useCheckDenoHookMutation } from "./mutations/checkDenoHook";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import { ErrorParseRule, makeLocalErrorParseRule } from "../../error/parse";
import { APIError, LocalError } from "../../error/error";
import { ReauthDialog } from "../../components/common/ReauthDialog";
import { TestSMSDialog } from "../../components/sms-provider/TestSMSDialog";
import Tooltip from "../../Tooltip";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { RedMessageBar_RemindConfigureSMSProviderInSMSProviderScreen } from "../../RedMessageBar";

const SECRETS = [AppSecretKey.SmsProviderSecrets, AppSecretKey.WebhookSecret];

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

export type FormModel = Omit<
  AppSecretConfigFormModel<FormState>,
  "saveWithState"
>;

enum SMSProviderType {
  Twilio = "twilio",
  Webhook = "webhook",
  Deno = "deno",
}

enum TwilioSenderType {
  MessagingServiceSID = "MessagingServiceSID",
  From = "From",
}

const MASK = "********";

interface ConfigFormState {
  enabled: boolean;
  providerType: SMSProviderType;
  webhookSecretKey: string | null;

  // twilio
  twilioCredentialType: TwilioCredentialType;
  twilioSID: string;
  twilioAuthToken: string | null;
  twilioAPIKeySID: string;
  twilioAPIKeySecret: string | null;
  twilioSenderType: TwilioSenderType;
  twilioMessagingServiceSID: string;
  twilioFrom: string;

  // webhook
  webhookURL: string;
  webhookTimeout: number;

  // deno
  denoHookURL: string;
  denoHookTimeout: number;
}

interface FormState extends ConfigFormState {
  resources: Resource[];
  diff: ResourcesDiffResult | null;

  isSMSRequiredForSomeEnabledFeatures: boolean;
  smsProviderConfigured: boolean;
}

function constructFormState(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): ConfigFormState {
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
    secrets.smsProviderSecrets?.customSMSProviderCredentials != null;

  if (isSMSGatewayIsTwilio && hasCustomTwilioCredentials) {
    enabled = true;
    providerType = SMSProviderType.Twilio;
  } else if (isSMSGatewayIsCustom && hasCustomProviderSecrets) {
    enabled = true;
    if (
      getHookKind(
        secrets.smsProviderSecrets!.customSMSProviderCredentials!.url
      ) === "denohook"
    ) {
    }
    providerType =
      getHookKind(
        secrets.smsProviderSecrets!.customSMSProviderCredentials!.url
      ) === "denohook"
        ? SMSProviderType.Deno
        : SMSProviderType.Webhook;
  } else {
    enabled = false;
    providerType = SMSProviderType.Twilio;
  }

  let twilioCredentialType: TwilioCredentialType = TwilioCredentialType.ApiKey;
  let twilioSID = "";
  let twilioAPIKeySID = "";
  let twilioAuthToken: string | null = "";
  let twilioAPIKeySecret: string | null = "";
  let twilioSenderType: TwilioSenderType = TwilioSenderType.From;
  let twilioMessagingServiceSID = "";
  let twilioFrom = "";

  if (enabled && providerType === SMSProviderType.Twilio) {
    twilioSID = secrets.smsProviderSecrets?.twilioCredentials?.accountSID ?? "";
    twilioCredentialType =
      secrets.smsProviderSecrets?.twilioCredentials?.credentialType ??
      TwilioCredentialType.AuthToken;
    switch (twilioCredentialType) {
      case TwilioCredentialType.AuthToken:
        twilioAuthToken =
          secrets.smsProviderSecrets?.twilioCredentials != null
            ? secrets.smsProviderSecrets.twilioCredentials.authToken ?? null
            : "";
        break;
      case TwilioCredentialType.ApiKey:
        twilioAPIKeySID =
          secrets.smsProviderSecrets?.twilioCredentials?.apiKeySID ?? "";
        twilioAPIKeySecret =
          secrets.smsProviderSecrets?.twilioCredentials != null
            ? secrets.smsProviderSecrets.twilioCredentials.apiKeySecret ?? null
            : "";
    }

    if (secrets.smsProviderSecrets?.twilioCredentials?.messagingServiceSID) {
      twilioSenderType = TwilioSenderType.MessagingServiceSID;
      twilioMessagingServiceSID =
        secrets.smsProviderSecrets.twilioCredentials.messagingServiceSID;
    } else if (secrets.smsProviderSecrets?.twilioCredentials?.from) {
      twilioSenderType = TwilioSenderType.From;
      twilioFrom = secrets.smsProviderSecrets.twilioCredentials.from;
    }
  }

  let webhookURL = "";
  let webhookTimeout = 30;

  let denoHookURL = "";
  let denoHookTimeout = 30;

  if (
    enabled &&
    (providerType === SMSProviderType.Webhook ||
      providerType === SMSProviderType.Deno) &&
    secrets.smsProviderSecrets?.customSMSProviderCredentials != null
  ) {
    if (
      getHookKind(
        secrets.smsProviderSecrets.customSMSProviderCredentials.url
      ) === "denohook"
    ) {
      denoHookURL = secrets.smsProviderSecrets.customSMSProviderCredentials.url;
    } else {
      webhookURL = secrets.smsProviderSecrets.customSMSProviderCredentials.url;
    }
    if (
      secrets.smsProviderSecrets.customSMSProviderCredentials.timeout != null
    ) {
      denoHookTimeout =
        secrets.smsProviderSecrets.customSMSProviderCredentials.timeout;
      webhookTimeout =
        secrets.smsProviderSecrets.customSMSProviderCredentials.timeout;
    }
  }
  return {
    enabled,
    providerType,
    webhookSecretKey: secrets.webhookSecret?.secret ?? null,

    twilioCredentialType,
    twilioSID,
    twilioAuthToken,
    twilioAPIKeySID,
    twilioAPIKeySecret,
    twilioSenderType,
    twilioMessagingServiceSID,
    twilioFrom,

    webhookURL,
    webhookTimeout,

    denoHookURL,
    denoHookTimeout,
  } satisfies ConfigFormState;
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _initialState: ConfigFormState,
  currentState: ConfigFormState,
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
        case SMSProviderType.Twilio: {
          const twilioCredentials: SMSProviderTwilioCredentials = {
            credentialType: currentState.twilioCredentialType,
            accountSID: currentState.twilioSID,
          };
          switch (currentState.twilioCredentialType) {
            case TwilioCredentialType.ApiKey:
              twilioCredentials.apiKeySID = currentState.twilioAPIKeySID;
              twilioCredentials.apiKeySecret = currentState.twilioAPIKeySecret;
              break;
            case TwilioCredentialType.AuthToken:
              twilioCredentials.authToken = currentState.twilioAuthToken;
              break;
          }
          switch (currentState.twilioSenderType) {
            case TwilioSenderType.From:
              twilioCredentials.from = currentState.twilioFrom;
              break;
            case TwilioSenderType.MessagingServiceSID:
              twilioCredentials.messagingServiceSID =
                currentState.twilioMessagingServiceSID;
              break;
          }
          secrets.smsProviderSecrets = { twilioCredentials: twilioCredentials };
          break;
        }
        case SMSProviderType.Webhook:
          secrets.smsProviderSecrets = {
            customSMSProviderCredentials: {
              url: currentState.webhookURL,
              timeout: currentState.webhookTimeout,
            },
          };
          break;
        case SMSProviderType.Deno:
          secrets.smsProviderSecrets = {
            customSMSProviderCredentials: {
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
  currentState: ConfigFormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  if (!currentState.enabled || !secrets.smsProviderSecrets) {
    // Remove all existing secrets
    return {
      smsProviderSecrets: {
        action: "set",
        setData: {},
      },
    };
  }

  switch (currentState.providerType) {
    case SMSProviderType.Twilio:
      if (secrets.smsProviderSecrets.twilioCredentials == null) {
        console.error("unexpected null twilioCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            twilioCredentials: {
              credentialType:
                secrets.smsProviderSecrets.twilioCredentials.credentialType,
              accountSID:
                secrets.smsProviderSecrets.twilioCredentials.accountSID,
              authToken: secrets.smsProviderSecrets.twilioCredentials.authToken,
              apiKeySID: secrets.smsProviderSecrets.twilioCredentials.apiKeySID,
              apiKeySecret:
                secrets.smsProviderSecrets.twilioCredentials.apiKeySecret,
              messagingServiceSID:
                secrets.smsProviderSecrets.twilioCredentials
                  .messagingServiceSID,
              from: secrets.smsProviderSecrets.twilioCredentials.from,
            },
          },
        },
      };
    case SMSProviderType.Webhook:
      if (secrets.smsProviderSecrets.customSMSProviderCredentials == null) {
        console.error("unexpected null customSMSProviderCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            customSMSProviderCredentials: {
              url: secrets.smsProviderSecrets.customSMSProviderCredentials.url,
              timeout:
                secrets.smsProviderSecrets.customSMSProviderCredentials.timeout,
            },
          },
        },
      };
    case SMSProviderType.Deno:
      if (secrets.smsProviderSecrets.customSMSProviderCredentials == null) {
        console.error("unexpected null customSMSProviderCredentials");
        return undefined;
      }
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {
            customSMSProviderCredentials: {
              url: secrets.smsProviderSecrets.customSMSProviderCredentials.url,
              timeout:
                secrets.smsProviderSecrets.customSMSProviderCredentials.timeout,
            },
          },
        },
      };
  }
}

const localErrorFromRequired: LocalError = {
  errorName: "__local",
  reason: "__local",
  info: {
    error: {
      messageID: "errors.validation.required",
    },
  },
};

const fromErrorRules: ErrorParseRule[] = [
  makeLocalErrorParseRule(
    localErrorFromRequired,
    localErrorFromRequired.info.error
  ),
];

const localErrorMessagingServiceSIDRequired: LocalError = {
  errorName: "__local",
  reason: "__local",
  info: {
    error: {
      messageID: "errors.validation.required",
    },
  },
};

const messagingServiceSIDErrorRules: ErrorParseRule[] = [
  makeLocalErrorParseRule(
    localErrorMessagingServiceSIDRequired,
    localErrorMessagingServiceSIDRequired.info.error
  ),
];

function makeSpecifiersFromState(state: ConfigFormState): ResourceSpecifier[] {
  const specifiers = [];
  if (state.denoHookURL) {
    specifiers.push(makeDenoScriptSpecifier(state.denoHookURL));
  }
  return specifiers;
}

function makeNewDenoScriptURL(): string {
  const rand = genRandomHexadecimalString();
  return `authgeardeno:///deno/sms.${rand}.ts`;
}

const DEFAULT_SMS_SCRIPT_TEMPLATE = `// This custom script will be executed when a message is triggered
// Sample script:
import { CustomSMSGatewayPayload, CustomSMSGatewayResponse } from "${DENO_TYPES_URL}";

export default async function (e: CustomSMSGatewayPayload): Promise<CustomSMSGatewayResponse> {
  const body = JSON.stringify(e);
  const response = await fetch("https://some.sms.gateway", {
    method: "POST",
    body: body,
  });

  if (!response.ok) {
    return {
      code: "delivery_rejected",
    }
  }

  return {
    code: "ok",
  }
}
`;

const CODE_EDITOR_OPTIONS = {
  minimap: {
    enabled: false,
  },
};

function useDenoScriptResourceIndex(state: FormState) {
  const resourceIdx = useMemo(() => {
    if (state.denoHookURL === "") {
      return -1;
    }
    const path = getDenoScriptPathFromURL(state.denoHookURL);
    for (const [idx, r] of state.resources.entries()) {
      if (r.path === path && r.nullableValue != null) {
        return idx;
      }
    }
    return -1;
  }, [state.denoHookURL, state.resources]);
  return resourceIdx;
}

function useTestSMSConfig(
  state: FormState
): SmsProviderConfigurationInput | null {
  const denoResourceIdx = useDenoScriptResourceIndex(state);

  return useMemo((): SmsProviderConfigurationInput | null => {
    if (!state.enabled) {
      return null;
    }
    switch (state.providerType) {
      case SMSProviderType.Twilio: {
        if (!state.twilioSID) {
          return null;
        }
        const twilio: SmsProviderConfigurationTwilioInput = {
          credentialType: state.twilioCredentialType,
          accountSID: state.twilioSID,
          authToken: state.twilioAuthToken ?? "",
          apiKeySID: state.twilioAPIKeySID,
          apiKeySecret: state.twilioAPIKeySecret ?? "",
        };
        switch (state.twilioCredentialType) {
          case TwilioCredentialType.ApiKey:
            twilio.apiKeySID = state.twilioAPIKeySID;
            twilio.apiKeySecret = state.twilioAPIKeySecret;
            break;
          case TwilioCredentialType.AuthToken:
            twilio.authToken = state.twilioAuthToken;
            break;
        }
        switch (state.twilioSenderType) {
          case TwilioSenderType.From:
            twilio.from = state.twilioFrom;
            break;
          case TwilioSenderType.MessagingServiceSID:
            twilio.messagingServiceSID = state.twilioMessagingServiceSID;
            break;
        }
        return {
          twilio: {
            credentialType: state.twilioCredentialType,
            accountSID: state.twilioSID,
            authToken: state.twilioAuthToken ?? "",
            apiKeySID: state.twilioAPIKeySID,
            apiKeySecret: state.twilioAPIKeySecret ?? "",
            messagingServiceSID: state.twilioMessagingServiceSID,
            from: state.twilioFrom,
          },
        };
      }
      case SMSProviderType.Webhook:
        if (!state.webhookURL) {
          return null;
        }
        return {
          webhook: {
            url: state.webhookURL,
            timeout: state.webhookTimeout,
          },
        };
      case SMSProviderType.Deno: {
        if (denoResourceIdx === -1) {
          return null;
        }
        const script = state.resources[denoResourceIdx].nullableValue ?? "";
        if (!script) {
          return null;
        }
        return {
          deno: {
            script: script,
            timeout: state.denoHookTimeout,
          },
        };
      }
    }
  }, [
    denoResourceIdx,
    state.denoHookTimeout,
    state.enabled,
    state.providerType,
    state.resources,
    state.twilioAPIKeySID,
    state.twilioAPIKeySecret,
    state.twilioAuthToken,
    state.twilioCredentialType,
    state.twilioFrom,
    state.twilioMessagingServiceSID,
    state.twilioSID,
    state.twilioSenderType,
    state.webhookTimeout,
    state.webhookURL,
  ]);
}

function computeIsSecretMasked(state: FormState): boolean {
  if (!state.enabled) {
    return false;
  }
  switch (state.providerType) {
    case SMSProviderType.Twilio:
      switch (state.twilioCredentialType) {
        case TwilioCredentialType.ApiKey:
          return state.twilioAPIKeySecret == null;
        case TwilioCredentialType.AuthToken:
          return state.twilioAuthToken == null;
      }
      throw new Error("unreachable code");
    case SMSProviderType.Webhook:
      return state.webhookSecretKey == null;
    case SMSProviderType.Deno:
      return false;
  }
  throw new Error("unreachable code");
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
  const {
    effectiveAppConfig,
    isLoading: loadingAppConfig,
    loadError: appConfigError,
    refetch: refetchAppConfig,
    secretConfig,
  } = useAppAndSecretConfigQuery(appID, secretToken);
  const configForm = useAppSecretConfigForm({
    appID,
    secretVisitToken: secretToken,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });
  const featureConfig = useAppFeatureConfigQuery(appID);
  const specifiers = useMemo(() => {
    return makeSpecifiersFromState(configForm.state);
  }, [configForm.state]);
  const resources = useResourceForm(
    appID,
    specifiers,
    (resources) => resources,
    (resources) => resources
  );
  const sendTestSMSHandle = useSendTestSMSMutation(appID);
  const checkDenoHookHandle = useCheckDenoHookMutation(appID);

  const [localError, setLocalError] = useState<APIError | null>(null);

  const state = useMemo<FormState>(() => {
    return {
      ...configForm.state,
      resources: resources.state,
      diff: resources.diff,

      isSMSRequiredForSomeEnabledFeatures:
        // primary authentication uses SMS.
        effectiveAppConfig?.authentication?.primary_authenticators?.includes(
          "oob_otp_sms"
        ) === true ||
        // secondary authenticatoin uses SMS AND secondary authentication is enabled.
        (effectiveAppConfig?.authentication?.secondary_authenticators?.includes(
          "oob_otp_sms"
        ) === true &&
          (effectiveAppConfig.authentication.secondary_authentication_mode ===
            "if_exists" ||
            effectiveAppConfig.authentication.secondary_authentication_mode ===
              "required")) ||
        // phone verification enabled.
        effectiveAppConfig?.verification?.claims?.phone_number?.enabled ===
          true,

      smsProviderConfigured:
        secretConfig?.smsProviderSecrets?.twilioCredentials != null ||
        secretConfig?.smsProviderSecrets?.customSMSProviderCredentials != null,
    };
  }, [
    configForm.state,
    resources.state,
    resources.diff,
    effectiveAppConfig?.authentication?.primary_authenticators,
    effectiveAppConfig?.authentication?.secondary_authenticators,
    effectiveAppConfig?.authentication?.secondary_authentication_mode,
    effectiveAppConfig?.verification?.claims?.phone_number?.enabled,
    secretConfig?.smsProviderSecrets?.twilioCredentials,
    secretConfig?.smsProviderSecrets?.customSMSProviderCredentials,
  ]);

  const form: FormModel = {
    isLoading: configForm.isLoading || resources.isLoading,
    isUpdating: configForm.isUpdating || resources.isUpdating,
    isDirty: configForm.isDirty || resources.isDirty,
    loadError: configForm.loadError ?? resources.loadError,
    updateError: configForm.updateError ?? resources.updateError,
    state,
    setState: (fn) => {
      const newState = fn(state);
      const { resources: newResources, ...configState } = newState;
      configForm.setState(() => ({
        ...configState,
      }));
      resources.setState(() => newResources);
    },
    reload: () => {
      resources.reload();
      configForm.reload();
    },
    reset: () => {
      resources.reset();
      configForm.reset();
    },
    save: async (ignoreConflict: boolean = false) => {
      await resources.save(ignoreConflict);
      await configForm.save(ignoreConflict);
    },
  };

  const validateForm = useCallback(async () => {
    setLocalError(null);
    if (!form.state.enabled) {
      return;
    }
    switch (form.state.providerType) {
      case SMSProviderType.Twilio:
        switch (form.state.twilioSenderType) {
          case TwilioSenderType.From:
            if (!form.state.twilioFrom) {
              setLocalError(localErrorFromRequired);
              throw new Error("twilioFrom is required");
            }
            break;
          case TwilioSenderType.MessagingServiceSID:
            if (!form.state.twilioMessagingServiceSID) {
              setLocalError(localErrorMessagingServiceSIDRequired);
              throw new Error("twilioMessagingServiceSID is required");
            }
            break;
        }
        break;
      default:
        break;
    }
  }, [
    form.state.enabled,
    form.state.providerType,
    form.state.twilioFrom,
    form.state.twilioMessagingServiceSID,
    form.state.twilioSenderType,
  ]);

  if (loadingAppConfig || form.isLoading || featureConfig.isLoading) {
    return <ShowLoading />;
  }

  if (appConfigError ?? form.loadError ?? featureConfig.loadError) {
    return (
      <ShowError
        error={form.loadError ?? featureConfig.loadError}
        onRetry={() => {
          refetchAppConfig().finally(() => {});
          form.reload();
          featureConfig.refetch().finally(() => {});
        }}
      />
    );
  }

  return (
    <FormContainer
      form={form}
      beforeSave={validateForm}
      hideFooterComponent={true}
      localError={
        checkDenoHookHandle.error ?? sendTestSMSHandle.error ?? localError
      }
    >
      <SMSProviderConfigurationContent
        form={form}
        effectiveAppConfig={effectiveAppConfig ?? undefined}
        sendTestSMSHandle={sendTestSMSHandle}
        checkDenoHookHandle={checkDenoHookHandle}
        isCustomSMSProviderDisabled={
          featureConfig.effectiveFeatureConfig?.messaging
            ?.custom_sms_provider_disabled ?? false
        }
      />
    </FormContainer>
  );
}

function SMSProviderConfigurationContent(props: {
  form: FormModel;
  effectiveAppConfig: PortalAPIAppConfig | undefined;
  sendTestSMSHandle: ReturnType<typeof useSendTestSMSMutation>;
  checkDenoHookHandle: ReturnType<typeof useCheckDenoHookMutation>;
  isCustomSMSProviderDisabled: boolean;
}) {
  const {
    form,
    effectiveAppConfig,
    checkDenoHookHandle,
    isCustomSMSProviderDisabled,
  } = props;
  const { isAuthgearOnce } = useSystemConfig();
  const { appID } = useParams() as { appID: string };
  const { state, setState } = form;
  const { isSMSRequiredForSomeEnabledFeatures, smsProviderConfigured } = state;
  const { renderToString } = useContext(MessageContext);
  const navigate = useNavigate();

  const [isReauthDialogHidden, setIsReauthDialogHidden] = useState(true);
  const [isTestSMSDialogHidden, setIsTestSMSDialogHidden] = useState(true);

  const { checkDenoHook, loading: checkDenoHookLoading } = checkDenoHookHandle;

  const isSecretMasked = useMemo(
    () => computeIsSecretMasked(form.state),
    [form.state]
  );

  const onChangeEnabled = useCallback(
    (_event, checked?: boolean) => {
      if (checked != null) {
        if (isSecretMasked) {
          setIsReauthDialogHidden(false);
          return;
        }
        setState((state) => {
          return {
            ...state,
            enabled: checked,
          };
        });
      }
    },
    [isSecretMasked, setState]
  );

  const triggerReauth = useCallback(() => {
    // We are going to leave, reset the form so that the confirmation dialog won't appear
    form.reset();

    startReauthentication<LocationState>(navigate, {
      isRevealSecrets: true,
    }).catch((e) => {
      // Normally there should not be any error.
      console.error(e);
    });
  }, [navigate, form]);

  const onRevealSecrets = useCallback(() => {
    setIsReauthDialogHidden(false);
  }, []);

  const testConfig = useTestSMSConfig(form.state);

  const onTestSMS = useCallback(async () => {
    if (isSecretMasked) {
      setIsReauthDialogHidden(false);
      return;
    }
    if (form.state.providerType === SMSProviderType.Deno) {
      await checkDenoHook(testConfig?.deno?.script ?? "");
    }
    setIsTestSMSDialogHidden(false);
  }, [
    checkDenoHook,
    form.state.providerType,
    isSecretMasked,
    testConfig?.deno?.script,
  ]);

  const onCancelTestSMS = useCallback(() => {
    setIsTestSMSDialogHidden(true);
  }, []);

  return (
    <>
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          {isAuthgearOnce ? (
            <FormattedMessage id="SMSProviderConfigurationScreen.title--authgearonce" />
          ) : (
            <FormattedMessage id="SMSProviderConfigurationScreen.title" />
          )}
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="SMSProviderConfigurationScreen.description" />
        </ScreenDescription>
        {isCustomSMSProviderDisabled ? (
          <FeatureDisabledMessageBar
            className={styles.widget}
            messageID="FeatureConfig.custom-sms-provider.disabled"
          />
        ) : null}
        {isAuthgearOnce &&
        isSMSRequiredForSomeEnabledFeatures &&
        !smsProviderConfigured ? (
          <div className={cn(styles.widget, "flex flex-col")}>
            <RedMessageBar_RemindConfigureSMSProviderInSMSProviderScreen className="self-start w-fit" />
          </div>
        ) : null}
        <Widget className={styles.widget} contentLayout="grid">
          <Toggle
            className={styles.columnFull}
            disabled={isCustomSMSProviderDisabled}
            checked={state.enabled}
            onChange={onChangeEnabled}
            label={renderToString(
              isAuthgearOnce
                ? "SMSProviderConfigurationScreen.enable.label--authgearonce"
                : "SMSProviderConfigurationScreen.enable.label"
            )}
            inlineLabel={true}
          />
        </Widget>

        {state.enabled ? (
          <Widget className={cn(styles.widget, "flex flex-col gap-y-4")}>
            <ProviderSection
              form={form}
              isSecretMasked={isSecretMasked}
              onRevealSecrets={onRevealSecrets}
            />
            <FormSection form={form} onRevealSecrets={onRevealSecrets} />
          </Widget>
        ) : null}

        {form.state.enabled ? (
          <>
            <div className={cn(styles.widget, "flex w-max pt-1 gap-4")}>
              <DefaultButton
                className="w-max"
                text={
                  <FormattedMessage id="SMSProviderConfigurationScreen.testSMS" />
                }
                disabled={testConfig == null || checkDenoHookLoading}
                onClick={onTestSMS}
              />
              {isSecretMasked ? (
                <PrimaryButton
                  className="w-max"
                  onClick={onRevealSecrets}
                  text={<FormattedMessage id="edit" />}
                />
              ) : (
                <FormSaveButton />
              )}
            </div>
          </>
        ) : (
          <div className={styles.widget}>
            <FormSaveButton />
          </div>
        )}
      </ScreenContent>
      <ReauthDialog
        isHidden={isReauthDialogHidden}
        onConfirm={useCallback(() => {
          triggerReauth();
        }, [triggerReauth])}
        onCancel={useCallback(() => {
          setIsReauthDialogHidden(true);
        }, [])}
      />
      {testConfig != null ? (
        <TestSMSDialog
          appID={appID}
          isHidden={isTestSMSDialogHidden}
          effectiveAppConfig={effectiveAppConfig}
          input={testConfig}
          onDismiss={onCancelTestSMS}
        />
      ) : null}
    </>
  );
}

function ProviderSection({
  isSecretMasked,
  form,
  onRevealSecrets,
}: {
  isSecretMasked: boolean;
  form: FormModel;
  onRevealSecrets: () => void;
}) {
  const onSelectProviderCallbacks = useMemo(() => {
    const makeCallback = (provider: SMSProviderType) => {
      return () => {
        if (isSecretMasked) {
          onRevealSecrets();
          return;
        }
        form.setState((state) => {
          return { ...state, providerType: provider };
        });
      };
    };

    return {
      twilio: makeCallback(SMSProviderType.Twilio),
      webhook: makeCallback(SMSProviderType.Webhook),
      deno: makeCallback(SMSProviderType.Deno),
    };
  }, [form, isSecretMasked, onRevealSecrets]);

  return (
    <div className="flex flex-col gap-y-3">
      <Text variant="xLarge">
        <FormattedMessage id="SMSProviderConfigurationScreen.provider.title" />
      </Text>
      <div className={styles.providerGrid}>
        <ProviderCard
          onClick={onSelectProviderCallbacks.twilio}
          isSelected={form.state.providerType === SMSProviderType.Twilio}
          logoSrc={logoTwilio}
        >
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.twilio" />
        </ProviderCard>
        <ProviderCard
          onClick={onSelectProviderCallbacks.webhook}
          isSelected={form.state.providerType === SMSProviderType.Webhook}
          logoSrc={logoWebhook}
        >
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.webhook" />
        </ProviderCard>
        <ProviderCard
          onClick={onSelectProviderCallbacks.deno}
          isSelected={form.state.providerType === SMSProviderType.Deno}
          logoSrc={logoDeno}
        >
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.deno" />
        </ProviderCard>
      </div>
      <Text block={true}>
        {form.state.providerType === SMSProviderType.Twilio ? (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.twilio.description"
            values={{
              href: "https://docs.authgear.com/customization/custom-providers/twilio",
            }}
          />
        ) : form.state.providerType === SMSProviderType.Webhook ? (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.webhook.description"
            values={{
              href: "https://docs.authgear.com/customization/custom-providers/webhook-custom-script",
            }}
          />
        ) : (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.deno.description"
            values={{
              href: "https://docs.authgear.com/customization/custom-providers/webhook-custom-script",
            }}
          />
        )}
      </Text>
    </div>
  );
}

function FormSection({
  form,
  onRevealSecrets,
}: {
  form: FormModel;
  onRevealSecrets: () => void;
}) {
  switch (form.state.providerType) {
    case SMSProviderType.Twilio:
      return <TwilioForm form={form} />;
    case SMSProviderType.Webhook:
      return <WebhookForm form={form} onRevealSecrets={onRevealSecrets} />;
    case SMSProviderType.Deno:
      return <DenoHookForm form={form} />;
  }
}

function TwilioForm({ form }: { form: FormModel }) {
  const { renderToString } = useContext(MessageContext);

  const onStringChangeCallbacks = useMemo(() => {
    const callbackFactory = (
      key:
        | "twilioSID"
        | "twilioAuthToken"
        | "twilioAPIKeySID"
        | "twilioAPIKeySecret"
        | "twilioMessagingServiceSID"
        | "twilioFrom"
    ) => {
      return (
        event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>
      ) => {
        form.setState((prevState) => {
          const value = event.currentTarget.value;
          const s: FormState = {
            ...prevState,
          };
          s[key] = value;
          return s;
        });
      };
    };
    return {
      twilioSID: callbackFactory("twilioSID"),
      twilioAuthToken: callbackFactory("twilioAuthToken"),
      twilioAPIKeySID: callbackFactory("twilioAPIKeySID"),
      twilioAPIKeySecret: callbackFactory("twilioAPIKeySecret"),
      twilioMessagingServiceSID: callbackFactory("twilioMessagingServiceSID"),
      twilioFrom: callbackFactory("twilioFrom"),
    };
  }, [form]);

  const isTwilioSecretMasked =
    form.state.twilioCredentialType === TwilioCredentialType.AuthToken
      ? form.state.twilioAuthToken == null
      : form.state.twilioAPIKeySecret == null;

  const credentialTypeOptions = useMemo<IChoiceGroupOption[]>(() => {
    return [
      {
        key: TwilioCredentialType.AuthToken,
        text: renderToString(
          "SMSProviderConfigurationScreen.form.twilio.credentialType.options.authToken"
        ),
      },
      {
        key: TwilioCredentialType.ApiKey,
        text: renderToString(
          "SMSProviderConfigurationScreen.form.twilio.credentialType.options.apiKey"
        ),
      },
    ];
  }, [renderToString]);

  const onCredentialTypeChange = useCallback(
    (_: unknown, option?: IChoiceGroupOption) => {
      if (option == null) {
        return;
      }
      form.setState((prev) => {
        return {
          ...prev,
          twilioCredentialType: option.key as TwilioCredentialType,
        };
      });
    },
    [form]
  );

  const senderOptions = useMemo<IChoiceGroupOption[]>(() => {
    return [
      {
        key: TwilioSenderType.MessagingServiceSID,
        text: renderToString(
          "SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID"
        ),
        // eslint-disable-next-line react/no-unstable-nested-components
        onRenderLabel: (
          props?: IChoiceGroupOption & IChoiceGroupOptionProps,
          render?: (
            props?: IChoiceGroupOption & IChoiceGroupOptionProps
          ) => JSX.Element | null
        ) => (
          <>
            {render?.(props)}
            <div className="inline-flex">
              <Tooltip tooltipMessageId="SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID.tooltip" />
            </div>
          </>
        ),
      },
      {
        key: TwilioSenderType.From,
        text: renderToString(
          "SMSProviderConfigurationScreen.form.twilio.twilioFrom"
        ), // eslint-disable-next-line react/no-unstable-nested-components
        onRenderLabel: (
          props?: IChoiceGroupOption & IChoiceGroupOptionProps,
          render?: (
            props?: IChoiceGroupOption & IChoiceGroupOptionProps
          ) => JSX.Element | null
        ) => (
          <>
            {render?.(props)}
            <div className="inline-flex">
              <Tooltip tooltipMessageId="SMSProviderConfigurationScreen.form.twilio.twilioFrom.tooltip" />
            </div>
          </>
        ),
      },
    ];
  }, [renderToString]);

  const onSenderTypeChange = useCallback(
    (_: unknown, option?: IChoiceGroupOption) => {
      if (option == null) {
        return;
      }
      form.setState((prev) => {
        return {
          ...prev,
          twilioSenderType: option.key as TwilioSenderType,
        };
      });
    },
    [form]
  );

  const horizontalChoiceGroupStyles: Partial<IChoiceGroupStyles> = useMemo(
    () => ({
      flexContainer: {
        display: "flex",
        columnGap: "16px",
      },
    }),
    []
  );

  return (
    <div className="flex flex-col gap-y-4">
      <FormTextField
        type="text"
        label={renderToString(
          "SMSProviderConfigurationScreen.form.twilio.twilioSID"
        )}
        value={form.state.twilioSID}
        required={true}
        onChange={onStringChangeCallbacks.twilioSID}
        disabled={isTwilioSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="account_sid"
      />
      <div className="flex flex-col gap-3">
        <ChoiceGroup
          selectedKey={form.state.twilioSenderType}
          options={senderOptions}
          onChange={onSenderTypeChange}
          disabled={isTwilioSecretMasked}
          label={renderToString(
            "SMSProviderConfigurationScreen.form.twilio.sender"
          )}
          required={true}
          styles={horizontalChoiceGroupStyles}
        />
        {form.state.twilioSenderType ===
        TwilioSenderType.MessagingServiceSID ? (
          <div className="flex flex-col gap-2">
            <FormTextField
              type="text"
              placeholder={renderToString(
                "SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID.placeholder"
              )}
              value={form.state.twilioMessagingServiceSID}
              onChange={onStringChangeCallbacks.twilioMessagingServiceSID}
              disabled={isTwilioSecretMasked}
              errorRules={messagingServiceSIDErrorRules}
              parentJSONPointer={/\/secrets\/\d+\/data/}
              fieldName="message_service_sid"
            />
            <Text>
              <FormattedMessage
                id="SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID.hint"
                values={{
                  href: "https://www.twilio.com/docs/messaging/services",
                }}
              />
            </Text>
          </div>
        ) : (
          <FormTextField
            type="text"
            placeholder={renderToString(
              "SMSProviderConfigurationScreen.form.twilio.twilioFrom.placeholder"
            )}
            value={form.state.twilioFrom}
            onChange={onStringChangeCallbacks.twilioFrom}
            disabled={isTwilioSecretMasked}
            parentJSONPointer={/\/secrets\/\d+\/data/}
            fieldName="from"
            errorRules={fromErrorRules}
          />
        )}
      </div>
      <div className="flex flex-col gap-3">
        <ChoiceGroup
          label={renderToString(
            "SMSProviderConfigurationScreen.form.twilio.credentialType"
          )}
          required={true}
          options={credentialTypeOptions}
          selectedKey={form.state.twilioCredentialType}
          onChange={onCredentialTypeChange}
          styles={horizontalChoiceGroupStyles}
          disabled={isTwilioSecretMasked}
        />
        {form.state.twilioCredentialType === TwilioCredentialType.AuthToken ? (
          <FormTextField
            type="text"
            placeholder={renderToString(
              "SMSProviderConfigurationScreen.form.twilio.credentialType.options.authToken"
            )}
            value={form.state.twilioAuthToken ?? MASK}
            onChange={onStringChangeCallbacks.twilioAuthToken}
            disabled={isTwilioSecretMasked}
            parentJSONPointer={/\/secrets\/\d+\/data/}
            fieldName="auth_token"
          />
        ) : null}
        {form.state.twilioCredentialType === TwilioCredentialType.ApiKey ? (
          <>
            <FormTextField
              type="text"
              label={renderToString(
                "SMSProviderConfigurationScreen.form.twilio.apiKeySID"
              )}
              value={form.state.twilioAPIKeySID}
              onChange={onStringChangeCallbacks.twilioAPIKeySID}
              disabled={isTwilioSecretMasked}
              parentJSONPointer={/\/secrets\/\d+\/data/}
              fieldName="api_key_sid"
            />
            <FormTextField
              type="text"
              label={renderToString(
                "SMSProviderConfigurationScreen.form.twilio.apiKeySecret"
              )}
              value={form.state.twilioAPIKeySecret ?? MASK}
              onChange={onStringChangeCallbacks.twilioAPIKeySecret}
              disabled={isTwilioSecretMasked}
              parentJSONPointer={/\/secrets\/\d+\/data/}
              fieldName="api_key_secret"
            />
          </>
        ) : null}
      </div>
    </div>
  );
}

function WebhookForm({
  form,
  onRevealSecrets,
}: {
  form: FormModel;
  onRevealSecrets: () => void;
}) {
  const { renderToString } = useContext(MessageContext);

  const onURLChange = useCallback(
    (event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      form.setState((prevState) => {
        const value = event.currentTarget.value;
        return {
          ...prevState,
          webhookURL: value,
        } satisfies FormState;
      });
    },
    [form]
  );

  const onTimeoutChange = useCallback(
    (event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const value = parseInt(event.currentTarget.value, 10);
      if (isNaN(value)) {
        return;
      }
      form.setState((prevState) => {
        return {
          ...prevState,
          webhookTimeout: value,
        } satisfies FormState;
      });
    },
    [form]
  );

  const { copyButtonProps, Feedback: CopyFeedbackComponent } = useCopyFeedback({
    textToCopy: form.state.webhookSecretKey ?? "",
  });

  const isWebhookSecretMasked = form.state.webhookSecretKey == null;

  return (
    <div className="flex flex-col gap-y-4">
      <FormTextField
        type="text"
        label={renderToString(
          "SMSProviderConfigurationScreen.form.webhook.url"
        )}
        value={form.state.webhookURL}
        required={true}
        onChange={onURLChange}
        disabled={isWebhookSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="url"
      />
      <CodeField
        label={renderToString(
          "SMSProviderConfigurationScreen.form.webhook.payload"
        )}
        description={renderToString(
          "SMSProviderConfigurationScreen.form.webhook.payload.description"
        )}
      >
        {`{
  "to": "+85298765432",
  "body": "You OTP is 123456"
}`}
      </CodeField>
      <div>
        <div className="flex items-end gap-x-2">
          <TextField
            className="flex-1"
            type="text"
            label={renderToString(
              "SMSProviderConfigurationScreen.form.webhook.signatureKey"
            )}
            value={
              isWebhookSecretMasked ? MASK : form.state.webhookSecretKey ?? ""
            }
            readOnly={true}
          />
          <DefaultButton
            className={styles.secretButton}
            id={copyButtonProps.id}
            onClick={
              !isWebhookSecretMasked ? copyButtonProps.onClick : onRevealSecrets
            }
            onMouseLeave={
              !isWebhookSecretMasked ? copyButtonProps.onMouseLeave : undefined
            }
            text={
              !isWebhookSecretMasked ? (
                <FormattedMessage id="copy" />
              ) : (
                <FormattedMessage id="reveal" />
              )
            }
          />
          <CopyFeedbackComponent />
        </div>
        <Text block={true} variant="medium" className="mt-2">
          <FormattedMessage
            id="SMSProviderConfigurationScreen.form.webhook.signatureKey.description"
            values={{
              href: `https://docs.authgear.com/customization/events-hooks/webhooks#verifying-signature`,
            }}
          />
        </Text>
      </div>
      <FormTextField
        type="number"
        label={renderToString(
          "SMSProviderConfigurationScreen.form.webhook.timeout"
        )}
        value={String(form.state.webhookTimeout)}
        onChange={onTimeoutChange}
        disabled={isWebhookSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="timeout"
        description={renderToString(
          "SMSProviderConfigurationScreen.form.webhook.timeout.description"
        )}
      />
    </div>
  );
}

function DenoHookForm({ form }: { form: FormModel }) {
  const { renderToString } = useContext(MessageContext);
  const { state, setState } = form;

  const onTimeoutChange = useCallback(
    (event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const value = parseInt(event.currentTarget.value, 10);
      if (isNaN(value)) {
        return;
      }
      setState((prevState) => {
        return {
          ...prevState,
          denoHookTimeout: value,
        } satisfies FormState;
      });
    },
    [setState]
  );

  const resourceIdx = useDenoScriptResourceIndex(form.state);

  // Generate a new script resource if one does not exist
  useEffect(() => {
    if (state.providerType !== SMSProviderType.Deno || resourceIdx !== -1) {
      return;
    }
    setState((prevState) => {
      return produce(prevState, (prevState) => {
        prevState.denoHookURL = makeNewDenoScriptURL();
        const path = getDenoScriptPathFromURL(prevState.denoHookURL);
        const specifier = makeDenoScriptSpecifier(prevState.denoHookURL);
        const r = prevState.resources.find((r) => r.path === path);
        if (r == null) {
          prevState.resources.push({
            path,
            specifier,
            nullableValue: DEFAULT_SMS_SCRIPT_TEMPLATE,
          });
        }
      });
    });
  }, [
    resourceIdx,
    setState,
    state.denoHookURL,
    state.providerType,
    state.resources,
  ]);

  const onChangeCode = useCallback(
    (newValue?: string) => {
      if (newValue == null) {
        return;
      }
      if (resourceIdx === -1) {
        return;
      }
      setState((prevState) =>
        produce(prevState, (prevState) => {
          prevState.resources[resourceIdx].nullableValue = newValue;
        })
      );
    },
    [resourceIdx, setState]
  );

  return (
    <div className="flex flex-col gap-y-4">
      <div>
        <Text block={true} variant="medium" className="font-semibold leading-5">
          <FormattedMessage id="SMSProviderConfigurationScreen.form.deno.script" />
        </Text>
        <CodeEditor
          className="block h-120"
          language="typescript"
          value={
            resourceIdx !== -1
              ? state.resources[resourceIdx].nullableValue ?? ""
              : ""
          }
          onChange={onChangeCode}
          options={CODE_EDITOR_OPTIONS}
        />
      </div>
      <FormTextField
        type="number"
        label={renderToString(
          "SMSProviderConfigurationScreen.form.deno.timeout"
        )}
        value={String(form.state.denoHookTimeout)}
        onChange={onTimeoutChange}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="timeout"
        description={renderToString(
          "SMSProviderConfigurationScreen.form.deno.timeout.description"
        )}
      />
    </div>
  );
}
