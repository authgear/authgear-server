import React, { useCallback, useContext, useState, useMemo } from "react";
import { useLocation, useParams, useNavigate } from "react-router-dom";
import { produce } from "immer";
import { Dialog, DialogFooter } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
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
import { AppSecretKey } from "./globalTypes.generated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import {
  ProviderCard,
  ProviderCardDescription,
} from "../../components/common/ProviderCard";

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

interface FormState {
  enabled: boolean;
  providerType: ProviderType;
  sendgridAPIKey: string;

  customHost: string;
  customPortString: string;
  customUsername: string;
  customPassword: string;

  isPasswordMasked: boolean;
}

function constructFormState(
  _config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  const enabled = secrets.smtpSecret != null;

  const isSendgrid =
    secrets.smtpSecret != null &&
    secrets.smtpSecret.host === SENDGRID_HOST &&
    String(secrets.smtpSecret.port) === SENDGRID_PORT_STRING &&
    secrets.smtpSecret.username === SENDGRID_USERNAME;
  const providerType = isSendgrid ? ProviderType.Sendgrid : ProviderType.Custom;
  const isPasswordMasked = enabled && secrets.smtpSecret?.password == null;

  let sendgridAPIKey = "";
  let customHost = "";
  let customPortString = "";
  let customUsername = "";
  let customPassword = "";

  switch (providerType) {
    case ProviderType.Sendgrid:
      sendgridAPIKey = secrets.smtpSecret?.password ?? "";
      break;
    case ProviderType.Custom:
      customHost = secrets.smtpSecret?.host ?? "";
      customPortString =
        secrets.smtpSecret?.port != null ? String(secrets.smtpSecret.port) : "";
      customUsername = secrets.smtpSecret?.username ?? "";
      customPassword = secrets.smtpSecret?.password ?? "";
      break;

    default:
      break;
  }
  return {
    enabled,
    providerType,
    sendgridAPIKey,
    customHost,
    customPortString,
    customUsername,
    customPassword,
    isPasswordMasked,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _initialState: FormState,
  currentState: FormState,
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
  _currentState: FormState
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
      },
    },
  };
}

const CUSTOM_PROVIDER_ICON_PROPS = {
  iconName: "Mail",
};

interface SMTPConfigurationScreenContentProps {
  isCustomSMTPDisabled: boolean;
  sendTestEmailHandle: UseSendTestEmailMutationReturnType;
  form: AppSecretConfigFormModel<FormState>;
}

const SMTPConfigurationScreenContent: React.VFC<SMTPConfigurationScreenContentProps> =
  function SMTPConfigurationScreenContent(props) {
    const { form, sendTestEmailHandle, isCustomSMTPDisabled } = props;
    const { state, setState } = form;
    const { sendTestEmail, loading } = sendTestEmailHandle;

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
          return state.sendgridAPIKey !== "";
        case ProviderType.Custom:
          return (
            state.customHost !== "" &&
            state.customPortString !== "" &&
            state.customUsername !== "" &&
            state.customPassword !== ""
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
          | "customHost"
          | "customUsername"
          | "customPassword"
      ) => {
        return (_: unknown, value?: string) => {
          if (value != null) {
            setState((state) => {
              return {
                ...state,
                [key]: value,
              };
            });
          }
        };
      };
      return {
        sendgridAPIKey: callbackFactory("sendgridAPIKey"),
        customHost: callbackFactory("customHost"),
        customUsername: callbackFactory("customUsername"),
        customPassword: callbackFactory("customPassword"),
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
              to: toAddress,
            };
            break;
          case ProviderType.Custom:
            input = {
              smtpHost: state.customHost,
              smtpPort: parseInt(state.customPortString, 10),
              smtpUsername: state.customUsername,
              smtpPassword: state.customPassword,
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
          <FormattedMessage id="SMTPConfigurationScreen.title" />
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
                    <FormattedMessage id="SMTPConfigurationScreen.custom.description" />
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
                </>
              ) : null}
              {form.state.providerType === ProviderType.Sendgrid ? (
                <>
                  <ProviderCardDescription>
                    <FormattedMessage id="SMTPConfigurationScreen.sendgrid.description" />
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
  const form = useAppSecretConfigForm({
    appID,
    secretVisitToken: secretToken,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });
  const featureConfig = useAppFeatureConfigQuery(appID);

  const sendTestEmailHandle = useSendTestEmailMutation(appID);

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
    <FormContainer form={form} localError={sendTestEmailHandle.error}>
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
