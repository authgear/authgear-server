import cn from "classnames";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import authgear from "@authgear/web";
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
  useRef,
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
import FormContainer from "../../FormContainer";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
  SMSProvider,
  SMSProviderTwilioCredentials,
  getHookKind,
} from "../../types";
import { produce } from "immer";
import { FormattedMessage, Context as MessageContext } from "../../intl";
import ScreenContent from "../../ScreenContent";
import styles from "./SMSProviderConfigurationScreen.module.css";
import logoTwilio from "../../images/twilio_logo.svg";
import logoWebhook from "../../images/webhook_logo.svg";
import logoAuthgear from "../../images/authgear_logo.svg";
import { startReauthentication } from "./Authenticated";
import { CodeField } from "../../components/common/CodeField";
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
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { TestSMSDialog } from "../../components/sms-provider/TestSMSDialog";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { RedMessageBar_RemindConfigureSMSProviderInSMSProviderScreen } from "../../RedMessageBar";
import ExternalLink from "../../ExternalLink";
import {
  IconRadioCards,
  IconRadioCardOption,
} from "../../components/v2/IconRadioCards/IconRadioCards";
import { TextField } from "../../components/v2/TextField/TextField";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import {
  Flex,
  IconButton as RadixIconButton,
  RadioGroup,
  Text,
  Tooltip as RadixTooltip,
} from "@radix-ui/themes";
import { CodeIcon, EyeOpenIcon, InfoCircledIcon } from "@radix-ui/react-icons";
import { FormField } from "../../components/v2/FormField/FormField";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { useFormContainerBaseContext } from "../../FormContainerBase";

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
  Authgear = "authgear",
  Twilio = "twilio",
  Webhook = "webhook",
  Deno = "deno",
}

enum TwilioSenderType {
  MessagingServiceSID = "MessagingServiceSID",
  From = "From",
}

const MASK = "********";

