import cn from "classnames";
import React, { useCallback, useContext, useState, useMemo, useRef } from "react";
import { useLocation, useParams, useNavigate } from "react-router-dom";
import { produce } from "immer";
import { Dialog, DialogFooter } from "@fluentui/react";
import { FormattedMessage, Context } from "../../intl";
import { parseSender } from "email-addresses";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import FormContainer from "../../FormContainer";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import ScreenContent from "../../ScreenContent";
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
import logoSendgrid from "../../images/sendgrid_logo.svg";
import logoAuthgear from "../../images/authgear_logo.svg";
import styles from "./SMTPConfigurationScreen.module.css";
import ExternalLink from "../../ExternalLink";
import { AppSecretKey } from "./globalTypes.generated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import { ErrorParseRule, ErrorParseRuleResult } from "../../error/parse";
import { APIError, APISMTPTestFailedError } from "../../error/error";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { RedMessageBar_RemindConfigureSMTPInSMTPConfigurationScreen } from "../../RedMessageBar";
import { useCalloutToast } from "../../components/v2/Callout/Callout";
import {
  IconRadioCards,
  IconRadioCardOption,
} from "../../components/v2/IconRadioCards/IconRadioCards";
import { TextField } from "../../components/v2/TextField/TextField";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { Text } from "@radix-ui/themes";
import { EnvelopeClosedIcon, EyeNoneIcon, EyeOpenIcon } from "@radix-ui/react-icons";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useFormContainerBaseContext } from "../../FormContainerBase";

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
  Authgear = "authgear",
  Sendgrid = "sendgrid",
  Custom = "custom",
}

const MASKED_PASSWORD_VALUE = "****************";

const SENDGRID_HOST = "smtp.sendgrid.net";
const SENDGRID_PORT_STRING = "587";
const SENDGRID_USERNAME = "apikey";

