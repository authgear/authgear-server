import { Context, FormattedMessage } from "../../intl";
import React, {
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
} from "react";
import cn from "classnames";
import { RadioGroup, Select, Separator, Text, Flex } from "@radix-ui/themes";
import ScreenContent from "../../ScreenContent";
import styles from "./BotProtectionConfigurationScreen.module.css";
import {
  BotProtectionConfig,
  BotProtectionProviderType,
  BotProtectionRequirements,
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
import { useErrorMessage } from "../../formbinding";
import FormContainer from "../../FormContainer";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { TextField } from "../../components/v2/TextField/TextField";
import {
  IconRadioCards,
  IconRadioCardOption,
} from "../../components/v2/IconRadioCards/IconRadioCards";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { FormField } from "../../components/v2/FormField/FormField";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { useFormContainerBaseContext } from "../../FormContainerBase";
import { startReauthentication } from "./Authenticated";
import { useSessionStorage } from "../../hook/useSessionStorage";
import ExternalLink from "../../ExternalLink";
import recaptchaV2LogoURL from "../../images/recaptchav2_logo.svg";
import cloudflareLogoURL from "../../images/cloudflare_logo.svg";

const MASKED_SECRET = "***************";

const SECRET_KEY_FORM_FIELD_ID = "secret-key-form-field";

const PROVIDER_ICON_SIZE = 32;

const DEFAULT_BOT_PROTECTION_REQUIREMENTS_SPECIFIC_AUTHENTICATOR: FormBotProtectionRequirementsFlowsSpecificAuthenticatorFlowConfigs =
  {
    passwordMode: "never",
    passwordlessViaEmailMode: "never",
    passwordlessViaSMSMode: "always",
  };
const DEFAULT_BOT_PROTECTION_REQUIREMENTS_ON_ENABLE: FormBotProtectionRequirements =
  {
    flows: {
      flowType: "specificAuthenticator",
      flowConfigs: {
        allSignupLogin: {
          allSignupLoginMode: "never",
        },
        specificAuthenticator:
          DEFAULT_BOT_PROTECTION_REQUIREMENTS_SPECIFIC_AUTHENTICATOR,
      },
    },
    resetPassword: {
      resetPasswordMode: "always",
    },
  };

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
  originalSecretKey: string | null;
  editingSecretKey: string;
}

interface FormRecaptchav2Configs {
  siteKey: string;
  originalSecretKey: string | null;
  editingSecretKey: string;
}

type FormBotProtectionProviderConfigs =
  | FormCloudflareConfigs
  | FormRecaptchav2Configs;

type FormBotProtectionRequirementsFlowsType =
  | "allSignupLogin"
  | "specificAuthenticator";
interface FormBotProtectionRequirementsFlowsAllSignupLoginFlowConfigs {
  allSignupLoginMode: BotProtectionRiskMode;
}

interface FormBotProtectionRequirementsFlowsSpecificAuthenticatorFlowConfigs {
  passwordMode: BotProtectionRiskMode;
  passwordlessViaSMSMode: BotProtectionRiskMode;
  passwordlessViaEmailMode: BotProtectionRiskMode;
}

interface FormBotProtectionRequirementsFlowConfigs {
  allSignupLogin: FormBotProtectionRequirementsFlowsAllSignupLoginFlowConfigs;
  specificAuthenticator: FormBotProtectionRequirementsFlowsSpecificAuthenticatorFlowConfigs;
}
interface FormBotProtectionRequirementsFlows {
  flowType: FormBotProtectionRequirementsFlowsType;
  flowConfigs: FormBotProtectionRequirementsFlowConfigs;
}

interface FormBotProtectionRequirementsResetPassword {
  resetPasswordMode: BotProtectionRiskMode;
}

interface FormBotProtectionRequirements {
  flows: FormBotProtectionRequirementsFlows;
  resetPassword: FormBotProtectionRequirementsResetPassword;
}

interface FormState {
  enabled: boolean;
  providerType: BotProtectionProviderType;
  providerConfigs: Record<
    BotProtectionProviderType,
    FormBotProtectionProviderConfigs
  >;
  requirements: FormBotProtectionRequirements;
}

