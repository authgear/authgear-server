import cn from "classnames";
import React, { useCallback, useContext, useState, useMemo } from "react";
import { useLocation, useParams, useNavigate } from "react-router-dom";
import { produce } from "immer";
import { Dialog, DialogFooter } from "@fluentui/react";
import { FormattedMessage, Context } from "../../intl";
import { parseSender } from "email-addresses";
import { useTextFieldTooltip } from "../../useTextFieldTooltip";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import FormContainer from "../../FormContainer";
import FormTextField from "../../FormTextField";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import Widget from "../../Widget";
import TextField from "../../TextField";
import Toggle from "../../Toggle";
import { startReauthentication } from "./Authenticated";
import {
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
} from "../../types";
import { useViewerQuery } from "./query/viewerQuery";
import {
  SendTestEmailOptions,
  useSendTestEmailMutation,
  UseSendTestEmailMutationReturnType,
} from "./mutations/sendTestEmail";
import logoSendgrid from "../../images/sendgrid_logo.png";
import styles from "./SMTPConfigurationScreen.module.css";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";
import ExternalLink from "../../ExternalLink";


import { AppSecretKey } from "./globalTypes.generated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import {
  ProviderCard,
  ProviderCardDescription,
} from "../../components/common/ProviderCard";
import { ErrorParseRule, ErrorParseRuleResult } from "../../error/parse";
import { APIError, APISMTPTestFailedError } from "../../error/error";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { RedMessageBar_RemindConfigureSMTPInSMTPConfigurationScreen } from "../../RedMessageBar";

interface LocationState {
  isEdit: boolean;
}
function isLocationState(raw: unknown): raw is LocationState {
  return (
    raw != null &&
    typeof raw === "object" &&
    (raw as Partial<LocationState>).isEdit != null
  );
}

enum ProviderType {
  Sendgrid = "sendgrid",
  Custom = "custom",
}

const MASKED_PASSWORD_VALUE = "****************";

const SENDGRID_HOST = "smtp.sendgrid.net";
const SENDGRID_PORT_STRING = "587";
const SENDGRID_USERNAME = "apikey";

interface ConfigFormState {
  enabled: boolean;
  providerType: ProviderType;
  sendgridAPIKey: string;
  sendgridSenderName: string;
  sendgridSenderAddress: string;

  customHost: string;
  customPortString: string;
  customUsername: string;
  customPassword: string;
  customSenderName: string;
  customSenderAddress: string;

  isPasswordMasked: boolean;
}

interface FormState extends ConfigFormState {
  isSMTPRequiredForSomeEnabledFeatures: boolean;
  smtpConfigured: boolean;
}

function constructFormState(
  _config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): ConfigFormState {
  const enabled = secrets.smtpSecret != null;

  const isSendgrid =
    secrets.smtpSecret?.host === SENDGRID_HOST &&
    String(secrets.smtpSecret.port) === SENDGRID_PORT_STRING &&
    secrets.smtpSecret.username === SENDGRID_USERNAME;
  const providerType = isSendgrid ? ProviderType.Sendgrid : ProviderType.Custom;
  const isPasswordMasked = enabled && secrets.smtpSecret?.password == null;

  let sendgridAPIKey = "";
  let sendgridSenderName = "";
  let sendgridSenderAddress = "";
  let customHost = "";
  let customPortString = "";
  let customUsername = "";
  let customPassword = "";
  let customSenderName = "";
  let customSenderAddress = "";

  const secretSender = secrets.smtpSecret?.sender
    ? parseSender(secrets.smtpSecret.sender)
    : null;

  switch (providerType) {
    case ProviderType.Sendgrid:
      sendgridAPIKey = secrets.smtpSecret?.password ?? "";
      if (secretSender?.type === "mailbox") {
        sendgridSenderName = secretSender.name ?? "";
        sendgridSenderAddress = secretSender.address;
      }
      break;
    case ProviderType.Custom:
      customHost = secrets.smtpSecret?.host ?? "";
      customPortString =
        secrets.smtpSecret?.port != null ? String(secrets.smtpSecret.port) : "";
      customUsername = secrets.smtpSecret?.username ?? "";
      customPassword = secrets.smtpSecret?.password ?? "";
      if (secretSender?.type === "mailbox") {
        customSenderName = secretSender.name ?? "";
        customSenderAddress = secretSender.address;
      }
      break;

    default:
      break;
  }
  return {
    enabled,
    providerType,
    sendgridAPIKey,
    sendgridSenderName,
    sendgridSenderAddress,
    customHost,
    customPortString,
    customUsername,
    customPassword,
    customSenderName,
    customSenderAddress,
    isPasswordMasked,
  };
}

