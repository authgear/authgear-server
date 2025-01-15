import cn from "classnames";
import { useLocation, useParams } from "react-router-dom";
import { AppSecretKey } from "./globalTypes.generated";
import React, { useCallback, useContext, useMemo, useState } from "react";
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
  isSecretMasked: boolean;

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
    secrets.smsProviderSecrets?.twilioCredentials != null
      ? secrets.smsProviderSecrets.twilioCredentials.authToken ?? null
      : "";
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
    isSecretMasked:
      secrets.smsProviderSecrets?.twilioCredentials != null &&
      secrets.smsProviderSecrets.twilioCredentials.authToken == null,

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
          disabled={form.state.isSecretMasked}
        />
      </Widget>

      {state.enabled ? (
        <Widget className={cn(styles.widget, "flex flex-col gap-y-4")}>
          <ProviderSection form={form} />
          <FormSection form={form} />
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
    </div>
  );
}

function FormSection({ form }: { form: AppSecretConfigFormModel<FormState> }) {
  switch (form.state.providerType) {
    case SMSProviderType.Twilio:
      return <TwilioForm form={form} />;
    case SMSProviderType.Webhook:
      return <></>;
    case SMSProviderType.Deno:
      return <></>;
  }
}

function TwilioForm({ form }: { form: AppSecretConfigFormModel<FormState> }) {
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
          };
        });
      };
    };
    return {
      twilioSID: callbackFactory("twilioSID"),
      twilioAuthToken: callbackFactory("twilioAuthToken"),
      twilioMessagingServiceSID: callbackFactory("twilioMessagingServiceSID"),
    };
  }, [form]);

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
        disabled={form.state.isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="account_sid"
      />
      <FormTextField
        type="text"
        label={renderToString(
          "SMSProviderConfigurationScreen.form.twilio.twilioAuthToken"
        )}
        value={
          form.state.isSecretMasked
            ? "********"
            : form.state.twilioAuthToken ?? ""
        }
        disabled={form.state.isSecretMasked}
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
        disabled={form.state.isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="message_service_sid"
      />
    </div>
  );
}