function constructFormRequirementsState(
  config: PortalAPIAppConfig
): FormBotProtectionRequirements {
  const requirements = config.bot_protection?.requirements;
  const isSpecificAuthenticatorConfigured =
    requirements?.oob_otp_email != null ||
    requirements?.oob_otp_sms != null ||
    requirements?.password != null;
  const dominantFlowType: FormBotProtectionRequirementsFlowsType =
    isSpecificAuthenticatorConfigured
      ? "specificAuthenticator"
      : "allSignupLogin";
  const flowConfigs = {
    allSignupLogin: {
      allSignupLoginMode: requirements?.signup_or_login?.mode ?? "never",
    },
    specificAuthenticator: {
      passwordMode: requirements?.password?.mode ?? "never",
      passwordlessViaSMSMode: requirements?.oob_otp_sms?.mode ?? "never",
      passwordlessViaEmailMode: requirements?.oob_otp_email?.mode ?? "never",
    },
  };

  const flows: FormBotProtectionRequirementsFlows = {
    flowType: dominantFlowType,
    flowConfigs,
  };
  const resetPassword: FormBotProtectionRequirementsResetPassword = {
    resetPasswordMode: requirements?.account_recovery?.mode ?? "never",
  };
  return {
    flows,
    resetPassword,
  };
}

function constructFormState(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig
): FormState {
  const enabled = config.bot_protection?.enabled ?? false;
  const providerType: BotProtectionProviderType =
    config.bot_protection?.provider?.type ?? "recaptchav2";

  const providerConfigs: Record<
    BotProtectionProviderType,
    FormBotProtectionProviderConfigs
  > = {
    cloudflare: {
      siteKey: "",
      originalSecretKey: "",
      editingSecretKey: "",
    },
    recaptchav2: {
      siteKey: "",
      originalSecretKey: "",
      editingSecretKey: "",
    },
  };

  if (enabled) {
    providerConfigs[providerType].siteKey =
      config.bot_protection?.provider?.site_key ?? "";
    providerConfigs[providerType].originalSecretKey =
      secrets.botProtectionProviderSecret?.secretKey ?? null;
    if (providerConfigs[providerType].originalSecretKey != null) {
      providerConfigs[providerType].editingSecretKey =
        providerConfigs[providerType].originalSecretKey;
    }
  }

  const requirements = constructFormRequirementsState(config);

  return {
    enabled,
    providerType,
    providerConfigs,
    requirements,
  };
}

function constructBotProtectionConfig(
  currentState: FormState
): BotProtectionConfig {
  const enabled = currentState.enabled;
  if (!enabled) {
    return {};
  }
  const signupOrLoginRequirements: Partial<BotProtectionRequirements> = {
    signup_or_login:
      currentState.requirements.flows.flowType === "allSignupLogin"
        ? {
            mode: currentState.requirements.flows.flowConfigs.allSignupLogin
              .allSignupLoginMode,
          }
        : undefined,
  };
  const accountRecoveryRequirements: Partial<BotProtectionRequirements> = {
    account_recovery: {
      mode: currentState.requirements.resetPassword.resetPasswordMode,
    },
  };
  const specificAuthenticatorRequirements: Partial<BotProtectionRequirements> =
    currentState.requirements.flows.flowType === "specificAuthenticator"
      ? {
          password: {
            mode: currentState.requirements.flows.flowConfigs
              .specificAuthenticator.passwordMode,
          },
          oob_otp_email: {
            mode: currentState.requirements.flows.flowConfigs
              .specificAuthenticator.passwordlessViaEmailMode,
          },
          oob_otp_sms: {
            mode: currentState.requirements.flows.flowConfigs
              .specificAuthenticator.passwordlessViaSMSMode,
          },
        }
      : {};
  const requirements: BotProtectionRequirements = {
    ...signupOrLoginRequirements,
    ...accountRecoveryRequirements,
    ...specificAuthenticatorRequirements,
  };

  let site_key: string | undefined =
    currentState.providerConfigs[currentState.providerType].siteKey;
  if (site_key === "") {
    site_key = undefined;
  }

  return {
    enabled,
    provider: {
      type: currentState.providerType,
      site_key,
    },
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
    config.bot_protection = constructBotProtectionConfig(currentState);

    const secretKey =
      currentState.providerConfigs[currentState.providerType].editingSecretKey;
    secrets.botProtectionProviderSecret = {
      secretKey: secretKey,
      type: currentState.providerType,
    };
    clearEmptyObject(config);
  });
}

