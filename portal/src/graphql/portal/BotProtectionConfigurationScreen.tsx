import { Context, FormattedMessage } from "@oursky/react-messageformat";
import React, { useCallback, useContext, useState } from "react";
import cn from "classnames";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import styles from "./BotProtectionConfigurationScreen.module.css";
import {
  BotProtectionProviderType,
  BotProtectionRiskMode,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
} from "../../types";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import { useLocation, useNavigate, useParams } from "react-router-dom";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { AppSecretKey } from "./globalTypes.generated";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { produce } from "immer";
import { clearEmptyObject } from "../../util/misc";
import FormContainer from "../../FormContainer";
import ScreenDescription from "../../ScreenDescription";
import Toggle from "../../Toggle";
import WidgetTitle from "../../WidgetTitle";
import { DefaultEffects, IButtonProps, Image, Label } from "@fluentui/react";
import { useSystemConfig } from "../../context/SystemConfigContext";
import recaptchaV2LogoURL from "../../images/recaptchav2_logo.svg";
import cloudflareLogoURL from "../../images/cloudflare_logo.svg";
import WidgetDescription from "../../WidgetDescription";
import FormTextField from "../../FormTextField";
import PrimaryButton from "../../PrimaryButton";
import { startReauthentication } from "./Authenticated";
import { useSessionStorage } from "../../hook/useSessionStorage";

const MASKED_SECRET = "***************";

const SECRET_KEY_FORM_FIELD_ID = "secret-key-form-field";

interface LocationState {
  isOAuthRedirect: boolean;
}
function isLocationState(raw: unknown): raw is LocationState {
  return (
    raw != null &&
    typeof raw === "object" &&
    (raw as Partial<LocationState>).isOAuthRedirect != null
  );
}

interface FormCloudflareConfigs {
  siteKey: string;
  secretKey: string | null;
}

interface FormRecaptchav2Configs {
  siteKey: string;
  secretKey: string | null;
}

type FormBotProtectionProviderConfigs = FormCloudflareConfigs | FormRecaptchav2Configs;


type FormBotProtectionRequirementsFlowsType = "allSignupLogin" | "specificAuthenticator"
interface FormBotProtectionRequirementsFlowsAllSignupLoginFlowConfigs {
  allSignupLoginMode: BotProtectionRiskMode;
}

interface FormBotProtectionRequirementsFlowsSpecificAuthenticatorFlowConfigs {
  passwordMode: BotProtectionRiskMode;
  passwordlessViaSMSMode: BotProtectionRiskMode;
  passwordlessViaEmailMode: BotProtectionRiskMode;
}
interface FormBotProtectionRequirementsFlows {
  flowType: FormBotProtectionRequirementsFlowsType;
  flowConfigs: {
    allSignupLogin: FormBotProtectionRequirementsFlowsAllSignupLoginFlowConfigs,
    specificAuthenticator: FormBotProtectionRequirementsFlowsSpecificAuthenticatorFlowConfigs
  }
}

interface FormBotProtectionRequirementsResetPassword {
  resetPasswordMode: BotProtectionRiskMode;
}

interface FormBotProtectionRequirements {
  flows: FormBotProtectionRequirementsFlows
  resetPassword: FormBotProtectionRequirementsResetPassword
}

interface FormState {
  enabled: boolean;
  providerType: BotProtectionProviderType;
  providerConfigs: Partial<
    Record<BotProtectionProviderType, FormBotProtectionProviderConfigs>
  >;
  requirements: FormBotProtectionRequirements;
}