function composeSender(name: string, address: string) {
  if (!name) {
    return address;
  }
  return `${name} <${address}>`;
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _initialState: ConfigFormState,
  currentState: ConfigFormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  const newSecrets = produce(secrets, (secrets) => {
    if (!currentState.enabled) {
      secrets.smtpSecret = null;
    } else {
      switch (currentState.providerType) {
        case ProviderType.Sendgrid:
          secrets.smtpSecret = {
            host: SENDGRID_HOST,
            port: Number(SENDGRID_PORT_STRING),
            username: SENDGRID_USERNAME,
            password: currentState.isPasswordMasked
              ? null
              : currentState.sendgridAPIKey,
            sender: composeSender(
              currentState.sendgridSenderName,
              currentState.sendgridSenderAddress
            ),
          };
          break;
        case ProviderType.Custom:
          secrets.smtpSecret = {
            host: currentState.customHost,
            port: Number(currentState.customPortString),
            username: currentState.customUsername,
            password: currentState.isPasswordMasked
              ? null
              : currentState.customPassword,
            sender: composeSender(
              currentState.customSenderName,
              currentState.customSenderAddress
            ),
          };
          break;
        default:
          console.error("unexpected provider type", currentState.providerType);
      }
    }
  });
  return [config, newSecrets];
}

function constructSecretUpdateInstruction(
  _config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _currentState: ConfigFormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  if (!secrets.smtpSecret) {
    return {
      smtpSecret: {
        action: "unset",
      },
    };
  }

  // the password is masked, no change
  if (!secrets.smtpSecret.password) {
    return undefined;
  }

  return {
    smtpSecret: {
      action: "set",
      data: {
        host: secrets.smtpSecret.host,
        port: secrets.smtpSecret.port,
        username: secrets.smtpSecret.username,
        password: secrets.smtpSecret.password,
        sender: secrets.smtpSecret.sender,
      },
    },
  };
}

const CUSTOM_PROVIDER_ICON_PROPS = {
  iconName: "Mail",
};

const ERROR_RULES: ErrorParseRule[] = [
  (apiError: APIError): ErrorParseRuleResult => {
    if (apiError.reason === "SMTPTestFailed") {
      return {
        parsedAPIErrors: [
          { message: (apiError as APISMTPTestFailedError).message },
        ],
        fullyHandled: true,
      };
    }
    return {
      parsedAPIErrors: [],
      fullyHandled: false,
    };
  },
];

interface SMTPConfigurationScreenContentProps {
  isCustomSMTPDisabled: boolean;
  sendTestEmailHandle: UseSendTestEmailMutationReturnType;
  form: AppSecretConfigFormModel<FormState>;
}

