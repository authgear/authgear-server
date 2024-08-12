import { Context, FormattedMessage } from "@oursky/react-messageformat";
import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
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
import FormContainer from "../../FormContainer";
import ScreenDescription from "../../ScreenDescription";
import Toggle from "../../Toggle";
import WidgetTitle from "../../WidgetTitle";
import {
  DetailsList,
  Dropdown,
  IButtonProps,
  IColumn,
  IDropdownOption,
  Image,
  Label,
  SelectionMode,
  Text,
} from "@fluentui/react";
import { useSystemConfig } from "../../context/SystemConfigContext";
import recaptchaV2LogoURL from "../../images/recaptchav2_logo.svg";
import cloudflareLogoURL from "../../images/cloudflare_logo.svg";
import WidgetDescription from "../../WidgetDescription";
import FormTextField from "../../FormTextField";
import PrimaryButton from "../../PrimaryButton";
import { startReauthentication } from "./Authenticated";
import { useSessionStorage } from "../../hook/useSessionStorage";
import HorizontalDivider from "../../HorizontalDivider";

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
  isSecretKeyEmpty: boolean;
}

interface FormRecaptchav2Configs {
  siteKey: string;
  secretKey: string | null;
  isSecretKeyEmpty: boolean;
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
  providerConfigs: Partial<
    Record<BotProtectionProviderType, FormBotProtectionProviderConfigs>
  >;
  requirements: FormBotProtectionRequirements;
}

function constructFormRequirementsState(
  config: PortalAPIAppConfig
): FormBotProtectionRequirements {
  const requirements = config.bot_protection?.requirements;
  // If any specific authenticator is configured, construct as specificAuthenticator, even if signup_or_login IS configured
  // otherwise, construct as allSignupLogin
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
  const siteKey = config.bot_protection?.provider?.site_key ?? "";
  const secretKey = secrets.botProtectionProviderSecret?.secretKey ?? null;
  const isSecretKeyEmpty = secrets.botProtectionProviderSecret == null; // secret key is empty if provider absent in authgear.secrets.yaml
  const providerConfigs: Partial<
    Record<BotProtectionProviderType, FormBotProtectionProviderConfigs>
  > = {
    [providerType]: {
      siteKey,
      secretKey,
      isSecretKeyEmpty,
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

function constructBotProtectionConfig(
  currentState: FormState
): BotProtectionConfig {
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
  return {
    enabled: currentState.enabled,
    provider: {
      type: currentState.providerType,
      site_key:
        currentState.providerConfigs[currentState.providerType]?.siteKey,
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
      },
    },
  } = useSystemConfig();

  return (
    <button
      type="button"
      style={{
        borderColor: isSelected ? themePrimary : "transparent",
      }}
      className={cn(className, styles.providerCard)}
      onClick={disabled ? undefined : onClick}
      tabIndex={0}
      disabled={disabled}
    >
      {logoSrc != null ? (
        <Image src={logoSrc} width={logoWidth} height={logoHeight} />
      ) : null}
      <Label className={styles.providerCardLabel}>{children}</Label>
    </button>
  );
}

export interface BotProtectionConfigurationContentProviderConfigFormFieldsProps {
  editing: boolean;
  onClickEdit: (e: React.MouseEvent<unknown>) => void;
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
      editing,
      onClickEdit,
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
                isSecretKeyEmpty: c["recaptchav2"]?.isSecretKeyEmpty ?? true,
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
                isSecretKeyEmpty: false,
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
                isSecretKeyEmpty: c["cloudflare"]?.isSecretKeyEmpty ?? true,
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
                isSecretKeyEmpty: false,
                siteKey: c["cloudflare"]?.siteKey ?? "",
              },
            };
          });
        }
      },
      [setProviderConfigs]
    );

    const secretInputClassname = editing
      ? styles.secretKeyInputWithoutEdit
      : styles.secretKeyInputWithEdit;

    const secretInputValue = editing
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
            readOnly={!editing}
          />
          {!editing ? (
            <PrimaryButton
              className={styles.secretKeyEditButton}
              onClick={onClickEdit}
              text={<FormattedMessage id="edit" />}
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
            readOnly={!editing}
          />
          {!editing ? (
            <PrimaryButton
              className={styles.secretKeyEditButton}
              onClick={onClickEdit}
              text={<FormattedMessage id="edit" />}
            />
          ) : null}
        </div>
      </>
    );
  };