function constructSecretUpdateInstruction(
  _config: PortalAPIAppConfig,
  _secrets: PortalAPISecretConfig,
  currentState: FormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  const enabled = currentState.enabled;
  const c = currentState.providerConfigs[currentState.providerType];

  const UNSET_INSTRUCTION: PortalAPISecretConfigUpdateInstruction = {
    botProtectionProviderSecret: {
      action: "unset",
    },
  };

  if (!enabled) {
    return UNSET_INSTRUCTION;
  }

  if (c.originalSecretKey == null) {
    return undefined;
  }

  const secretKey = c.editingSecretKey === "" ? null : c.editingSecretKey;
  return {
    botProtectionProviderSecret: {
      action: "set",
      data: {
        secretKey,
        type: currentState.providerType,
      },
    },
  };
}

// ---------------------------------------------------------------------------
// Requirements row components
// ---------------------------------------------------------------------------

interface RequirementSelectFieldProps {
  label: React.ReactNode;
  value: string;
  options: { value: string; label: string }[];
  onValueChange: (value: string) => void;
  disabled?: boolean;
}

function RequirementSelectField({
  label,
  value,
  options,
  onValueChange,
  disabled,
}: RequirementSelectFieldProps): React.ReactElement {
  return (
    <FormField size="2" labelSize="2" label={label} labelSpace="1">
      <Select.Root
        value={value}
        onValueChange={onValueChange}
        disabled={disabled}
      >
        <Select.Trigger
          variant="surface"
          className={styles.requirementSelectTrigger}
        />
        <Select.Content>
          {options.map((opt) => (
            <Select.Item key={opt.value} value={opt.value}>
              {opt.label}
            </Select.Item>
          ))}
        </Select.Content>
      </Select.Root>
    </FormField>
  );
}

interface RequirementRadioRowProps {
  label: React.ReactNode;
  value: BotProtectionRiskMode;
  onValueChange: (value: BotProtectionRiskMode) => void;
  neverLabel: string;
  alwaysLabel: string;
  showSeparator?: boolean;
}

function RequirementRadioRow({
  label,
  value,
  onValueChange,
  neverLabel,
  alwaysLabel,
  showSeparator = true,
}: RequirementRadioRowProps): React.ReactElement {
  return (
    <>
      <div className={styles.requirementRadioRow}>
        <Text
          as="p"
          size="2"
          weight="medium"
          className={styles.requirementRadioRowLabel}
        >
          {label}
        </Text>
        <RadioGroup.Root
          value={value}
          onValueChange={(v) => onValueChange(v as BotProtectionRiskMode)}
        >
          <Flex gap="6" align="center">
            <Text as="label" size="2" className={styles.requirementRadioOption}>
              <Flex gap="2" align="center">
                <RadioGroup.Item value="never" />
                {neverLabel}
              </Flex>
            </Text>
            <Text as="label" size="2" className={styles.requirementRadioOption}>
              <Flex gap="2" align="center">
                <RadioGroup.Item value="always" />
                {alwaysLabel}
              </Flex>
            </Text>
          </Flex>
        </RadioGroup.Root>
      </div>
      {showSeparator ? (
        <Separator size="4" className={styles.requirementRowSeparator} />
      ) : null}
    </>
  );
}

// ---------------------------------------------------------------------------
// Provider section
// ---------------------------------------------------------------------------

interface BotProtectionProviderSectionProps {
  form: AppSecretConfigFormModel<FormState>;
  enabled: boolean;
  onChangeEnabled: (checked: boolean) => void;
}