// Matches v2 IconRadioCards storybook inner icon size (SquareIcon iconSize).
const PROVIDER_RADIO_ICON_SIZE = "1.375rem";

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
  const providerType = !enabled
    ? ProviderType.Authgear
    : isSendgrid
      ? ProviderType.Sendgrid
      : ProviderType.Custom;
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
    const { showToast } = useCalloutToast();

    const [isDialogHidden, setIsDialogHidden] = useState(true);
    const [toAddress, setToAddress] = useState("");
    const [showPassword, setShowPassword] = useState(false);
    const { viewer } = useViewerQuery();
    const { renderToString } = useContext(Context);
    const { isDirty } = useFormContainerBaseContext();
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    const onChangeProviderType = useCallback(
      (value: ProviderType) => {
        setState((s) => ({
          ...s,
          enabled: value !== ProviderType.Authgear,
          providerType: value,
        }));
      },
      [setState]
    );

    const fieldCallbacks = useMemo(() => {
      const make =
        (
          key:
            | "sendgridAPIKey"
            | "sendgridSenderName"
            | "sendgridSenderAddress"
            | "customHost"
            | "customUsername"
            | "customPassword"
            | "customSenderName"
            | "customSenderAddress"
        ) =>
        (e: React.ChangeEvent<HTMLInputElement>) => {
          const value = e.target.value;
          setState((s) => {
            const next: FormState = { ...s };
            next[key] = value;
            return next;
          });
        };
      return {
        sendgridAPIKey: make("sendgridAPIKey"),
        sendgridSenderName: make("sendgridSenderName"),
        sendgridSenderAddress: make("sendgridSenderAddress"),
        customHost: make("customHost"),
        customUsername: make("customUsername"),
        customPassword: make("customPassword"),
        customSenderName: make("customSenderName"),
        customSenderAddress: make("customSenderAddress"),
      };
    }, [setState]);

    const onCustomPortChange = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        if (value === "" || !isNaN(Number(value))) {
          setState((s) => ({ ...s, customPortString: value }));
        }
      },
      [setState]
    );

    const onToggleShowPassword = useCallback(() => {
      setShowPassword((v) => !v);
    }, []);

    const navigate = useNavigate();

    const onClickEdit = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();

        const locationState: LocationState = {
          isEdit: true,
        };

        startReauthentication(navigate, locationState).catch((e) => {
          console.error(e);
        });
      },
      [navigate]
    );

    const openSendTestEmailDialogButtonEnabled = useMemo(() => {
      switch (state.providerType) {
        case ProviderType.Authgear:
          return false;
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
        default:
          return false;
      }
    }, [state]);

    const sendTestEmailButtonEnabled = useMemo(() => toAddress !== "", [toAddress]);

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
          case ProviderType.Authgear:
            return;
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
          default:
            return;
        }

        sendTestEmail(input).then(
          () => {
            showToast({
              type: "success",
              text: (
                <FormattedMessage id="SMTPConfigurationScreen.send-test-email.toast.success" />
              ),
            });
            setIsDialogHidden(true);
          },
          () => {
            setIsDialogHidden(true);
          }
        );
      },
      [sendTestEmail, showToast, state, toAddress]
    );

    const onChangeToAddress = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        setToAddress(e.target.value);
      },
      []
    );

    const dialogContentProps = useMemo(
      () => ({
        title: renderToString(
          "SMTPConfigurationScreen.send-test-email-dialog.title"
        ),
        subText: renderToString(
          "SMTPConfigurationScreen.send-test-email-dialog.description"
        ),
      }),
      [renderToString]
    );

    const providerOptions = useMemo(
      (): IconRadioCardOption<ProviderType>[] => [
        {
          value: ProviderType.Authgear,
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
            <FormattedMessage id="SMTPConfigurationScreen.provider.authgear" />
          ),
          disabled: state.isPasswordMasked,
        },
        {
          value: ProviderType.Sendgrid,
          icon: (
            <img
              src={logoSendgrid}
              alt=""
              className="object-contain"
              style={{
                width: PROVIDER_RADIO_ICON_SIZE,
                height: PROVIDER_RADIO_ICON_SIZE,
              }}
            />
          ),
          title: (
            <FormattedMessage id="SMTPConfigurationScreen.provider.sendgrid" />
          ),
          disabled: state.isPasswordMasked || isCustomSMTPDisabled,
        },
        {
          value: ProviderType.Custom,
          icon: (
            <EnvelopeClosedIcon
              width={PROVIDER_RADIO_ICON_SIZE}
              height={PROVIDER_RADIO_ICON_SIZE}
            />
          ),
          title: (
            <FormattedMessage id="SMTPConfigurationScreen.provider.custom" />
          ),
          disabled: state.isPasswordMasked || isCustomSMTPDisabled,
        },
      ],
      [isCustomSMTPDisabled, state.isPasswordMasked]
    );

    const providerDescription = useMemo(() => {
      switch (state.providerType) {
        case ProviderType.Authgear:
          return (
            <FormattedMessage id="SMTPConfigurationScreen.authgear.description" />
          );
        case ProviderType.Sendgrid:
          return (
            <FormattedMessage
              id="SMTPConfigurationScreen.sendgrid.description"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                DocLink: (chunks: React.ReactNode) => (
                  <ExternalLink href="https://docs.authgear.com/customization/custom-providers/custom-email-provider">
                    {chunks}
                  </ExternalLink>
                ),
              }}
            />
          );
        case ProviderType.Custom:
          return (
            <FormattedMessage
              id="SMTPConfigurationScreen.custom.description"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                DocLink: (chunks: React.ReactNode) => (
                  <ExternalLink href="https://docs.authgear.com/customization/custom-providers/custom-email-provider">
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

    const showSettings =
      state.providerType === ProviderType.Sendgrid ||
      state.providerType === ProviderType.Custom;

    return (
      <ScreenContent
        className={cn(isDirty ? styles.contentWithSaveBar : null)}
      >
        <div
          ref={contentWidthAnchorRef}
          className={styles.contentWidthAnchor}
          aria-hidden
        />
        <div className={cn(styles.widget, styles.pageHeader)}>
          <h1 className={styles.pageTitle}>
            {isAuthgearOnce ? (
              <FormattedMessage id="SMTPConfigurationScreen.title--authgearonce" />
            ) : (
              <FormattedMessage id="SMTPConfigurationScreen.title" />
            )}
          </h1>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="SMTPConfigurationScreen.description" />
          </Text>
        </div>
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

        <div className={cn(styles.widget, styles.providerSelector)}>
          <IconRadioCards
            size="3"
            value={state.providerType}
            onValueChange={onChangeProviderType}
            options={providerOptions}
            numberOfColumns={3}
            itemFillSpaces={true}
          />
          {providerDescription != null ? (
            <Text as="p" size="1" color="gray" className={styles.providerDescription}>
              {providerDescription}
            </Text>
          ) : null}
        </div>

        {showSettings ? (
          <div
            className={cn(
              styles.widget,
              "border border-[var(--gray-5)] rounded-lg p-6 flex gap-8 bg-white"
            )}
          >
            <Text
              as="p"
              size="3"
              weight="medium"
              className="shrink-0 w-[200px]"
            >
              <FormattedMessage id="SMTPConfigurationScreen.settings.label" />
            </Text>
            <div className="flex-1 flex flex-col gap-4 min-w-0">
              {state.providerType === ProviderType.Sendgrid ? (
                <>
                  <TextField
                    size="2"
                    labelSize="2"
                    type="password"
                    label={
                      <FormattedMessage id="SMTPConfigurationScreen.sendgrid.api-key.label" />
                    }
                    hint={renderToString(
                      "SMTPConfigurationScreen.sendgrid.api-key.tooltip"
                    )}
                    value={
                      state.isPasswordMasked
                        ? MASKED_PASSWORD_VALUE
                        : state.sendgridAPIKey
                    }
                    disabled={state.isPasswordMasked}
                    onChange={fieldCallbacks.sendgridAPIKey}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="password"
                  />
                  <TextField
                    size="2"
                    labelSize="2"
                    label={
                      <FormattedMessage id="SMTPConfigurationScreen.senderName.label" />
                    }
                    value={state.sendgridSenderName}
                    disabled={state.isPasswordMasked}
                    onChange={fieldCallbacks.sendgridSenderName}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="__THIS_IS_INTENTIONALLY_CHANGED_TO_A_NONEXISTENT_FIELD_NAME__"
                  />
                  <TextField
                    size="2"
                    labelSize="2"
                    label={
                      <FormattedMessage id="SMTPConfigurationScreen.senderAddress.label" />
                    }
                    value={state.sendgridSenderAddress}
                    disabled={state.isPasswordMasked}
                    onChange={fieldCallbacks.sendgridSenderAddress}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="sender"
                  />
                </>
              ) : null}
              {state.providerType === ProviderType.Custom ? (
                <>
                  <TextField
                    size="2"
                    labelSize="2"
                    label={
                      <FormattedMessage id="SMTPConfigurationScreen.host.label" />
                    }
                    hint={renderToString(
                      "SMTPConfigurationScreen.host.tooltip"
                    )}
                    value={state.customHost}
                    disabled={state.isPasswordMasked}
                    onChange={fieldCallbacks.customHost}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="host"
                  />
                  <TextField
                    size="2"
                    labelSize="2"
                    label={
                      <FormattedMessage id="SMTPConfigurationScreen.port.label" />
                    }
                    hint={renderToString(
                      "SMTPConfigurationScreen.port.tooltip"
                    )}
                    value={state.customPortString}
                    disabled={state.isPasswordMasked}
                    onChange={onCustomPortChange}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="port"
                  />
                  <TextField
                    size="2"
                    labelSize="2"
                    label={
                      <FormattedMessage id="SMTPConfigurationScreen.username.label" />
                    }
                    value={state.customUsername}
                    disabled={state.isPasswordMasked}
                    onChange={fieldCallbacks.customUsername}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="username"
                  />
                  <TextField
                    size="2"
                    labelSize="2"
                    type={showPassword ? "text" : "password"}
                    label={
                      <FormattedMessage id="SMTPConfigurationScreen.password.label" />
                    }
                    value={
                      state.isPasswordMasked
                        ? MASKED_PASSWORD_VALUE
                        : state.customPassword
                    }
                    disabled={state.isPasswordMasked}
                    onChange={fieldCallbacks.customPassword}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="password"
                    suffixPlain
                    suffix={
                      !state.isPasswordMasked ? (
                        <button
                          type="button"
                          onClick={onToggleShowPassword}
                          className="flex items-center text-[var(--gray-9)]"
                        >
                          {showPassword ? <EyeNoneIcon /> : <EyeOpenIcon />}
                        </button>
                      ) : undefined
                    }
                  />
                  <TextField
                    size="2"
                    labelSize="2"
                    label={
                      <FormattedMessage id="SMTPConfigurationScreen.senderName.label" />
                    }
                    value={state.customSenderName}
                    disabled={state.isPasswordMasked}
                    onChange={fieldCallbacks.customSenderName}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="__THIS_IS_INTENTIONALLY_CHANGED_TO_A_NONEXISTENT_FIELD_NAME__"
                  />
                  <TextField
                    size="2"
                    labelSize="2"
                    label={
                      <FormattedMessage id="SMTPConfigurationScreen.senderAddress.label" />
                    }
                    value={state.customSenderAddress}
                    disabled={state.isPasswordMasked}
                    onChange={fieldCallbacks.customSenderAddress}
                    parentJSONPointer={/\/secrets\/\d+\/data/}
                    fieldName="sender"
                  />
                </>
              ) : null}
              {state.isPasswordMasked ? (
                <div>
                  <PrimaryButton
                    size="3"
                    disabled={isCustomSMTPDisabled}
                    onClick={onClickEdit}
                    text={<FormattedMessage id="edit" />}
                  />
                </div>
              ) : (
                <div>
                  <SecondaryButton
                    size="2"
                    onClick={onClickSendTestEmail}
                    disabled={!openSendTestEmailDialogButtonEnabled}
                    text={
                      <FormattedMessage id="SMTPConfigurationScreen.send-test-email.label" />
                    }
                  />
                </div>
              )}
            </div>
          </div>
        ) : null}

        <Dialog
          hidden={isDialogHidden}
          onDismiss={onDismissDialog}
          dialogContentProps={dialogContentProps}
        >
          <TextField
            size="3"
            type="email"
            placeholder="user@example.com"
            value={toAddress}
            onChange={onChangeToAddress}
          />
          <DialogFooter>
            <PrimaryButton
              size="3"
              onClick={onClickSend}
              disabled={!sendTestEmailButtonEnabled || loading}
              text={<FormattedMessage id="send" />}
            />
            <SecondaryButton
              size="3"
              onClick={onDismissDialog}
              disabled={loading}
              text={<FormattedMessage id="cancel" />}
            />
          </DialogFooter>
        </Dialog>
        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
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
      featureConfig.refetch().finally(() => {});
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
      hideFooterComponent
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