interface BotProtectionConfigurationContentProviderSectionProps {
  form: AppSecretConfigFormModel<FormState>;
}
const BotProtectionConfigurationContentProviderSection: React.VFC<BotProtectionConfigurationContentProviderSectionProps> =
  function BotProtectionConfigurationContentProviderSection(props) {
    const { form } = props;
    const { state, setState } = form;
    const [storedFormState, setStoredFormState, removeStoredFormState] =
      useSessionStorage<FormState>(
        "bot-protection-config-screen-form-state",
        state
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

    const [editing, setediting] = useState(
      locationState?.isOAuthRedirect ??
        state.providerConfigs[state.providerType]?.isSecretKeyEmpty ??
        false
    );

    const navigate = useNavigate();
    const onClickEdit = useCallback(
      (e: React.MouseEvent<unknown>) => {
        e.preventDefault();
        e.stopPropagation();

        if (state.providerConfigs[state.providerType]?.secretKey != null) {
          setediting(true);
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
      <section className={styles.section}>
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
          editing={editing}
          onClickEdit={onClickEdit}
          setProviderConfigs={setBotProtectionProviderConfigs}
          providerConfigs={state.providerConfigs}
          providerType={state.providerType}
        />
      </section>
    );
  };

interface RequirementFlowHeaderListItem {
  label: string;
  mode: BotProtectionRiskMode;
  flowType: FormBotProtectionRequirementsFlowsType;
  onSelectDependsOnAuthenticator: () => void;
  onSelectMode: (mode: BotProtectionRiskMode) => void;
}

interface BotProtectionConfigurationContentRequirementsSectionFlowHeaderProps {
  requirements: FormBotProtectionRequirements;
  setRequirements: (
    fn: (r: FormBotProtectionRequirements) => FormBotProtectionRequirements
  ) => void;
}

const BotProtectionConfigurationContentRequirementsSectionFlowHeader: React.VFC<BotProtectionConfigurationContentRequirementsSectionFlowHeaderProps> =
  function BotProtectionConfigurationContentRequirementsSectionFlowHeader(
    props
  ) {
    const { requirements, setRequirements } = props;
    const { renderToString } = useContext(Context);

    const onRenderRequirementConfigLabel = useCallback(
      (
        item?: RequirementFlowHeaderListItem,
        _index?: number,
        _column?: IColumn
      ) => {
        if (item == null) {
          return null;
        }
        return (
          <div className={styles.requirementConfigLabelContainer}>
            <Text block={true} className={styles.requirementConfigLabel}>
              {item.label}
            </Text>
          </div>
        );
      },
      []
    );
    const DEPENDS_ON_AUTHENTICATOR_OPTION_KEY = "dependsOnSpecialAuthenticator";
    const onDropdownChange = useCallback(
      (
        _e: React.FormEvent<unknown>,
        option?: IDropdownOption<RequirementFlowHeaderListItem>,
        _index?: number
      ) => {
        if (option == null) {
          return;
        }
        switch (option.key) {
          case DEPENDS_ON_AUTHENTICATOR_OPTION_KEY: {
            option.data?.onSelectDependsOnAuthenticator();
            break;
          }
          default: {
            option.data?.onSelectMode(option.key as BotProtectionRiskMode);
          }
        }
      },
      []
    );
    const onRenderDropdown = useCallback(
      (
        item?: RequirementFlowHeaderListItem,
        index?: number,
        _column?: IColumn
      ) => {
        if (item == null || index == null) {
          return null;
        }

        const riskModeOptions: IDropdownOption<RequirementFlowHeaderListItem>[] =
          [
            {
              key: "never",
              text: renderToString(
                "BotProtectionConfigurationScreen.requirements.flows.config.riskMode.never"
              ),
              data: item,
            },
            {
              key: "always",
              text: renderToString(
                "BotProtectionConfigurationScreen.requirements.flows.config.riskMode.always"
              ),
              data: item,
            },
          ];

        const flowTypeOptions: IDropdownOption<RequirementFlowHeaderListItem>[] =
          [
            {
              key: DEPENDS_ON_AUTHENTICATOR_OPTION_KEY,
              text: renderToString(
                "BotProtectionConfigurationScreen.requirements.flows.type.dependsOnAuthenticator"
              ),
              data: item,
            },
          ];

        const options = [...riskModeOptions, ...flowTypeOptions];

        const selectedKey =
          item.flowType === "specificAuthenticator"
            ? DEPENDS_ON_AUTHENTICATOR_OPTION_KEY
            : item.mode;
        return (
          <Dropdown
            className={styles.requirementDropdownContainer}
            options={options}
            selectedKey={selectedKey}
            onChange={onDropdownChange}
          />
        );
      },
      [onDropdownChange, renderToString]
    );
    const requirementFlowHeaderColumns: IColumn[] = useMemo(() => {
      return [
        {
          key: "label",
          minWidth: 200,
          name: "",
          onRender: onRenderRequirementConfigLabel,
        },
        {
          key: "mode",
          minWidth: 300,
          maxWidth: 300,
          name: "",
          onRender: onRenderDropdown,
        },
      ];
    }, [onRenderDropdown, onRenderRequirementConfigLabel]);

    const flowHeaderListItems: RequirementFlowHeaderListItem[] = useMemo(() => {
      return [
        {
          label: renderToString(
            "BotProtectionConfigurationScreen.requirements.flows.config.allSignupLogin.label"
          ),
          mode: requirements.flows.flowConfigs.allSignupLogin
            .allSignupLoginMode,
          flowType: requirements.flows.flowType,
          onSelectDependsOnAuthenticator: () => {
            setRequirements((requirements) => ({
              ...requirements,
              flows: {
                ...requirements.flows,
                flowType: "specificAuthenticator",
              },
            }));
          },
          onSelectMode: (mode: BotProtectionRiskMode) => {
            setRequirements((requirements) => ({
              ...requirements,
              flows: {
                flowType: "allSignupLogin",
                flowConfigs: {
                  ...requirements.flows.flowConfigs,
                  allSignupLogin: {
                    allSignupLoginMode: mode,
                  },
                },
              },
            }));
          },
        },
      ];
    }, [
      renderToString,
      requirements.flows.flowConfigs.allSignupLogin.allSignupLoginMode,
      requirements.flows.flowType,
      setRequirements,
    ]);

    return (
      <DetailsList
        compact={true}
        columns={requirementFlowHeaderColumns}
        isHeaderVisible={false}
        selectionMode={SelectionMode.none}
        items={flowHeaderListItems}
      />
    );
  };

interface RequirementConfigListItem {
  label: string;
  mode: BotProtectionRiskMode;
  onChangeMode: (mode: BotProtectionRiskMode) => void;
}

interface BotProtectionConfigurationContentRequirementsSectionProps {
  requirements: FormBotProtectionRequirements;
  setRequirements: (
    fn: (r: FormBotProtectionRequirements) => FormBotProtectionRequirements
  ) => void;
}
const BotProtectionConfigurationContentRequirementsSection: React.VFC<BotProtectionConfigurationContentRequirementsSectionProps> =
  function BotProtectionConfigurationContentRequirementsSection(props) {
    const { requirements, setRequirements } = props;
    const { renderToString } = useContext(Context);

    const onRenderRequirementConfigLabel = useCallback(
      (
        item?: RequirementConfigListItem,
        _index?: number,
        _column?: IColumn
      ) => {
        if (item == null) {
          return null;
        }
        return (
          <div className={styles.requirementConfigLabelContainer}>
            <Text block={true} className={styles.requirementConfigLabel}>
              {item.label}
            </Text>
          </div>
        );
      },
      []
    );
    const makeDropdownOnChange = useCallback(() => {
      return (
        _e: React.FormEvent<unknown>,
        option?: IDropdownOption<RequirementConfigListItem>,
        _index?: number
      ) => {
        if (option == null) {
          return;
        }
        option.data?.onChangeMode(option.key as BotProtectionRiskMode);
      };
    }, []);
    const onRenderDropdown = useCallback(
      (item?: RequirementConfigListItem, index?: number, _column?: IColumn) => {
        if (item == null || index == null) {
          return null;
        }

        const options: IDropdownOption<RequirementConfigListItem>[] = [
          {
            key: "never",
            text: renderToString(
              "BotProtectionConfigurationScreen.requirements.flows.config.riskMode.never"
            ),
            data: item,
          },
          {
            key: "always",
            text: renderToString(
              "BotProtectionConfigurationScreen.requirements.flows.config.riskMode.always"
            ),
            data: item,
          },
        ];

        return (
          <Dropdown
            className={styles.requirementDropdownContainer}
            options={options}
            selectedKey={item.mode}
            onChange={makeDropdownOnChange()}
          />
        );
      },
      [makeDropdownOnChange, renderToString]
    );
    const requirementConfigColumns: IColumn[] = useMemo(() => {
      return [
        {
          key: "label",
          minWidth: 200,
          name: "",
          onRender: onRenderRequirementConfigLabel,
        },
        {
          key: "mode",
          minWidth: 300,
          maxWidth: 300,
          name: "",
          onRender: onRenderDropdown,
        },
      ];
    }, [onRenderDropdown, onRenderRequirementConfigLabel]);

    const setRequirementsFlowConfigs = useCallback(
      (
        fn: (
          r: FormBotProtectionRequirementsFlowConfigs
        ) => FormBotProtectionRequirementsFlowConfigs
      ) => {
        setRequirements((requirements) => ({
          ...requirements,
          flows: {
            ...requirements.flows,
            flowConfigs: fn(requirements.flows.flowConfigs),
          },
        }));
      },
      [setRequirements]
    );
    const flowConfigItems: RequirementConfigListItem[] = useMemo(() => {
      switch (requirements.flows.flowType) {
        case "specificAuthenticator": {
          return [
            {
              label: renderToString(
                "BotProtectionConfigurationScreen.requirements.flows.config.password.label"
              ),
              mode: requirements.flows.flowConfigs.specificAuthenticator
                .passwordMode,
              onChangeMode: (mode: BotProtectionRiskMode) => {
                setRequirementsFlowConfigs((flowConfigs) => ({
                  ...flowConfigs,
                  specificAuthenticator: {
                    ...flowConfigs.specificAuthenticator,
                    passwordMode: mode,
                  },
                }));
              },
            },
            {
              label: renderToString(
                "BotProtectionConfigurationScreen.requirements.flows.config.passwordlessSMS.label"
              ),
              mode: requirements.flows.flowConfigs.specificAuthenticator
                .passwordlessViaSMSMode,
              onChangeMode: (mode: BotProtectionRiskMode) => {
                setRequirementsFlowConfigs((flowConfigs) => ({
                  ...flowConfigs,
                  specificAuthenticator: {
                    ...flowConfigs.specificAuthenticator,
                    passwordlessViaSMSMode: mode,
                  },
                }));
              },
            },
            {
              label: renderToString(
                "BotProtectionConfigurationScreen.requirements.flows.config.passwordlessEmail.label"
              ),
              mode: requirements.flows.flowConfigs.specificAuthenticator
                .passwordlessViaEmailMode,
              onChangeMode: (mode: BotProtectionRiskMode) => {
                setRequirementsFlowConfigs((flowConfigs) => ({
                  ...flowConfigs,
                  specificAuthenticator: {
                    ...flowConfigs.specificAuthenticator,
                    passwordlessViaEmailMode: mode,
                  },
                }));
              },
            },
          ];
        }
        default:
          return [];
      }
    }, [renderToString, requirements.flows, setRequirementsFlowConfigs]);

    const resetPasswordConfigItems: RequirementConfigListItem[] =
      useMemo(() => {
        return [
          {
            label: renderToString(
              "BotProtectionConfigurationScreen.requirements.resetPassword.config.resetPassword.label"
            ),
            mode: requirements.resetPassword.resetPasswordMode,
            onChangeMode: (mode: BotProtectionRiskMode) => {
              setRequirements((requirements) => ({
                ...requirements,
                resetPassword: {
                  resetPasswordMode: mode,
                },
              }));
            },
          },
        ];
      }, [
        renderToString,
        requirements.resetPassword.resetPasswordMode,
        setRequirements,
      ]);

    return (
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <WidgetTitle>
            <FormattedMessage id="BotProtectionConfigurationScreen.requirements.title" />
          </WidgetTitle>
        </div>
        <div>
          <BotProtectionConfigurationContentRequirementsSectionFlowHeader
            requirements={requirements}
            setRequirements={setRequirements}
          />
          <DetailsList
            compact={true}
            columns={requirementConfigColumns}
            isHeaderVisible={false}
            selectionMode={SelectionMode.none}
            items={flowConfigItems}
          />
        </div>
        <HorizontalDivider />
        <div>
          <div>
            <DetailsList
              compact={true}
              columns={requirementConfigColumns}
              isHeaderVisible={false}
              selectionMode={SelectionMode.none}
              items={resetPasswordConfigItems}
            />
          </div>
        </div>
      </section>
    );
  };
export interface BotProtectionConfigurationContentProps {
  form: AppSecretConfigFormModel<FormState>;
}

const DEFAULT_BOT_PROTECTION_REQUIREMENTS_ON_ENABLE: FormBotProtectionRequirements =
  {
    flows: {
      flowType: "specificAuthenticator",
      flowConfigs: {
        allSignupLogin: {
          allSignupLoginMode: "never",
        },
        specificAuthenticator: {
          passwordMode: "never",
          passwordlessViaEmailMode: "never",
          passwordlessViaSMSMode: "always",
        },
      },
    },
    resetPassword: {
      resetPasswordMode: "always",
    },
  };

const BotProtectionConfigurationContent: React.VFC<BotProtectionConfigurationContentProps> =
  function BotProtectionConfigurationContent(props) {
    const { form } = props;
    const { state, setState } = form;
    const { renderToString } = useContext(Context);

    const onChangeEnabled = useCallback(
      (_event, checked?: boolean) => {
        if (checked != null) {
          setState((state) => {
            return {
              ...state,
              requirements: DEFAULT_BOT_PROTECTION_REQUIREMENTS_ON_ENABLE,
              enabled: checked,
            };
          });
        }
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
              <BotProtectionConfigurationContentProviderSection form={form} />
              <HorizontalDivider className="my-6" />
              <BotProtectionConfigurationContentRequirementsSection
                requirements={state.requirements}
                setRequirements={setRequirements}
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