const BotProtectionProviderSection: React.VFC<BotProtectionProviderSectionProps> =
  function BotProtectionProviderSection({ form, enabled, onChangeEnabled }) {
    const { state, setState } = form;
    const { renderToString } = useContext(Context);

    const [storedFormState, setStoredFormState, removeStoredFormState] =
      useSessionStorage<FormState>(
        "bot-protection-config-screen-form-state",
        state
      );

    const locationState = useLocationEffect((state: LocationState) => {
      if (state.isOAuthRedirect) {
        window.location.hash = "";
        window.location.hash = "#" + SECRET_KEY_FORM_FIELD_ID;

        setState((state) => {
          return produce(storedFormState, (storedFormState) => {
            for (const [providerType, providerConfig] of Object.entries(
              storedFormState.providerConfigs
            )) {
              if (storedFormState.providerType === providerType) {
                const newlyFetchedProviderConfig =
                  state.providerConfigs[providerType];
                storedFormState.providerConfigs[providerType] = {
                  ...providerConfig,
                  originalSecretKey:
                    newlyFetchedProviderConfig.originalSecretKey,
                  editingSecretKey: newlyFetchedProviderConfig.editingSecretKey,
                };
              }
            }
          });
        });

        removeStoredFormState();
      }
    });

    const [reauthed, setReauthed] = useState(locationState?.isOAuthRedirect);

    const editing = useMemo(() => {
      const currentProviderConfig = state.providerConfigs[state.providerType];
      const shouldMaskSecretKeyIfNotReauthed =
        currentProviderConfig.originalSecretKey == null;
      return reauthed ?? !shouldMaskSecretKeyIfNotReauthed;
    }, [reauthed, state.providerConfigs, state.providerType]);

    const navigate = useNavigate();
    const onClickEdit = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();

        if (
          state.providerConfigs[state.providerType].originalSecretKey != null
        ) {
          setReauthed(true);
          return;
        }

        const locationState: LocationState = {
          isOAuthRedirect: true,
        };

        setStoredFormState({ ...state });

        startReauthentication(navigate, locationState).catch((e) => {
          console.error(e);
          removeStoredFormState();
        });
      },
      [navigate, removeStoredFormState, setStoredFormState, state]
    );

    const onChangeProviderType = useCallback(
      (value: BotProtectionProviderType) => {
        setState((state) => ({ ...state, providerType: value }));
      },
      [setState]
    );

    const onChangeRecaptchaV2SiteKey = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((state) => ({
          ...state,
          providerConfigs: {
            ...state.providerConfigs,
            recaptchav2: {
              ...state.providerConfigs.recaptchav2,
              siteKey: value,
            },
          },
        }));
      },
      [setState]
    );

    const onChangeRecaptchaV2SecretKey = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((state) => ({
          ...state,
          providerConfigs: {
            ...state.providerConfigs,
            recaptchav2: {
              ...state.providerConfigs.recaptchav2,
              editingSecretKey: value,
            },
          },
        }));
      },
      [setState]
    );

    const onChangeCloudflareSiteKey = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((state) => ({
          ...state,
          providerConfigs: {
            ...state.providerConfigs,
            cloudflare: {
              ...state.providerConfigs.cloudflare,
              siteKey: value,
            },
          },
        }));
      },
      [setState]
    );

    const onChangeCloudflareSecretKey = useCallback(
      (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value;
        setState((state) => ({
          ...state,
          providerConfigs: {
            ...state.providerConfigs,
            cloudflare: {
              ...state.providerConfigs.cloudflare,
              editingSecretKey: value,
            },
          },
        }));
      },
      [setState]
    );

    const providerOptions = useMemo(
      (): IconRadioCardOption<BotProtectionProviderType>[] => [
        {
          value: "recaptchav2",
          icon: (
            <img
              src={recaptchaV2LogoURL}
              alt=""
              className="object-contain"
              style={{ width: PROVIDER_ICON_SIZE, height: PROVIDER_ICON_SIZE }}
            />
          ),
          title: (
            <FormattedMessage id="BotProtectionConfigurationScreen.provider.recaptchaV2.label" />
          ),
        },
        {
          value: "cloudflare",
          icon: (
            <img
              src={cloudflareLogoURL}
              alt=""
              className="object-contain"
              style={{ width: PROVIDER_ICON_SIZE, height: PROVIDER_ICON_SIZE }}
            />
          ),
          title: (
            <FormattedMessage id="BotProtectionConfigurationScreen.provider.cloudflare.label" />
          ),
        },
      ],
      []
    );

    const providerDescription = useMemo(() => {
      if (state.providerType === "recaptchav2") {
        return (
          <FormattedMessage
            id="BotProtectionConfigurationScreen.provider.recaptchaV2.description"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              ExternalLink: (chunks: React.ReactNode) => (
                <ExternalLink href="https://developers.google.com/recaptcha/docs/settings">
                  {chunks}
                </ExternalLink>
              ),
            }}
          />
        );
      }
      return (
        <FormattedMessage
          id="BotProtectionConfigurationScreen.provider.cloudflare.description"
          values={{
            // eslint-disable-next-line react/no-unstable-nested-components
            ExternalLink: (chunks: React.ReactNode) => (
              <ExternalLink href="https://developers.cloudflare.com/turnstile/get-started/#get-a-sitekey-and-secret-key">
                {chunks}
              </ExternalLink>
            ),
          }}
        />
      );
    }, [state.providerType]);

    const siteKeyLabel =
      state.providerType === "recaptchav2"
        ? renderToString(
            "BotProtectionConfigurationScreen.provider.recaptchav2.siteKey.label"
          )
        : renderToString(
            "BotProtectionConfigurationScreen.provider.cloudflare.siteKey.label"
          );

    const secretKeyLabel =
      state.providerType === "recaptchav2"
        ? renderToString(
            "BotProtectionConfigurationScreen.provider.recaptchav2.secretKey.label"
          )
        : renderToString(
            "BotProtectionConfigurationScreen.provider.cloudflare.secretKey.label"
          );

    const onChangeSiteKey =
      state.providerType === "recaptchav2"
        ? onChangeRecaptchaV2SiteKey
        : onChangeCloudflareSiteKey;

    const onChangeSecretKey =
      state.providerType === "recaptchav2"
        ? onChangeRecaptchaV2SecretKey
        : onChangeCloudflareSecretKey;

    const secretInputValue = editing
      ? state.providerConfigs[state.providerType].editingSecretKey
      : MASKED_SECRET;

    const secretKeyField = useMemo(
      () => ({
        parentJSONPointer: /secrets\/\d+\/data/,
        fieldName: "secret_key",
      }),
      []
    );
    const secretKeyFieldProps = useErrorMessage(secretKeyField);

    return (
      <SettingsSectionCard
        className={styles.widget}
        title={
          <FormattedMessage id="BotProtectionConfigurationScreen.settings.label" />
        }
        contentClassName="gap-4"
      >
        <div className={styles.enableToggle}>
          <Toggle
            checked={enabled}
            onCheckedChange={onChangeEnabled}
            textWeight="medium"
            text={
              <FormattedMessage id="BotProtectionConfigurationScreen.enable.label" />
            }
          />
        </div>
        {enabled ? (
          <>
            <div className={styles.providerSelector}>
              <IconRadioCards
                size="3"
                value={state.providerType}
                onValueChange={onChangeProviderType}
                options={providerOptions}
                itemFillSpaces={true}
              />
              <Text
                as="p"
                size="1"
                color="gray"
                className={styles.providerDescription}
              >
                {providerDescription}
              </Text>
            </div>
            <TextField
              size="2"
              labelSize="2"
              type="text"
              label={siteKeyLabel}
              value={state.providerConfigs[state.providerType].siteKey}
              required={true}
              onChange={onChangeSiteKey}
              parentJSONPointer="/bot_protection/provider"
              fieldName="site_key"
            />
            <FormField
              size="2"
              labelSize="2"
              label={secretKeyLabel}
              htmlFor={SECRET_KEY_FORM_FIELD_ID}
              required={true}
              labelSpace="1"
              parentJSONPointer={/secrets\/\d+\/data/}
              fieldName="secret_key"
            >
              <div className={styles.secretKeyInputRow}>
                <TextField.Input
                  id={SECRET_KEY_FORM_FIELD_ID}
                  size="2"
                  inputClassName={styles.secretKeyInput}
                  type="text"
                  value={secretInputValue}
                  required={true}
                  onChange={onChangeSecretKey}
                  readOnly={!editing}
                  disabled={secretKeyFieldProps.disabled}
                  error={secretKeyFieldProps.errorMessage}
                >
                  {null}
                </TextField.Input>
                {!editing ? (
                  <SecondaryButton
                    size="2"
                    onClick={onClickEdit}
                    text={<FormattedMessage id="edit" />}
                  />
                ) : null}
              </div>
            </FormField>
          </>
        ) : null}
      </SettingsSectionCard>
    );
  };