// Matches v2 IconRadioCards storybook inner icon size (SquareIcon iconSize).
const PROVIDER_RADIO_ICON_SIZE = "1.375rem";

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
    config.messaging?.sms_gateway?.provider === "twilio";
  const hasCustomTwilioCredentials =
    secrets.smsProviderSecrets?.twilioCredentials != null;

  const isSMSGatewayIsCustom =
    config.messaging?.sms_gateway?.provider === "custom";
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
    providerType = SMSProviderType.Authgear;
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
        case SMSProviderType.Authgear:
          config.messaging.sms_gateway = undefined;
          config.messaging.sms_provider = undefined;
          return;
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
        case SMSProviderType.Authgear:
          secrets.smsProviderSecrets = null;
          break;
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
    case SMSProviderType.Authgear:
      return {
        smsProviderSecrets: {
          action: "set",
          setData: {},
        },
      };
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
  const specifiers: ResourceSpecifier[] = [];
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
      case SMSProviderType.Authgear:
        return null;
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
    case SMSProviderType.Authgear:
      return false;
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
      if (!authgear.canReauthenticate()) {
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

  // eslint-disable-next-line react-hooks/preserve-manual-memoization
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
    getIsDirty: () => configForm.getIsDirty() || resources.getIsDirty(),
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
  const { getIsDirty } = useFormContainerBaseContext();
  const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
  const navigate = useNavigate();
  const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

  const [isReauthDialogOpen, setIsReauthDialogOpen] = useState(false);
  const [isTestSMSDialogHidden, setIsTestSMSDialogHidden] = useState(true);

  const { checkDenoHook, loading: checkDenoHookLoading } = checkDenoHookHandle;

  const isSecretMasked = useMemo(
    () => computeIsSecretMasked(form.state),
    [form.state]
  );

  const onChangeProviderType = useCallback(
    (value: SMSProviderType) => {
      if (isSecretMasked) {
        setIsReauthDialogOpen(true);
        return;
      }
      setState((s) => ({
        ...s,
        enabled: value !== SMSProviderType.Authgear,
        providerType: value,
      }));
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
    setIsReauthDialogOpen(true);
  }, []);

  const testConfig = useTestSMSConfig(form.state);

  const onTestSMS = useCallback(async () => {
    if (isSecretMasked) {
      setIsReauthDialogOpen(true);
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

  const providerOptions = useMemo(
    (): IconRadioCardOption<SMSProviderType>[] => [
      {
        value: SMSProviderType.Authgear,
        icon: (
          <img
            src={logoAuthgear}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.authgear" />
        ),
      },
      {
        value: SMSProviderType.Twilio,
        icon: (
          <img
            src={logoTwilio}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.twilio" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
      {
        value: SMSProviderType.Webhook,
        icon: (
          <img
            src={logoWebhook}
            alt=""
            className="object-contain"
            style={{
              width: PROVIDER_RADIO_ICON_SIZE,
              height: PROVIDER_RADIO_ICON_SIZE,
            }}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.webhook" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
      {
        value: SMSProviderType.Deno,
        icon: (
          <CodeIcon
            width={PROVIDER_RADIO_ICON_SIZE}
            height={PROVIDER_RADIO_ICON_SIZE}
          />
        ),
        title: (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.deno" />
        ),
        disabled: isCustomSMSProviderDisabled,
      },
    ],
    [isCustomSMSProviderDisabled]
  );

  const providerDescription = useMemo(() => {
    switch (state.providerType) {
      case SMSProviderType.Authgear:
        return (
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.authgear.description" />
        );
      case SMSProviderType.Twilio:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.twilio.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://docs.authgear.com/customization/custom-providers/twilio">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      case SMSProviderType.Webhook:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.webhook.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://docs.authgear.com/customization/custom-providers/webhook-custom-script">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      case SMSProviderType.Deno:
        return (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.deno.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://docs.authgear.com/customization/custom-providers/webhook-custom-script">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      default:
        return null;
    }
  }, [state.providerType]);

  const showSettings = state.providerType !== SMSProviderType.Authgear;

  return (
    <>
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            {isAuthgearOnce ? (
              <FormattedMessage id="SMSProviderConfigurationScreen.title--authgearonce" />
            ) : (
              <FormattedMessage id="SMSProviderConfigurationScreen.title" />
            )}
          </Text>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="SMSProviderConfigurationScreen.description" />
          </Text>
        </div>
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

        <div className={cn(styles.widget, styles.providerSelector)}>
          <IconRadioCards
            size="3"
            value={state.providerType}
            onValueChange={onChangeProviderType}
            options={providerOptions}
            itemFillSpaces={true}
          />
          {providerDescription != null ? (
            <Text
              as="p"
              size="1"
              color="gray"
              className={styles.providerDescription}
            >
              {providerDescription}
            </Text>
          ) : null}
        </div>

        {showSettings ? (
          <SettingsSectionCard
            className={styles.widget}
            contentClassName="gap-4"
            title={
              <FormattedMessage id="SMSProviderConfigurationScreen.settings.label" />
            }
          >
            <FormSection form={form} onRevealSecrets={onRevealSecrets} />
            {isSecretMasked ? (
              <div>
                <PrimaryButton
                  size="3"
                  disabled={isCustomSMSProviderDisabled}
                  onClick={onRevealSecrets}
                  text={<FormattedMessage id="edit" />}
                />
              </div>
            ) : (
              <div>
                <SecondaryButton
                  size="2"
                  // eslint-disable-next-line @typescript-eslint/strict-void-return
                  onClick={onTestSMS}
                  disabled={testConfig == null || checkDenoHookLoading}
                  text={
                    <FormattedMessage id="SMSProviderConfigurationScreen.testSMS" />
                  }
                />
              </div>
            )}
          </SettingsSectionCard>
        ) : null}

        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
      <ConfirmationDialog
        open={isReauthDialogOpen}
        onOpenChange={(open) => {
          if (!open) {
            setIsReauthDialogOpen(false);
          }
        }}
        title={<FormattedMessage id="ReauthDialog.title" />}
        description={<FormattedMessage id="ReauthDialog.description" />}
        confirmText={<FormattedMessage id="confirm" />}
        cancelText={<FormattedMessage id="cancel" />}
        confirmColor="indigo"
        onConfirm={triggerReauth}
        onCancel={useCallback(() => {
          setIsReauthDialogOpen(false);
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

function FormSection({
  form,
  onRevealSecrets,
}: {
  form: FormModel;
  onRevealSecrets: () => void;
}) {
  switch (form.state.providerType) {
    case SMSProviderType.Authgear:
      return null;
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
      return (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        form.setState((prevState) => {
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

  const onCredentialTypeChange = useCallback(
    (value: string) => {
      form.setState((prev) => ({
        ...prev,
        twilioCredentialType: value as TwilioCredentialType,
      }));
    },
    [form]
  );

  const onSenderTypeChange = useCallback(
    (value: string) => {
      form.setState((prev) => ({
        ...prev,
        twilioSenderType: value as TwilioSenderType,
      }));
    },
    [form]
  );

  return (
    <div className="flex flex-col gap-y-4">
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.twilioSID" />
        }
        value={form.state.twilioSID}
        required={true}
        onChange={onStringChangeCallbacks.twilioSID}
        disabled={isTwilioSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="account_sid"
      />
      <div className="flex flex-col gap-3">
        <FormField
          size="2"
          labelSize="2"
          label={
            <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.sender" />
          }
          required={true}
          labelSpace="1"
        >
          <RadioGroup.Root
            value={form.state.twilioSenderType}
            onValueChange={onSenderTypeChange}
            disabled={isTwilioSecretMasked}
          >
            <Flex direction="column" gap="2">
              <Text as="label" size="2">
                <Flex gap="2" align="center">
                  <RadioGroup.Item
                    value={TwilioSenderType.MessagingServiceSID}
                  />
                  <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID" />
                  <Tooltip
                    content={
                      <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID.tooltip" />
                    }
                  >
                    <InfoCircledIcon className={styles.senderInfoIcon} />
                  </Tooltip>
                </Flex>
              </Text>
              <Text as="label" size="2">
                <Flex gap="2" align="center">
                  <RadioGroup.Item value={TwilioSenderType.From} />
                  <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.twilioFrom" />
                  <Tooltip
                    content={
                      <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.twilioFrom.tooltip" />
                    }
                  >
                    <InfoCircledIcon className={styles.senderInfoIcon} />
                  </Tooltip>
                </Flex>
              </Text>
            </Flex>
          </RadioGroup.Root>
        </FormField>
        {form.state.twilioSenderType ===
        TwilioSenderType.MessagingServiceSID ? (
          <div className="flex flex-col">
            <TextField
              size="2"
              labelSize="2"
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
            <Text as="p" size="1" color="gray">
              <FormattedMessage
                id="SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID.hint"
                values={{
                  // eslint-disable-next-line react/no-unstable-nested-components
                  ExternalLink: (chunks: React.ReactNode) => (
                    <ExternalLink href="https://www.twilio.com/docs/messaging/services">
                      {chunks}
                    </ExternalLink>
                  ),
                }}
              />
            </Text>
          </div>
        ) : (
          <TextField
            size="2"
            labelSize="2"
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
      <div className="flex flex-col gap-4">
        <FormField
          size="2"
          labelSize="2"
          label={
            <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.credentialType" />
          }
          required={true}
          labelSpace="1"
        >
          <RadioGroup.Root
            value={form.state.twilioCredentialType}
            onValueChange={onCredentialTypeChange}
            disabled={isTwilioSecretMasked}
          >
            <Flex direction="column" gap="2">
              <Text as="label" size="2">
                <Flex gap="2" align="center">
                  <RadioGroup.Item value={TwilioCredentialType.AuthToken} />
                  <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.credentialType.options.authToken" />
                </Flex>
              </Text>
              <Text as="label" size="2">
                <Flex gap="2" align="center">
                  <RadioGroup.Item value={TwilioCredentialType.ApiKey} />
                  <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.credentialType.options.apiKey" />
                </Flex>
              </Text>
            </Flex>
          </RadioGroup.Root>
        </FormField>
        {form.state.twilioCredentialType === TwilioCredentialType.AuthToken ? (
          <TextField
            size="2"
            labelSize="2"
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
            <TextField
              size="2"
              labelSize="2"
              type="text"
              label={
                <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.apiKeySID" />
              }
              value={form.state.twilioAPIKeySID}
              onChange={onStringChangeCallbacks.twilioAPIKeySID}
              disabled={isTwilioSecretMasked}
              parentJSONPointer={/\/secrets\/\d+\/data/}
              fieldName="api_key_sid"
            />
            <TextField
              size="2"
              labelSize="2"
              type="text"
              label={
                <FormattedMessage id="SMSProviderConfigurationScreen.form.twilio.apiKeySecret" />
              }
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

function RevealIconButton({
  onClick,
}: {
  onClick: () => void;
}): React.ReactElement {
  const { renderToString } = useContext(MessageContext);

  return (
    <RadixTooltip content={renderToString("reveal")}>
      <RadixIconButton
        type="button"
        variant="ghost"
        color="gray"
        size="1"
        aria-label={renderToString("reveal")}
        onClick={onClick}
        className={styles.copyIconButton}
      >
        <EyeOpenIcon width="1rem" height="1rem" />
      </RadixIconButton>
    </RadixTooltip>
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
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      form.setState((prevState) => {
        return {
          ...prevState,
          webhookURL: value,
        } satisfies FormState;
      });
    },
    [form]
  );

  const onTimeoutChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = parseInt(e.target.value, 10);
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

  const isWebhookSecretMasked = form.state.webhookSecretKey == null;
  const webhookSecretKey = form.state.webhookSecretKey ?? "";

  return (
    <div className="flex flex-col gap-y-4">
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.webhook.url" />
        }
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
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.webhook.signatureKey" />
        }
        value={isWebhookSecretMasked ? MASK : webhookSecretKey}
        readOnly={true}
        suffixPlain={true}
        suffix={
          isWebhookSecretMasked ? (
            <RevealIconButton onClick={onRevealSecrets} />
          ) : webhookSecretKey.length > 0 ? (
            <CopyIconButton textToCopy={webhookSecretKey} />
          ) : undefined
        }
        hint={
          <FormattedMessage
            id="SMSProviderConfigurationScreen.form.webhook.signatureKey.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://docs.authgear.com/customization/events-hooks/webhooks#verifying-signature">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        }
      />
      <TextField
        size="2"
        labelSize="2"
        type="number"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.webhook.timeout" />
        }
        hint={renderToString(
          "SMSProviderConfigurationScreen.form.webhook.timeout.description"
        )}
        value={String(form.state.webhookTimeout)}
        onChange={onTimeoutChange}
        disabled={isWebhookSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="timeout"
      />
    </div>
  );
}

function DenoHookForm({ form }: { form: FormModel }) {
  const { renderToString } = useContext(MessageContext);
  const { state, setState } = form;

  const onTimeoutChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = parseInt(e.target.value, 10);
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
      <div className="border border-[var(--gray-5)] rounded-xl overflow-hidden">
        <div className="px-6 py-4 border-b border-[var(--gray-5)]">
          <Text as="p" size="3" weight="medium">
            <FormattedMessage id="SMSProviderConfigurationScreen.form.deno.script" />
          </Text>
        </div>
        <div className="px-0 py-0">
          <CodeEditor
            className="block h-[412px]"
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
      </div>
      <TextField
        size="2"
        labelSize="2"
        type="number"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.deno.timeout" />
        }
        hint={renderToString(
          "SMSProviderConfigurationScreen.form.deno.timeout.description"
        )}
        value={String(form.state.denoHookTimeout)}
        onChange={onTimeoutChange}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="timeout"
      />
    </div>
  );
}
