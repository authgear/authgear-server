import React, { useCallback, useContext, useState, useMemo } from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
import produce from "immer";
import {
  Text,
  FontIcon,
  Image,
  Label,
  DefaultEffects,
  Dialog,
  DialogFooter,
  IIconProps,
  IButtonProps,
} from "@fluentui/react";
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
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useViewerQuery } from "./query/viewerQuery";
import {
  useSendTestEmailMutation,
  UseSendTestEmailMutationReturnType,
} from "./mutations/sendTestEmail";
import logoSendgrid from "../../images/sendgrid_logo.png";
import styles from "./SMTPConfigurationScreen.module.css";
import PrimaryButton from "../../PrimaryButton";
import DefaultButton from "../../DefaultButton";

type ProviderType = "sendgrid" | "custom";

const MASKED_PASSWORD_VALUE = "****************";

const SENDGRID_HOST = "smtp.sendgrid.net";
const SENDGRID_PORT_STRING = "587";
const SENDGRID_USERNAME = "apikey";

interface FormState {
  enabled: boolean;
  host: string;
  portString: string;
  username: string;
  password: string;
  isPasswordMasked: boolean;
}

function constructFormState(
  _config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  const enabled = secrets.smtpSecret != null;
  const host = secrets.smtpSecret?.host ?? "";
  const portString =
    secrets.smtpSecret?.port != null ? String(secrets.smtpSecret.port) : "";
  const username = secrets.smtpSecret?.username ?? "";
  const password = secrets.smtpSecret?.password ?? "";
  const isPasswordMasked = enabled && secrets.smtpSecret?.password == null;
  return {
    enabled,
    host,
    portString,
    username,
    password,
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
    } else if (currentState.isPasswordMasked) {
      secrets.smtpSecret = {
        host: currentState.host,
        port: Number(currentState.portString),
        username: currentState.username,
        password: null,
      };
    } else {
      secrets.smtpSecret = {
        host: currentState.host,
        port: Number(currentState.portString),
        username: currentState.username,
        password: currentState.password,
      };
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

interface ProviderCardProps {
  className?: string;
  iconProps?: IIconProps;
  logoSrc?: any;
  children?: React.ReactNode;
  onClick?: IButtonProps["onClick"];
  isSelected?: boolean;
  disabled?: boolean;
}

const PROVIDER_CARD_ICON_STYLE = {
  width: "32px",
  height: "32px",
  fontSize: "32px",
};

const CUSTOM_PROVIDER_ICON_PROPS = {
  iconName: "Mail",
};

function ProviderCard(props: ProviderCardProps) {
  const {
    className,
    disabled,
    isSelected,
    children,
    onClick,
    iconProps,
    logoSrc,
  } = props;
  const {
    themes: {
      main: {
        palette: { themePrimary },
        semanticColors: { disabledBackground: backgroundColor },
      },
    },
  } = useSystemConfig();
  return (
    <div
      style={{
        boxShadow: disabled ? undefined : DefaultEffects.elevation4,
        borderColor: isSelected ? themePrimary : "transparent",
        backgroundColor: disabled ? backgroundColor : undefined,
        cursor: disabled ? "not-allowed" : undefined,
      }}
      className={cn(className, styles.providerCard)}
      onClick={disabled ? undefined : onClick}
    >
      {iconProps != null ? (
        <FontIcon {...iconProps} style={PROVIDER_CARD_ICON_STYLE} />
      ) : null}
      {logoSrc != null ? <Image src={logoSrc} width={32} height={32} /> : null}
      <Label>{children}</Label>
    </div>
  );
}

interface ProviderDescriptionProps {
  children?: React.ReactNode;
}

function ProviderDescription(props: ProviderDescriptionProps) {
  const { children } = props;
  const {
    themes: {
      main: {
        semanticColors: { bodySubtext: color },
      },
    },
  } = useSystemConfig();

  return (
    <Text
      variant="small"
      block={true}
      style={{
        color,
      }}
      className={styles.columnFull}
    >
      {children}
    </Text>
  );
}

interface SMTPConfigurationScreenContentProps {
  sendTestEmailHandle: UseSendTestEmailMutationReturnType;
  form: AppSecretConfigFormModel<FormState>;
}

const SMTPConfigurationScreenContent: React.VFC<SMTPConfigurationScreenContentProps> =
  function SMTPConfigurationScreenContent(props) {
    const { form, sendTestEmailHandle } = props;
    const { state, setState } = form;
    const { sendTestEmail, loading } = sendTestEmailHandle;

    const [isDialogHidden, setIsDialogHidden] = useState(true);
    const [toAddress, setToAddress] = useState("");
    const { viewer } = useViewerQuery();
    const { renderToString } = useContext(Context);

    const openSendTestEmailDialogButtonEnabled = useMemo(() => {
      return (
        state.host !== "" &&
        state.portString !== "" &&
        state.username !== "" &&
        state.password !== ""
      );
    }, [state]);

    const sendTestEmailButtonEnabled = useMemo(() => {
      return toAddress !== "";
    }, [toAddress]);

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

    const onChangeHost = useCallback(
      (_, value?: string) => {
        if (value != null) {
          setState((state) => {
            return {
              ...state,
              host: value,
            };
          });
        }
      },
      [setState]
    );

    const onChangePort = useCallback(
      (_, value?: string) => {
        if (value != null) {
          if (value === "") {
            setState((state) => {
              return {
                ...state,
                portString: "",
              };
            });
          } else {
            const port = Number(value);
            if (!isNaN(port)) {
              setState((state) => {
                return {
                  ...state,
                  portString: value,
                };
              });
            }
          }
        }
      },
      [setState]
    );

    const onChangeUsername = useCallback(
      (_, value?: string) => {
        if (value != null) {
          setState((state) => {
            return {
              ...state,
              username: value,
            };
          });
        }
      },
      [setState]
    );

    const onChangePassword = useCallback(
      (_, value?: string) => {
        if (value != null) {
          setState((state) => {
            return {
              ...state,
              password: value,
            };
          });
        }
      },
      [setState]
    );

    const onClickEdit = useCallback((e: React.MouseEvent<unknown>) => {
      e.preventDefault();
      e.stopPropagation();

      startReauthentication().catch((e) => {
        // Normally there should not be any error.
        console.error(e);
      });
    }, []);

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

        sendTestEmail({
          smtpHost: state.host,
          smtpPort: parseInt(state.portString, 10),
          smtpUsername: state.username,
          smtpPassword: state.password,
          to: toAddress,
        }).then(
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

    const providerType: ProviderType = useMemo(() => {
      const isSendgrid =
        state.host === SENDGRID_HOST &&
        state.portString === SENDGRID_PORT_STRING &&
        state.username === SENDGRID_USERNAME;
      return isSendgrid ? "sendgrid" : "custom";
    }, [state]);

    const onClickProviderSendgrid = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();

        setState((state) => {
          return {
            ...state,
            host: SENDGRID_HOST,
            portString: SENDGRID_PORT_STRING,
            username: SENDGRID_USERNAME,
            password: "",
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
            host: "",
            portString: "",
            username: "",
            password: "",
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

        <Widget className={styles.widget} contentLayout="grid">
          <Toggle
            className={styles.columnFull}
            checked={state.enabled}
            onChange={onChangeEnabled}
            label={renderToString("SMTPConfigurationScreen.enable.label")}
            inlineLabel={true}
            disabled={state.isPasswordMasked}
          />
          {state.enabled ? (
            <>
              <ProviderCard
                className={styles.columnLeft}
                onClick={onClickProviderSendgrid}
                isSelected={providerType === "sendgrid"}
                disabled={state.isPasswordMasked}
                logoSrc={logoSendgrid}
              >
                <FormattedMessage id="SMTPConfigurationScreen.provider.sendgrid" />
              </ProviderCard>
              <ProviderCard
                className={styles.columnRight}
                onClick={onClickProviderCustom}
                isSelected={providerType === "custom"}
                disabled={state.isPasswordMasked}
                iconProps={CUSTOM_PROVIDER_ICON_PROPS}
              >
                <FormattedMessage id="SMTPConfigurationScreen.provider.custom" />
              </ProviderCard>
              {providerType === "custom" ? (
                <>
                  <ProviderDescription>
                    <FormattedMessage id="SMTPConfigurationScreen.custom.description" />
                  </ProviderDescription>
                  <FormTextField
                    className={styles.columnLeft}
                    type="text"
                    label={renderToString("SMTPConfigurationScreen.host.label")}
                    value={state.host}
                    disabled={state.isPasswordMasked}
                    required={true}
                    onChange={onChangeHost}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="host"
                    {...hostProps}
                  />
                  <FormTextField
                    className={styles.columnLeft}
                    type="text"
                    label={renderToString("SMTPConfigurationScreen.port.label")}
                    value={state.portString}
                    disabled={state.isPasswordMasked}
                    required={true}
                    onChange={onChangePort}
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
                    value={state.username}
                    disabled={state.isPasswordMasked}
                    required={true}
                    onChange={onChangeUsername}
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
                        : state.password
                    }
                    disabled={state.isPasswordMasked}
                    required={true}
                    onChange={onChangePassword}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="password"
                  />
                </>
              ) : null}
              {providerType === "sendgrid" ? (
                <>
                  <ProviderDescription>
                    <FormattedMessage id="SMTPConfigurationScreen.sendgrid.description" />
                  </ProviderDescription>
                  <FormTextField
                    className={styles.columnLeft}
                    type="password"
                    label={renderToString(
                      "SMTPConfigurationScreen.sendgrid.api-key.label"
                    )}
                    value={
                      state.isPasswordMasked
                        ? MASKED_PASSWORD_VALUE
                        : state.password
                    }
                    required={true}
                    disabled={state.isPasswordMasked}
                    onChange={onChangePassword}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="password"
                    {...sendgridAPIKeyProps}
                  />
                </>
              ) : null}
              {state.isPasswordMasked ? (
                <PrimaryButton
                  className={styles.columnSmall}
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

const SMTPConfigurationScreen: React.VFC = function SMTPConfigurationScreen() {
  const { appID } = useParams() as { appID: string };
  const form = useAppSecretConfigForm({
    appID,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });

  const sendTestEmailHandle = useSendTestEmailMutation(appID);

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form} localError={sendTestEmailHandle.error}>
      <SMTPConfigurationScreenContent
        sendTestEmailHandle={sendTestEmailHandle}
        form={form}
      />
    </FormContainer>
  );
};

export default SMTPConfigurationScreen;