// ---------------------------------------------------------------------------
// Requirements section
// ---------------------------------------------------------------------------

interface BotProtectionRequirementsSectionProps {
  requirements: FormBotProtectionRequirements;
  setRequirements: (
    fn: (r: FormBotProtectionRequirements) => FormBotProtectionRequirements
  ) => void;
  isDirty: boolean;
}

const BotProtectionRequirementsSection: React.VFC<BotProtectionRequirementsSectionProps> =
  function BotProtectionRequirementsSection({
    requirements,
    setRequirements,
    isDirty,
  }) {
    const { renderToString } = useContext(Context);

    const neverLabel = renderToString(
      "BotProtectionConfigurationScreen.requirements.flows.config.riskMode.never"
    );
    const alwaysLabel = renderToString(
      "BotProtectionConfigurationScreen.requirements.flows.config.riskMode.always"
    );

    const riskModeOptions = useMemo(
      () => [
        {
          value: "never",
          label: neverLabel,
        },
        {
          value: "always",
          label: alwaysLabel,
        },
      ],
      [neverLabel, alwaysLabel]
    );

    const allSignupLoginOptions = useMemo(
      () => [
        ...riskModeOptions,
        {
          value: "dependsOnAuthenticator",
          label: renderToString(
            "BotProtectionConfigurationScreen.requirements.flows.type.dependsOnAuthenticator"
          ),
        },
      ],
      [riskModeOptions, renderToString]
    );

    const isSpecificAuthenticator =
      requirements.flows.flowType === "specificAuthenticator";

    const allSignupLoginValue = isSpecificAuthenticator
      ? "dependsOnAuthenticator"
      : requirements.flows.flowConfigs.allSignupLogin.allSignupLoginMode;

    const onChangeAllSignupLogin = useCallback(
      (value: string) => {
        if (value === "dependsOnAuthenticator") {
          setRequirements((r) => ({
            ...r,
            flows: {
              ...r.flows,
              flowConfigs: {
                ...r.flows.flowConfigs,
                specificAuthenticator:
                  DEFAULT_BOT_PROTECTION_REQUIREMENTS_SPECIFIC_AUTHENTICATOR,
              },
              flowType: "specificAuthenticator",
            },
          }));
        } else {
          const mode = value as BotProtectionRiskMode;
          setRequirements((r) => ({
            ...r,
            flows: {
              flowType: "allSignupLogin",
              flowConfigs: {
                specificAuthenticator: {
                  passwordMode: mode,
                  passwordlessViaSMSMode: mode,
                  passwordlessViaEmailMode: mode,
                },
                allSignupLogin: { allSignupLoginMode: mode },
              },
            },
          }));
        }
      },
      [setRequirements]
    );

    const onChangePasswordMode = useCallback(
      (value: string) => {
        setRequirements((r) => ({
          ...r,
          flows: {
            ...r.flows,
            flowConfigs: {
              ...r.flows.flowConfigs,
              specificAuthenticator: {
                ...r.flows.flowConfigs.specificAuthenticator,
                passwordMode: value as BotProtectionRiskMode,
              },
            },
          },
        }));
      },
      [setRequirements]
    );

    const onChangePasswordlessViaSMSMode = useCallback(
      (value: string) => {
        setRequirements((r) => ({
          ...r,
          flows: {
            ...r.flows,
            flowConfigs: {
              ...r.flows.flowConfigs,
              specificAuthenticator: {
                ...r.flows.flowConfigs.specificAuthenticator,
                passwordlessViaSMSMode: value as BotProtectionRiskMode,
              },
            },
          },
        }));
      },
      [setRequirements]
    );

    const onChangePasswordlessViaEmailMode = useCallback(
      (value: string) => {
        setRequirements((r) => ({
          ...r,
          flows: {
            ...r.flows,
            flowConfigs: {
              ...r.flows.flowConfigs,
              specificAuthenticator: {
                ...r.flows.flowConfigs.specificAuthenticator,
                passwordlessViaEmailMode: value as BotProtectionRiskMode,
              },
            },
          },
        }));
      },
      [setRequirements]
    );

    const onChangeResetPasswordMode = useCallback(
      (value: string) => {
        setRequirements((r) => ({
          ...r,
          resetPassword: { resetPasswordMode: value as BotProtectionRiskMode },
        }));
      },
      [setRequirements]
    );

    return (
      <SettingsSectionCard
        className={cn(
          styles.widget,
          isDirty && styles.settingsCardSaveBarClearance
        )}
        title={
          <FormattedMessage id="BotProtectionConfigurationScreen.requirements.title" />
        }
        contentClassName="gap-4"
      >
        <RequirementSelectField
          label={
            <FormattedMessage id="BotProtectionConfigurationScreen.requirements.flows.config.allSignupLogin.label" />
          }
          value={allSignupLoginValue}
          options={allSignupLoginOptions}
          onValueChange={onChangeAllSignupLogin}
        />

        {isSpecificAuthenticator ? (
          <div className={styles.requirementsNestedCard}>
            <RequirementRadioRow
              label={
                <FormattedMessage id="BotProtectionConfigurationScreen.requirements.flows.config.password.label" />
              }
              value={
                requirements.flows.flowConfigs.specificAuthenticator
                  .passwordMode
              }
              onValueChange={onChangePasswordMode}
              neverLabel={neverLabel}
              alwaysLabel={alwaysLabel}
            />
            <RequirementRadioRow
              label={
                <FormattedMessage id="BotProtectionConfigurationScreen.requirements.flows.config.passwordlessSMS.label" />
              }
              value={
                requirements.flows.flowConfigs.specificAuthenticator
                  .passwordlessViaSMSMode
              }
              onValueChange={onChangePasswordlessViaSMSMode}
              neverLabel={neverLabel}
              alwaysLabel={alwaysLabel}
            />
            <RequirementRadioRow
              label={
                <FormattedMessage id="BotProtectionConfigurationScreen.requirements.flows.config.passwordlessEmail.label" />
              }
              value={
                requirements.flows.flowConfigs.specificAuthenticator
                  .passwordlessViaEmailMode
              }
              onValueChange={onChangePasswordlessViaEmailMode}
              neverLabel={neverLabel}
              alwaysLabel={alwaysLabel}
              showSeparator={false}
            />
          </div>
        ) : null}

        <Separator size="4" />

        <RequirementSelectField
          label={
            <FormattedMessage id="BotProtectionConfigurationScreen.requirements.resetPassword.config.resetPassword.label" />
          }
          value={requirements.resetPassword.resetPasswordMode}
          options={riskModeOptions}
          onValueChange={onChangeResetPasswordMode}
        />
      </SettingsSectionCard>
    );
  };