function constructFormRequirementsState(
  config: PortalAPIAppConfig,
): FormBotProtectionRequirements {
  const requirements = config.bot_protection?.requirements
  // If any specific authenticator is configured, construct as specificAuthenticator, even if signup_or_login IS configured
  // otherwise, construct as allSignupLogin
  const isSpecificAuthenticatorConfigured = (
    requirements?.oob_otp_email != null ||
    requirements?.oob_otp_sms != null ||
    requirements?.password != null
  );
  const dominantFlowType: FormBotProtectionRequirementsFlowsType = isSpecificAuthenticatorConfigured ? "specificAuthenticator" : "allSignupLogin"
  const flowConfigs = {
    allSignupLogin: {
      allSignupLoginMode: requirements?.signup_or_login?.mode ?? "never",
    },
    specificAuthenticator: {
      passwordMode: requirements?.password?.mode ?? "never",
      passwordlessViaSMSMode: requirements?.oob_otp_sms?.mode ?? "never",
      passwordlessViaEmailMode: requirements?.oob_otp_email?.mode ?? "never",
    }
  }

  const flows: FormBotProtectionRequirementsFlows = {
    flowType: dominantFlowType,
    flowConfigs,
  };
  const resetPassword: FormBotProtectionRequirementsResetPassword = {
    resetPasswordMode: requirements?.account_recovery?.mode ?? "never",
  }
  return {
    flows,
    resetPassword,
  }
}

