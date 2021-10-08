import React, { useCallback, useContext, useState, useMemo } from "react";
import { useParams } from "react-router-dom";
import produce from "immer";
import {
  Text,
  FontIcon,
  Image,
  Label,
  Toggle,
  PrimaryButton,
  DefaultButton,
  DefaultEffects,
  Dialog,
  DialogFooter,
  TextField,
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
import { startReauthentication } from "./Authenticated";
import { PortalAPIAppConfig, PortalAPISecretConfig } from "../../types";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useViewerQuery } from "./query/viewerQuery";
import {
  useSendTestEmailMutation,
  UseSendTestEmailMutationReturnType,
} from "./mutations/sendTestEmail";
import logoSendgrid from "../../images/sendgrid_logo.png";
import styles from "./SMTPConfigurationScreen.module.scss";

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

interface ProviderCardProps {
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
  const { disabled, isSelected, children, onClick, iconProps, logoSrc } = props;
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
      className={styles.providerCard}
      onClick={disabled ? undefined : onClick}
    >
      {iconProps != null && (
        <FontIcon {...iconProps} style={PROVIDER_CARD_ICON_STYLE} />
      )}
      {logoSrc != null && <Image src={logoSrc} width={32} height={32} />}
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
      className={styles.control}
      variant="small"
      block={true}
      style={{
        color,
      }}
    >
      {children}
    </Text>
  );
}

interface SMTPConfigurationScreenContentProps {
  sendTestEmailHandle: UseSendTestEmailMutationReturnType;
  form: AppSecretConfigFormModel<FormState>;
}

const SMTPConfigurationScreenContent: React.FC<SMTPConfigurationScreenContentProps> =
  function SMTPConfigurationScreenContent(props) {
    const { form, sendTestEmailHandle } = props;
    const { state, setState } = form;
    const { sendTestEmail, loading } = sendTestEmailHandle;

    const [isDialogHidden, setIsDialogHidden] = useState(true);
    const [toAddress, setToAddress] = useState("");
    const { viewer } = useViewerQuery();
    const { renderToString } = useContext(Context);

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
      <ScreenContent className={styles.root}>
        <ScreenTitle>
          <FormattedMessage id="SMTPConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="SMTPConfigurationScreen.description" />
        </ScreenDescription>

        <Widget className={styles.widget}>
          <Toggle
            className={styles.control}
            checked={state.enabled}
            onChange={onChangeEnabled}
            label={renderToString("SMTPConfigurationScreen.enable.label")}
            inlineLabel={true}
            disabled={state.isPasswordMasked}
          />
          {state.enabled && (
            <>
              <div className={styles.providerCards}>
                <ProviderCard
                  onClick={onClickProviderSendgrid}
                  isSelected={providerType === "sendgrid"}
                  disabled={state.isPasswordMasked}
                  logoSrc={logoSendgrid}
                >
                  <FormattedMessage id="SMTPConfigurationScreen.provider.sendgrid" />
                </ProviderCard>
                <ProviderCard
                  onClick={onClickProviderCustom}
                  isSelected={providerType === "custom"}
                  disabled={state.isPasswordMasked}
                  iconProps={CUSTOM_PROVIDER_ICON_PROPS}
                >
                  <FormattedMessage id="SMTPConfigurationScreen.provider.custom" />
                </ProviderCard>
              </div>
              {providerType === "custom" && (
                <>
                  <ProviderDescription>
                    <FormattedMessage id="SMTPConfigurationScreen.custom.description" />
                  </ProviderDescription>
                  <FormTextField
                    className={styles.control}
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
                    className={styles.control}
                    type="number"
                    min="1"
                    step="1"
                    max="65535"
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
                    className={styles.control}
                    type="text"
                    label={renderToString(
                      "SMTPConfigurationScreen.username.label"
                    )}
                    value={state.username}
                    disabled={state.isPasswordMasked}
                    onChange={onChangeUsername}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="username"
                  />
                  <FormTextField
                    className={styles.control}
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
                    onChange={onChangePassword}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="password"
                  />
                </>
              )}
              {providerType === "sendgrid" && (
                <>
                  <ProviderDescription>
                    <FormattedMessage id="SMTPConfigurationScreen.sendgrid.description" />
                  </ProviderDescription>
                  <FormTextField
                    className={styles.control}
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
              )}
              {state.isPasswordMasked ? (
                <PrimaryButton className={styles.control} onClick={onClickEdit}>
                  <FormattedMessage id="edit" />
                </PrimaryButton>
              ) : (
                <DefaultButton
                  className={styles.control}
                  onClick={onClickSendTestEmail}
                >
                  <FormattedMessage id="SMTPConfigurationScreen.send-test-email.label" />
                </DefaultButton>
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
                  <PrimaryButton onClick={onClickSend} disabled={loading}>
                    <FormattedMessage id="send" />
                  </PrimaryButton>
                  <DefaultButton onClick={onDismissDialog} disabled={loading}>
                    <FormattedMessage id="cancel" />
                  </DefaultButton>
                </DialogFooter>
              </Dialog>
            </>
          )}
        </Widget>
      </ScreenContent>
    );
  };

const SMTPConfigurationScreen: React.FC = function SMTPConfigurationScreen() {
  const { appID } = useParams();
  const form = useAppSecretConfigForm(
    appID,
    constructFormState,
    constructConfig
  );

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