// ---------------------------------------------------------------------------
// Main content
// ---------------------------------------------------------------------------

export interface BotProtectionConfigurationContentProps {
  form: AppSecretConfigFormModel<FormState>;
}

const BotProtectionConfigurationContent: React.VFC<BotProtectionConfigurationContentProps> =
  function BotProtectionConfigurationContent({ form }) {
    const { state, setState } = form;
    const { getIsDirty } = useFormContainerBaseContext();
    const isDirty = useMemo(() => getIsDirty(), [getIsDirty]);
    const contentWidthAnchorRef = useRef<HTMLDivElement>(null);

    const onChangeEnabled = useCallback(
      (checked: boolean) => {
        setState((state) => ({
          ...state,
          requirements: DEFAULT_BOT_PROTECTION_REQUIREMENTS_ON_ENABLE,
          enabled: checked,
        }));
      },
      [setState]
    );

    const setRequirements = useCallback(
      (
        fn: (r: FormBotProtectionRequirements) => FormBotProtectionRequirements
      ) => {
        setState((state) => ({
          ...state,
          requirements: fn(state.requirements),
        }));
      },
      [setState]
    );

    return (
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        {/* Page header */}
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Text as="p" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="BotProtectionConfigurationScreen.title" />
          </Text>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="BotProtectionConfigurationScreen.description" />
          </Text>
        </div>

        {/* Settings card: enable toggle + provider config when enabled */}
        <BotProtectionProviderSection
          form={form}
          enabled={state.enabled}
          onChangeEnabled={onChangeEnabled}
        />

        {/* Requirements — only when enabled */}
        {state.enabled ? (
          <BotProtectionRequirementsSection
            requirements={state.requirements}
            setRequirements={setRequirements}
            isDirty={isDirty}
          />
        ) : null}

        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    );
  };

// ---------------------------------------------------------------------------
// Screen wrappers (unchanged)
// ---------------------------------------------------------------------------

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
    <FormContainer form={form} canSave={true}>
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