const SMTPConfigurationScreenContent: React.VFC<SMTPConfigurationScreenContentProps> =
  function SMTPConfigurationScreenContent(props) {
    const { form, sendTestEmailHandle, isCustomSMTPDisabled } = props;
    const { state, setState } = form;
    const { isSMTPRequiredForSomeEnabledFeatures, smtpConfigured } = state;
    const { sendTestEmail, loading } = sendTestEmailHandle;

    const { isAuthgearOnce } = useSystemConfig();

    const [isDialogHidden, setIsDialogHidden] = useState(true);
    const [toAddress, setToAddress] = useState("");
    const { viewer } = useViewerQuery();
    const { renderToString } = useContext(Context);

    const openSendTestEmailDialogButtonEnabled = useMemo(() => {
      if (!state.enabled) {
        return false;
      }
      switch (state.providerType) {
        case ProviderType.Sendgrid:
          return (
            state.sendgridAPIKey !== "" && state.sendgridSenderAddress !== ""
          );
        case ProviderType.Custom:
          return (
            state.customHost !== "" &&
            state.customPortString !== "" &&
            state.customUsername !== "" &&
            state.customPassword !== "" &&
            state.customSenderAddress !== ""
          );
      }
    }, [state]);

    const sendTestEmailButtonEnabled = useMemo(() => {
      return toAddress !== "";
    }, [toAddress]);

    const onStringChangeCallbacks = useMemo(() => {
      const callbackFactory = (
        key:
          | "sendgridAPIKey"
          | "sendgridSenderName"
          | "sendgridSenderAddress"
          | "customHost"
          | "customUsername"
          | "customPassword"
          | "customSenderName"
          | "customSenderAddress"
      ) => {
        return (_: unknown, value?: string) => {
          if (value != null) {
            setState((state) => {
              const s: FormState = {
                ...state,
              };
              s[key] = value;
              return s;
            });
          }
        };
      };
      return {
        sendgridAPIKey: callbackFactory("sendgridAPIKey"),
        sendgridSenderName: callbackFactory("sendgridSenderName"),
        sendgridSenderAddress: callbackFactory("sendgridSenderAddress"),
        customHost: callbackFactory("customHost"),
        customUsername: callbackFactory("customUsername"),
        customPassword: callbackFactory("customPassword"),
        customSenderName: callbackFactory("customSenderName"),
        customSenderAddress: callbackFactory("customSenderAddress"),
      };
    }, [setState]);

    const onCustomPortChange = useCallback(
      (_: unknown, value?: string) => {
        if (value != null) {
          let newValue: string | undefined;
          if (value !== "") {
            const port = Number(value);
            if (!isNaN(port)) {
              newValue = value;
            }
          } else {
            newValue = "";
          }
          if (newValue !== undefined) {
            const v = newValue;
            setState((state) => {
              return {
                ...state,
                customPortString: v,
              };
            });
          }
        }
      },
      [setState]
    );

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

    const navigate = useNavigate();

    const onClickEdit = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();

        const state: LocationState = {
          isEdit: true,
        };

        startReauthentication(navigate, state).catch((e) => {
          // Normally there should not be any error.
          console.error(e);
        });
      },
      [navigate]
    );

    const onClickSendTestEmail = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();

        setToAddress(viewer?.email ?? "");
        setIsDialogHidden(false);
      },
      [viewer]
    );

    const onDismissDialog = useCallback(
      (e?: React.MouseEvent<unknown>) => {
        e?.preventDefault();
        e?.stopPropagation();

        if (!loading) {
          setIsDialogHidden(true);
        }
      },
      [loading]
    );

    const onClickSend = useCallback(
      (e?: React.MouseEvent<unknown>) => {
        e?.preventDefault();
        e?.stopPropagation();

        let input: SendTestEmailOptions;
        switch (state.providerType) {
          case ProviderType.Sendgrid:
            input = {
              smtpHost: SENDGRID_HOST,
              smtpPort: parseInt(SENDGRID_PORT_STRING, 10),
              smtpUsername: SENDGRID_USERNAME,
              smtpPassword: state.sendgridAPIKey,
              smtpSender: composeSender(
                state.sendgridSenderName,
                state.sendgridSenderAddress
              ),
              to: toAddress,
            };
            break;
          case ProviderType.Custom:
            input = {
              smtpHost: state.customHost,
              smtpPort: parseInt(state.customPortString, 10),
              smtpUsername: state.customUsername,
              smtpPassword: state.customPassword,
              smtpSender: composeSender(
                state.customSenderName,
                state.customSenderAddress
              ),
              to: toAddress,
            };
            break;
        }

        sendTestEmail(input).then(
          () => {
            setIsDialogHidden(true);
          },
          () => {
            setIsDialogHidden(true);
          }
        );
      },
      [sendTestEmail, state, toAddress]
    );

    const onChangeToAddress = useCallback((_, value?: string) => {
      if (value != null) {
        setToAddress(value);
      }
    }, []);

    const dialogContentProps = useMemo(() => {
      return {
        title: renderToString(
          "SMTPConfigurationScreen.send-test-email-dialog.title"
        ),
        subText: renderToString(
          "SMTPConfigurationScreen.send-test-email-dialog.description"
        ),
      };
    }, [renderToString]);

    const onClickProviderSendgrid = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();

        setState((state) => {
          return {
            ...state,
            providerType: ProviderType.Sendgrid,
          };
        });
      },
      [setState]
    );

    const onClickProviderCustom = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();

        setState((state) => {
          return {
            ...state,
            providerType: ProviderType.Custom,
          };
        });
      },
      [setState]
    );

    const hostProps = useTextFieldTooltip({
      tooltipLabel: renderToString("SMTPConfigurationScreen.host.tooltip"),
    });

    const portProps = useTextFieldTooltip({
      tooltipLabel: renderToString("SMTPConfigurationScreen.port.tooltip"),
    });

    const sendgridAPIKeyProps = useTextFieldTooltip({
      tooltipLabel: renderToString(
        "SMTPConfigurationScreen.sendgrid.api-key.tooltip"
      ),
    });

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          {isAuthgearOnce ? (
            <FormattedMessage id="SMTPConfigurationScreen.title--authgearonce" />
          ) : (
            <FormattedMessage id="SMTPConfigurationScreen.title" />
          )}
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="SMTPConfigurationScreen.description" />
        </ScreenDescription>
        {isCustomSMTPDisabled ? (
          <FeatureDisabledMessageBar
            className={styles.widget}
            messageID="FeatureConfig.custom-smtp.disabled"
          />
        ) : null}
        {isAuthgearOnce &&
          isSMTPRequiredForSomeEnabledFeatures &&
          !smtpConfigured ? (
          <div className={cn(styles.widget, "flex flex-col")}>
            <RedMessageBar_RemindConfigureSMTPInSMTPConfigurationScreen className="self-start w-fit" />
          </div>
        ) : null}
        <Widget className={styles.widget} contentLayout="grid">
          <Toggle
            className={styles.columnFull}
            checked={state.enabled}
            onChange={onChangeEnabled}
            label={renderToString("SMTPConfigurationScreen.enable.label")}
            inlineLabel={true}
            disabled={state.isPasswordMasked || isCustomSMTPDisabled}
          />
          {state.enabled ? (
            <>
              <ProviderCard
                className={styles.columnLeft}
                onClick={onClickProviderSendgrid}
                isSelected={state.providerType === ProviderType.Sendgrid}
                disabled={state.isPasswordMasked}
                logoSrc={logoSendgrid}
              >
                <FormattedMessage id="SMTPConfigurationScreen.provider.sendgrid" />
              </ProviderCard>
              <ProviderCard
                className={styles.columnRight}
                onClick={onClickProviderCustom}
                isSelected={state.providerType === ProviderType.Custom}
                disabled={state.isPasswordMasked}
                iconProps={CUSTOM_PROVIDER_ICON_PROPS}
              >
                <FormattedMessage id="SMTPConfigurationScreen.provider.custom" />
              </ProviderCard>
              {form.state.providerType === ProviderType.Custom ? (
                <>
                  <ProviderCardDescription>
                    <FormattedMessage
                      id="SMTPConfigurationScreen.custom.description"
                      values={{
                        DocLink: (chunks: React.ReactNode) => (
                          <ExternalLink href="https://docs.authgear.com/customization/custom-providers/custom-email-provider">
                            {chunks}
                          </ExternalLink>
                        ),
                      }}
                    />
                  </ProviderCardDescription>
                  <FormTextField
                    className={styles.columnLeft}
                    type="text"
                    label={renderToString("SMTPConfigurationScreen.host.label")}
                    value={state.customHost}
                    disabled={state.isPasswordMasked}
                    required={true}
                    onChange={onStringChangeCallbacks.customHost}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="host"
                    {...hostProps}
                  />
                  <FormTextField
                    className={styles.columnLeft}
                    type="text"
                    label={renderToString("SMTPConfigurationScreen.port.label")}
                    value={state.customPortString}
                    disabled={state.isPasswordMasked}
                    required={true}
                    onChange={onCustomPortChange}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="port"
                    {...portProps}
                  />
                  <FormTextField
                    className={styles.columnLeft}
                    type="text"
                    label={renderToString(
                      "SMTPConfigurationScreen.username.label"
                    )}
                    value={state.customUsername}
                    disabled={state.isPasswordMasked}
                    required={true}
                    onChange={onStringChangeCallbacks.customUsername}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="username"
                  />
                  <FormTextField
                    className={styles.columnLeft}
                    type="password"
                    label={renderToString(
                      "SMTPConfigurationScreen.password.label"
                    )}
                    value={
                      state.isPasswordMasked
                        ? MASKED_PASSWORD_VALUE
                        : state.customPassword
                    }
                    disabled={state.isPasswordMasked}
                    required={true}
                    onChange={onStringChangeCallbacks.customPassword}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="password"
                  />
                  <FormTextField
                    className={styles.columnLeft}
                    type="text"
                    label={renderToString(
                      "SMTPConfigurationScreen.senderName.label"
                    )}
                    value={state.customSenderName}
                    disabled={state.isPasswordMasked}
                    onChange={onStringChangeCallbacks.customSenderName}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    /* Otherwise, the field is registered twice, and the error will be shown twice. */
                    /* Luckily, this field will not have any error so we can work around this way. */
                    fieldName="__THIS_IS_INTENTIONALLY_CHANGED_TO_A_NONEXISTENT_FIELD_NAME__"
                  />
                  <FormTextField
                    className={styles.columnLeft}
                    type="text"
                    label={renderToString(
                      "SMTPConfigurationScreen.senderAddress.label"
                    )}
                    value={state.customSenderAddress}
                    disabled={state.isPasswordMasked}
                    required={true}
                    onChange={onStringChangeCallbacks.customSenderAddress}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="sender"
                  />
                </>
              ) : null}
              {form.state.providerType === ProviderType.Sendgrid ? (
                <>
                  <ProviderCardDescription>
                    <FormattedMessage
                      id="SMTPConfigurationScreen.sendgrid.description"
                      values={{
                        DocLink: (chunks: React.ReactNode) => (
                          <ExternalLink href="https://docs.authgear.com/customization/custom-providers/custom-email-provider">
                            {chunks}
                          </ExternalLink>
                        ),
                      }}
                    />
                  </ProviderCardDescription>
                  <FormTextField
                    className={styles.columnLeft}
                    type="password"
                    label={renderToString(
                      "SMTPConfigurationScreen.sendgrid.api-key.label"
                    )}
                    value={
                      state.isPasswordMasked
                        ? MASKED_PASSWORD_VALUE
                        : state.sendgridAPIKey
                    }
                    required={true}
                    disabled={state.isPasswordMasked}
                    onChange={onStringChangeCallbacks.sendgridAPIKey}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="password"
                    {...sendgridAPIKeyProps}
                  />

                  <FormTextField
                    className={styles.columnLeft}
                    type="text"
                    label={renderToString(
                      "SMTPConfigurationScreen.senderName.label"
                    )}
                    value={state.sendgridSenderName}
                    disabled={state.isPasswordMasked}
                    onChange={onStringChangeCallbacks.sendgridSenderName}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="sender"
                  />
                  <FormTextField
                    className={styles.columnLeft}
                    type="text"
                    label={renderToString(
                      "SMTPConfigurationScreen.senderAddress.label"
                    )}
                    value={state.sendgridSenderAddress}
                    disabled={state.isPasswordMasked}
                    required={true}
                    onChange={onStringChangeCallbacks.sendgridSenderAddress}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="sender"
                  />
                </>
              ) : null}
              {state.isPasswordMasked ? (
                <PrimaryButton
                  className={styles.columnSmall}
                  disabled={isCustomSMTPDisabled}
                  onClick={onClickEdit}
                  text={<FormattedMessage id="edit" />}
                />
              ) : (
                <DefaultButton
                  className={styles.columnSmall}
                  onClick={onClickSendTestEmail}
                  disabled={!openSendTestEmailDialogButtonEnabled}
                  text={
                    <FormattedMessage id="SMTPConfigurationScreen.send-test-email.label" />
                  }
                />
              )}
              <Dialog
                hidden={isDialogHidden}
                onDismiss={onDismissDialog}
                dialogContentProps={dialogContentProps}
              >
                <TextField
                  type="email"
                  placeholder="user@example.com"
                  value={toAddress}
                  required={true}
                  onChange={onChangeToAddress}
                />
                <DialogFooter>
                  <PrimaryButton
                    onClick={onClickSend}
                    disabled={!sendTestEmailButtonEnabled || loading}
                    text={<FormattedMessage id="send" />}
                  />
                  <DefaultButton
                    onClick={onDismissDialog}
                    disabled={loading}
                    text={<FormattedMessage id="cancel" />}
                  />
                </DialogFooter>
              </Dialog>
            </>
          ) : null}
        </Widget>
      </ScreenContent>
    );
  };

