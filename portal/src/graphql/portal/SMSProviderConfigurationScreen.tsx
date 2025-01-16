import cn from "classnames";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import { AppSecretKey } from "./globalTypes.generated";
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
import FormContainer from "../../FormContainer";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
  SMSProvider,
  getHookKind,
} from "../../types";
import { produce } from "immer";
import {
  FormattedMessage,
  Context as MessageContext,
} from "@oursky/react-messageformat";
import { Text } from "@fluentui/react";
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

enum SMSProviderType {
  Twilio = "twilio",
  Webhook = "webhook",
  Deno = "deno",
}

interface ConfigFormState {
  enabled: boolean;
  providerType: SMSProviderType;
  webhookSecretKey: string | null;

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

interface FormState extends ConfigFormState {
  resources: Resource[];
  diff: ResourcesDiffResult | null;
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

  let twilioSID = "";
  let twilioAuthToken: string | null = "";
  let twilioMessagingServiceSID = "";

  if (enabled && providerType === SMSProviderType.Twilio) {
    twilioSID = secrets.smsProviderSecrets?.twilioCredentials?.accountSID ?? "";
    twilioAuthToken =
      secrets.smsProviderSecrets?.twilioCredentials != null
        ? secrets.smsProviderSecrets.twilioCredentials.authToken ?? null
        : "";
    twilioMessagingServiceSID =
      secrets.smsProviderSecrets?.twilioCredentials?.messagingServiceSID ?? "";
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
        case SMSProviderType.Twilio:
          secrets.smsProviderSecrets = {
            twilioCredentials: {
              accountSID: currentState.twilioSID,
              authToken: currentState.twilioAuthToken,
              messagingServiceSID: currentState.twilioMessagingServiceSID,
            },
          };
          break;
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
              accountSID:
                secrets.smsProviderSecrets.twilioCredentials.accountSID,
              authToken: secrets.smsProviderSecrets.twilioCredentials.authToken,
              messagingServiceSID:
                secrets.smsProviderSecrets.twilioCredentials
                  .messagingServiceSID,
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
import { CustomSMSGatewayPayload } from "${DENO_TYPES_URL}";

export default async function (e: CustomSMSGatewayPayload): Promise<void> {
     const response = await fetch("https://some.sms.gateway");
     if (!response.ok) {
          throw new Error("Failed to send sms");
     }
}
`;

const CODE_EDITOR_OPTIONS = {
  minimap: {
    enabled: false,
  },
};

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

  const state = useMemo<FormState>(() => {
    return {
      ...configForm.state,
      resources: resources.state,
      diff: resources.diff,
    };
  }, [configForm.state, resources.state, resources.diff]);

  const form: AppSecretConfigFormModel<FormState> = {
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
      <SMSProviderConfigurationContent form={form} />
    </FormContainer>
  );
}

function SMSProviderConfigurationContent(props: {
  form: AppSecretConfigFormModel<FormState>;
}) {
  const { form } = props;
  const { state, setState } = form;
  const { renderToString } = useContext(MessageContext);
  const navigate = useNavigate();

  const onChangeEnabled = useCallback(
    (_event, checked?: boolean) => {
      if (checked != null) {
        setState((state) => {
          return {
            ...state,
            enabled: checked,
          };
        });
      }
    },
    [setState]
  );

  const onRevealSecrets = useCallback(() => {
    const state: LocationState = {
      isRevealSecrets: true,
    };

    startReauthentication(navigate, state).catch((e) => {
      // Normally there should not be any error.
      console.error(e);
    });
  }, [navigate]);

  return (
    <ScreenContent>
      <ScreenTitle className={styles.widget}>
        <FormattedMessage id="SMSProviderConfigurationScreen.title" />
      </ScreenTitle>
      <ScreenDescription className={styles.widget}>
        <FormattedMessage id="SMSProviderConfigurationScreen.description" />
      </ScreenDescription>

      <Widget className={styles.widget} contentLayout="grid">
        <Toggle
          className={styles.columnFull}
          checked={state.enabled}
          onChange={onChangeEnabled}
          label={renderToString("SMSProviderConfigurationScreen.enable.label")}
          inlineLabel={true}
        />
      </Widget>

      {state.enabled ? (
        <Widget className={cn(styles.widget, "flex flex-col gap-y-4")}>
          <ProviderSection form={form} />
          <FormSection form={form} onRevealSecrets={onRevealSecrets} />
        </Widget>
      ) : null}
    </ScreenContent>
  );
}

function ProviderSection({
  form,
}: {
  form: AppSecretConfigFormModel<FormState>;
}) {
  const onSelectTwilio = useCallback(() => {
    form.setState((state) => {
      return { ...state, providerType: SMSProviderType.Twilio };
    });
  }, [form]);

  const onSelectWebhook = useCallback(() => {
    form.setState((state) => {
      return { ...state, providerType: SMSProviderType.Webhook };
    });
  }, [form]);
  const onSelectDeno = useCallback(() => {
    form.setState((state) => {
      return { ...state, providerType: SMSProviderType.Deno };
    });
  }, [form]);

  return (
    <div className="flex flex-col gap-y-3">
      <Text variant="xLarge">
        <FormattedMessage id="SMSProviderConfigurationScreen.provider.title" />
      </Text>
      <div className={styles.providerGrid}>
        <ProviderCard
          onClick={onSelectTwilio}
          isSelected={form.state.providerType === SMSProviderType.Twilio}
          logoSrc={logoTwilio}
        >
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.twilio" />
        </ProviderCard>
        <ProviderCard
          onClick={onSelectWebhook}
          isSelected={form.state.providerType === SMSProviderType.Webhook}
          logoSrc={logoWebhook}
        >
          <FormattedMessage id="SMSProviderConfigurationScreen.provider.webhook" />
        </ProviderCard>
        <ProviderCard
          onClick={onSelectDeno}
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
              href: "https://docs.authgear.com/how-to-guide/integration/custom-sms-provider/twilio",
            }}
          />
        ) : form.state.providerType === SMSProviderType.Webhook ? (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.webhook.description"
            values={{
              href: "https://docs.authgear.com/how-to-guide/integration/custom-sms-provider/webhook-custom-script",
            }}
          />
        ) : (
          <FormattedMessage
            id="SMSProviderConfigurationScreen.provider.deno.description"
            values={{
              href: "https://docs.authgear.com/how-to-guide/integration/custom-sms-provider/webhook-custom-script",
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
  form: AppSecretConfigFormModel<FormState>;
  onRevealSecrets: () => void;
}) {
  switch (form.state.providerType) {
    case SMSProviderType.Twilio:
      return <TwilioForm form={form} onRevealSecrets={onRevealSecrets} />;
    case SMSProviderType.Webhook:
      return <WebhookForm form={form} onRevealSecrets={onRevealSecrets} />;
    case SMSProviderType.Deno:
      return <DenoHookForm form={form} />;
  }
}

function TwilioForm({
  form,
  onRevealSecrets,
}: {
  form: AppSecretConfigFormModel<FormState>;
  onRevealSecrets: () => void;
}) {
  const { renderToString } = useContext(MessageContext);

  const onChangeCallbacks = useMemo(() => {
    const callbackFactory = (
      key: "twilioSID" | "twilioAuthToken" | "twilioMessagingServiceSID"
    ) => {
      return (
        event: React.FormEvent<HTMLInputElement | HTMLTextAreaElement>
      ) => {
        form.setState((prevState) => {
          const value = event.currentTarget.value;
          return {
            ...prevState,
            [key]: value,
          } satisfies FormState;
        });
      };
    };
    return {
      twilioSID: callbackFactory("twilioSID"),
      twilioAuthToken: callbackFactory("twilioAuthToken"),
      twilioMessagingServiceSID: callbackFactory("twilioMessagingServiceSID"),
    };
  }, [form]);

  const isTwilioSecretMasked = form.state.twilioAuthToken == null;

  return (
    <div className="flex flex-col gap-y-4">
      <FormTextField
        type="text"
        label={renderToString(
          "SMSProviderConfigurationScreen.form.twilio.twilioSID"
        )}
        value={form.state.twilioSID}
        required={true}
        onChange={onChangeCallbacks.twilioSID}
        disabled={isTwilioSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="account_sid"
      />
      <FormTextField
        type="text"
        label={renderToString(
          "SMSProviderConfigurationScreen.form.twilio.twilioAuthToken"
        )}
        value={
          isTwilioSecretMasked ? "********" : form.state.twilioAuthToken ?? ""
        }
        disabled={isTwilioSecretMasked}
        required={true}
        onChange={onChangeCallbacks.twilioAuthToken}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="auth_token"
      />
      <FormTextField
        type="text"
        label={renderToString(
          "SMSProviderConfigurationScreen.form.twilio.twilioMessagingServiceSID"
        )}
        value={form.state.twilioMessagingServiceSID}
        onChange={onChangeCallbacks.twilioMessagingServiceSID}
        disabled={isTwilioSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="message_service_sid"
      />
      {isTwilioSecretMasked ? (
        <PrimaryButton
          className="w-min"
          onClick={onRevealSecrets}
          text={<FormattedMessage id="edit" />}
        />
      ) : null}
    </div>
  );
}

function WebhookForm({
  form,
  onRevealSecrets,
}: {
  form: AppSecretConfigFormModel<FormState>;
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

  // eslint-disable-next-line no-useless-assignment
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
              isWebhookSecretMasked
                ? "********"
                : form.state.webhookSecretKey ?? ""
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
              href: `https://docs.authgear.com/integrate/events-hooks/webhooks#verifying-signature`,
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
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="timeout"
        description={renderToString(
          "SMSProviderConfigurationScreen.form.webhook.timeout.description"
        )}
      />
    </div>
  );
}

function DenoHookForm({ form }: { form: AppSecretConfigFormModel<FormState> }) {
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