function constructFormState(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  const enabled = config.bot_protection?.enabled ?? false;
  const providerType: BotProtectionProviderType =
    config.bot_protection?.provider?.type ?? "recaptchav2";
  const siteKey = config.bot_protection?.provider?.site_key ?? "";
  const secretKey = secrets.botProtectionProviderSecret?.secretKey ?? null;
  const providerConfigs: Partial<
    Record<BotProtectionProviderType, FormBotProtectionProviderConfigs>
  > = {
    [providerType]: {
      siteKey,
      secretKey,
    },
  };
  const requirements = constructFormRequirementsState(config);

  return {
    enabled,
    providerType,
    providerConfigs,
    requirements,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  return produce([config, secrets], ([config, secrets]) => {
    config.bot_protection ??= {};
    config.bot_protection.provider ??= {};
    config.bot_protection = {
      enabled: currentState.enabled,
      provider: {
        type: currentState.providerType,
        site_key:
          currentState.providerConfigs[currentState.providerType]?.siteKey,
      },
    };

    const secretKey =
      currentState.providerConfigs[currentState.providerType]?.secretKey;
    if (secretKey != null) {
      secrets.botProtectionProviderSecret = {
        secretKey: secretKey,
        type: currentState.providerType,
      };
    }
    clearEmptyObject(config);
  });
}

function constructSecretUpdateInstruction(
  _config: PortalAPIAppConfig,
  _secrets: PortalAPISecretConfig,
  currentState: FormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  const secretKey =
    currentState.providerConfigs[currentState.providerType]?.secretKey;
  if (secretKey == null) {
    return undefined;
  }
  return {
    botProtectionProviderSecret: {
      action: "set",
      data: {
        secretKey: secretKey,
        type: currentState.providerType,
      },
    },
  };
}

interface ProviderCardProps {
  className?: string;
  logoSrc?: any;
  logoWidth?: number;
  logoHeight?: number;
  children?: React.ReactNode;
  onClick?: IButtonProps["onClick"];
  isSelected?: boolean;
  disabled?: boolean;
}

function ProviderCard(props: ProviderCardProps) {
  const {
    className,
    disabled,
    isSelected,
    children,
    onClick,
    logoSrc,
    logoHeight = 48,
    logoWidth = 48,
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
      {logoSrc != null ? (
        <Image src={logoSrc} width={logoWidth} height={logoHeight} />
      ) : null}
      <Label className={styles.providerCardLabel}>{children}</Label>
    </div>
  );
}

export interface BotProtectionConfigurationContentProviderConfigFormFieldsProps {
  revealed: boolean;
  onClickReveal: (e: React.MouseEvent<unknown>) => void;
  providerConfigs: Partial<
    Record<BotProtectionProviderType, FormBotProtectionProviderConfigs>
  >;
  setProviderConfigs: (
    fn: (
      c: Partial<
        Record<BotProtectionProviderType, FormBotProtectionProviderConfigs>
      >
    ) => Partial<
      Record<BotProtectionProviderType, FormBotProtectionProviderConfigs>
    >
  ) => void;
  providerType: BotProtectionProviderType;
}

const BotProtectionConfigurationContentProviderConfigFormFields: React.VFC<BotProtectionConfigurationContentProviderConfigFormFieldsProps> =
  function BotProtectionConfigurationContentProviderConfigFormFields(props) {
    const {
      revealed,
      onClickReveal,
      providerConfigs,
      setProviderConfigs,
      providerType,
    } = props;
    const { renderToString } = useContext(Context);

    const onChangeRecaptchaV2SiteKey = useCallback(
      (_, value?: string) => {
        if (value != null) {
          setProviderConfigs((c) => {
            return {
              ...c,
              recaptchav2: {
                secretKey: c["recaptchav2"]?.secretKey ?? null,
                siteKey: value,
              },
            };
          });
        }
      },
      [setProviderConfigs]
    );

    const onChangeRecaptchaV2SecretKey = useCallback(
      (_, value?: string) => {
        if (value != null) {
          setProviderConfigs((c) => {
            return {
              ...c,
              recaptchav2: {
                secretKey: value,
                siteKey: c["recaptchav2"]?.siteKey ?? "",
              },
            };
          });
        }
      },
      [setProviderConfigs]
    );

    const onChangeCloudflareSiteKey = useCallback(
      (_, value?: string) => {
        if (value != null) {
          setProviderConfigs((c) => {
            return {
              ...c,
              cloudflare: {
                secretKey: c["cloudflare"]?.secretKey ?? null,
                siteKey: value,
              },
            };
          });
        }
      },
      [setProviderConfigs]
    );

    const onChangeCloudflareSecretKey = useCallback(
      (_, value?: string) => {
        if (value != null) {
          setProviderConfigs((c) => {
            return {
              ...c,
              cloudflare: {
                secretKey: value,
                siteKey: c["cloudflare"]?.siteKey ?? "",
              },
            };
          });
        }
      },
      [setProviderConfigs]
    );

    const secretInputClassname = revealed
      ? styles.secretKeyInputWithoutReveal
      : styles.secretKeyInputWithReveal;

    const secretInputValue = revealed
      ? providerConfigs[providerType]?.secretKey ?? ""
      : MASKED_SECRET;

    return providerType === "recaptchav2" ? (
      <>
        <WidgetDescription>
          <FormattedMessage id="BotProtectionConfigurationScreen.provider.recaptchaV2.description" />
        </WidgetDescription>
        <FormTextField
          type="text"
          label={renderToString(
            "BotProtectionConfigurationScreen.provider.recaptchav2.siteKey.label"
          )}
          value={providerConfigs[providerType]?.siteKey ?? ""}
          required={true}
          onChange={onChangeRecaptchaV2SiteKey}
          parentJSONPointer=""
          fieldName="siteKey"
        />
        <div className={styles.secretKeyInputContainer}>
          <FormTextField
            className={secretInputClassname}
            id={SECRET_KEY_FORM_FIELD_ID}
            type="text"
            label={renderToString(
              "BotProtectionConfigurationScreen.provider.recaptchav2.secretKey.label"
            )}
            value={secretInputValue}
            required={true}
            onChange={onChangeRecaptchaV2SecretKey}
            parentJSONPointer=""
            fieldName="secretKey"
            readOnly={!revealed}
          />
          {!revealed ? (
            <PrimaryButton
              className={styles.secretKeyRevealButton}
              onClick={onClickReveal}
              text={<FormattedMessage id="reveal" />}
            />
          ) : null}
        </div>
      </>
    ) : (
      <>
        <WidgetDescription>
          <FormattedMessage id="BotProtectionConfigurationScreen.provider.cloudflare.description" />
        </WidgetDescription>
        <FormTextField
          type="text"
          label={renderToString(
            "BotProtectionConfigurationScreen.provider.cloudflare.siteKey.label"
          )}
          value={providerConfigs[providerType]?.siteKey ?? ""}
          required={true}
          onChange={onChangeCloudflareSiteKey}
          parentJSONPointer=""
          fieldName="siteKey"
        />
        <div className={styles.secretKeyInputContainer}>
          <FormTextField
            className={secretInputClassname}
            id={SECRET_KEY_FORM_FIELD_ID}
            type="text"
            label={renderToString(
              "BotProtectionConfigurationScreen.provider.cloudflare.secretKey.label"
            )}
            value={secretInputValue}
            required={true}
            onChange={onChangeCloudflareSecretKey}
            parentJSONPointer=""
            fieldName="secretKey"
            readOnly={!revealed}
          />
          {!revealed ? (
            <PrimaryButton
              className={styles.secretKeyRevealButton}
              onClick={onClickReveal}
              text={<FormattedMessage id="reveal" />}
            />
          ) : null}
        </div>
      </>
    );
  };

export interface BotProtectionConfigurationContentProps {
  form: AppSecretConfigFormModel<FormState>;
}

const BotProtectionConfigurationContent: React.VFC<BotProtectionConfigurationContentProps> =
  function BotProtectionConfigurationContent(props) {
    const { form } = props;
    const { state, setState } = form;
    const [storedFormState, setStoredFormState, removeStoredFormState] =
      useSessionStorage<FormState>(
        "bot-protection-config-screen-form-state",
        state
      );

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

    const onClickProviderRecaptchaV2 = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();
        if (state.providerType === "recaptchav2") {
          return;
        }
        setState((state) => {
          return {
            ...state,
            providerType: "recaptchav2",
          };
        });
      },
      [setState, state.providerType]
    );

    const onClickProviderCloudflare = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();
        if (state.providerType === "cloudflare") {
          return;
        }
        setState((state) => {
          return {
            ...state,
            providerType: "cloudflare",
          };
        });
      },
      [setState, state.providerType]
    );

    const locationState = useLocationEffect((state: LocationState) => {
      if (state.isOAuthRedirect) {
        window.location.hash = "";
        window.location.hash = "#" + SECRET_KEY_FORM_FIELD_ID;

        // Restore form state from local storage on reauth redirection
        setState((state) => {
          const providerConfigs = storedFormState.providerConfigs;
          const providerConfigsWithUnmodifiedSecretKey = Object.fromEntries(
            Object.entries(providerConfigs).map(
              ([providerType, providerConfig]) => {
                const _providerType = providerType as BotProtectionProviderType; // workaround ts unable to parse BotProtectionProviderType
                return [
                  _providerType,
                  {
                    ...providerConfig,
                    secretKey:
                      state.providerConfigs[_providerType]?.secretKey ?? null,
                  },
                ];
              }
            )
          );
          return {
            ...storedFormState,
            providerConfigs: providerConfigsWithUnmodifiedSecretKey,
          };
        });

        // Remove local storage form state after consuming
        removeStoredFormState();
      }
    });

    const [revealed, setRevealed] = useState(
      locationState?.isOAuthRedirect ?? false
    );

    const navigate = useNavigate();
    const onClickReveal = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();

        if (state.providerConfigs[state.providerType]?.secretKey != null) {
          setRevealed(true);
          return;
        }

        const locationState: LocationState = {
          isOAuthRedirect: true,
        };

        // Save form state to local storage, for later restoration on reauth redirect
        setStoredFormState({
          ...state,
          providerConfigs: Object.fromEntries(
            Object.entries(state.providerConfigs).map(
              ([providerType, providerConfig]) => {
                const _providerType = providerType as BotProtectionProviderType; // workaround ts unable to parse BotProtectionProviderType
                return [
                  _providerType,
                  {
                    ...providerConfig,
                    secretKey: null,
                  },
                ];
              }
            )
          ),
        }); // do not store secretKey

        startReauthentication(navigate, locationState).catch((e) => {
          // Normally there should not be any error.
          console.error(e);

          // Remove form state from local storage, since reauth failed, it will not be used
          removeStoredFormState();
        });
      },
      [navigate, removeStoredFormState, setStoredFormState, state]
    );

    const setBotProtectionProviderConfigs = useCallback(
      (
        fn: (
          c: Partial<
            Record<BotProtectionProviderType, FormBotProtectionProviderConfigs>
          >
        ) => Partial<
          Record<BotProtectionProviderType, FormBotProtectionProviderConfigs>
        >
      ) => {
        setState((state) => ({
          ...state,
          providerConfigs: fn(state.providerConfigs),
        }));
      },
      [setState]
    );

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="BotProtectionConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="BotProtectionConfigurationScreen.description" />
        </ScreenDescription>
        <div className={styles.content}>
          <Toggle
            // TODO: figure out 4px gap between toggle and label
            checked={state.enabled}
            onChange={onChangeEnabled}
            label={renderToString(
              "BotProtectionConfigurationScreen.enable.label"
            )}
            inlineLabel={false}
          />
          {state.enabled ? (
            <div className={styles.enabledContent}>
              <WidgetTitle>
                <FormattedMessage id="BotProtectionConfigurationScreen.challengeProvider.title" />
              </WidgetTitle>
              <div className={styles.providerCardContainer}>
                <ProviderCard
                  className={styles.columnLeft}
                  onClick={onClickProviderRecaptchaV2}
                  isSelected={state.providerType === "recaptchav2"}
                  logoSrc={recaptchaV2LogoURL}
                >
                  <FormattedMessage id="BotProtectionConfigurationScreen.provider.recaptchaV2.label" />
                </ProviderCard>
                <ProviderCard
                  className={styles.columnRight}
                  onClick={onClickProviderCloudflare}
                  isSelected={state.providerType === "cloudflare"}
                  logoSrc={cloudflareLogoURL}
                >
                  <FormattedMessage id="BotProtectionConfigurationScreen.provider.cloudflare.label" />
                </ProviderCard>
              </div>
              <BotProtectionConfigurationContentProviderConfigFormFields
                revealed={revealed}
                onClickReveal={onClickReveal}
                setProviderConfigs={setBotProtectionProviderConfigs}
                providerConfigs={state.providerConfigs}
                providerType={state.providerType}
              />
            </div>
          ) : null}
        </div>
      </ScreenContent>
    );
  };

const BotProtectionConfigurationScreen1: React.VFC<{
  appID: string;
  secretToken: string | null;
}> = function BotProtectionConfigurationScreen1({ appID, secretToken }) {
  const form = useAppSecretConfigForm({
    appID,
    secretVisitToken: secretToken,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <BotProtectionConfigurationContent form={form} />
    </FormContainer>
  );
};

const SECRETS = [AppSecretKey.BotProtectionProviderSecret];

const BotProtectionConfigurationScreen: React.VFC =
  function BotProtectionConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const location = useLocation();
    const [shouldRefreshToken] = useState<boolean>(() => {
      const { state } = location;
      if (isLocationState(state) && state.isOAuthRedirect) {
        return true;
      }
      return false;
    });
    const { token, error, retry } = useAppSecretVisitToken(
      appID,
      SECRETS,
      shouldRefreshToken
    );
    if (error) {
      return <ShowError error={error} onRetry={retry} />;
    }

    if (token === undefined) {
      return <ShowLoading />;
    }

    return (
      <BotProtectionConfigurationScreen1 appID={appID} secretToken={token} />
    );
  };

export default BotProtectionConfigurationScreen;