const SMTPConfigurationScreen1: React.VFC<{
  appID: string;
  secretToken: string | null;
}> = function SMTPConfigurationScreen1({ appID, secretToken }) {
  const configQuery = useAppAndSecretConfigQuery(appID, secretToken);
  const configForm = useAppSecretConfigForm({
    appID,
    secretVisitToken: secretToken,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });
  const featureConfig = useAppFeatureConfigQuery(appID);

  const sendTestEmailHandle = useSendTestEmailMutation(appID);

  const state = useMemo<FormState>(() => {
    return {
      ...configForm.state,
      isSMTPRequiredForSomeEnabledFeatures:
        // primary authentication uses email OTP.
        configQuery.effectiveAppConfig?.authentication?.primary_authenticators?.includes(
          "oob_otp_email"
        ) === true ||
        // secondary authentication uses email OTP and secondary authentication is enabled.
        (configQuery.effectiveAppConfig?.authentication?.secondary_authenticators?.includes(
          "oob_otp_email"
        ) === true &&
          (configQuery.effectiveAppConfig.authentication
            .secondary_authentication_mode === "if_exists" ||
            configQuery.effectiveAppConfig.authentication
              .secondary_authentication_mode === "required")) ||
        configQuery.effectiveAppConfig?.verification?.claims?.email?.enabled ===
        true,
      smtpConfigured: configQuery.secretConfig?.smtpSecret != null,
    };
  }, [
    configForm.state,
    configQuery.effectiveAppConfig?.authentication?.primary_authenticators,
    configQuery.effectiveAppConfig?.authentication
      ?.secondary_authentication_mode,
    configQuery.effectiveAppConfig?.authentication?.secondary_authenticators,
    configQuery.effectiveAppConfig?.verification?.claims?.email?.enabled,
    configQuery.secretConfig?.smtpSecret,
  ]);

  const form: AppSecretConfigFormModel<FormState> = {
    isLoading:
      configQuery.isLoading || configForm.isLoading || featureConfig.isLoading,
    isUpdating: configForm.isUpdating,
    isDirty: configForm.isDirty,
    loadError:
      configQuery.loadError ??
      (configForm.loadError || featureConfig.loadError),
    updateError: configForm.updateError,
    state,
    setState: (fn) => {
      const newState = fn(state);
      configForm.setState(() => ({
        ...newState,
      }));
    },
    reload: () => {
      configForm.reload();
      featureConfig.refetch().finally(() => { });
    },
    reset: () => {
      configForm.reset();
    },
    save: async (ignoreConflict: boolean = false) => {
      await configForm.save(ignoreConflict);
    },
    saveWithState: async (
      state: FormState,
      ignoreConflict: boolean = false
    ) => {
      await configForm.saveWithState(state, ignoreConflict);
    },
  };

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return (
      <ShowError
        error={form.loadError}
        onRetry={() => {
          form.reload();
        }}
      />
    );
  }

  return (
    <FormContainer
      form={form}
      errorRules={ERROR_RULES}
      localError={sendTestEmailHandle.error}
    >
      <SMTPConfigurationScreenContent
        isCustomSMTPDisabled={
          featureConfig.effectiveFeatureConfig?.messaging
            ?.custom_smtp_disabled ?? false
        }
        sendTestEmailHandle={sendTestEmailHandle}
        form={form}
      />
    </FormContainer>
  );
};

const SECRETS = [AppSecretKey.SmtpSecret];

const SMTPConfigurationScreen: React.VFC = function SMTPConfigurationScreen() {
  const { appID } = useParams() as { appID: string };
  const location = useLocation();
  const [shouldRefreshToken] = useState<boolean>(() => {
    const { state } = location;
    if (isLocationState(state) && state.isEdit) {
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

  return <SMTPConfigurationScreen1 appID={appID} secretToken={token} />;
};

export default SMTPConfigurationScreen;
