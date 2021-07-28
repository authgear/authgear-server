import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { Navigate, useNavigate, useParams } from "react-router-dom";
import produce from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import cn from "classnames";
import {
  Checkbox,
  ChoiceGroup,
  Dropdown,
  FontIcon,
  IChoiceGroupOption,
  IDropdownOption,
  Label,
  Text,
  Link,
  DirectionalHint,
  TooltipHost,
} from "@fluentui/react";
import {
  AuthenticatorsFeatureConfig,
  IdentityFeatureConfig,
  IdentityType,
  LoginIDKeyType,
  PortalAPIAppConfig,
  PortalAPIFeatureConfig,
  SecondaryAuthenticationMode,
  SecondaryAuthenticatorType,
  VerificationClaimsConfig,
} from "../../types";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenHeader from "../../ScreenHeader";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import OnboardingFormContainer from "./OnboardingFormContainer";
import styles from "./OnboardingConfigAppScreen.module.scss";
import LabelWithTooltip from "../../LabelWithTooltip";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";

const primaryAuthenticatorTypes = ["password", "oob"] as const;
type PrimaryAuthenticatorType = typeof primaryAuthenticatorTypes[number];

const identitiesButtonItems: IdentitiesButton[] = [
  {
    labelId: "Onboarding.identities.email",
    iconName: "Mail",
    loginIDType: "email",
  },
  {
    labelId: "Onboarding.identities.phone",
    iconName: "CellPhone",
    loginIDType: "phone",
  },
  {
    labelId: "Onboarding.identities.username",
    iconName: "ContactCard",
    loginIDType: "username",
  },
  {
    labelId: "Onboarding.identities.sso",
    iconName: "Globe",
    identityType: "oauth",
  },
  {
    labelId: "Onboarding.identities.anonymous-user",
    iconName: "FollowUser",
    identityType: "anonymous",
  },
];

const secondaryAuthenticatorOptions: SecondaryAuthenticatorOption[] = [
  {
    labelId: "Onboarding.secondary-authenticators.totp",
    authenticatorType: "totp",
  },
  {
    labelId: "Onboarding.secondary-authenticators.oob-sms",
    authenticatorType: "oob_otp_sms",
  },
  {
    labelId: "Onboarding.secondary-authenticators.oob-email",
    authenticatorType: "oob_otp_email",
  },
  {
    labelId: "Onboarding.secondary-authenticators.additional-password",
    authenticatorType: "password",
  },
];

// sort login id based on button order
function sortLoginIDKeyTypes(loginIDTypes: LoginIDKeyType[]): LoginIDKeyType[] {
  const indexMap = new Map<LoginIDKeyType, number>();
  identitiesButtonItems.forEach((btn: IdentitiesButton, idx) => {
    if (btn.loginIDType) {
      indexMap.set(btn.loginIDType, idx);
    }
  });

  return loginIDTypes.sort((a, b) => {
    if (indexMap.get(a) === undefined) return 1;
    if (indexMap.get(b) === undefined) return -1;
    return indexMap.get(a)! - indexMap.get(b)!;
  });
}

// sort secondary authenticators based on checkbox order
function sortSecondaryAuthenticatorTypes(
  authenticatorTypes: SecondaryAuthenticatorType[]
): SecondaryAuthenticatorType[] {
  const indexMap = new Map<SecondaryAuthenticatorType, number>();
  secondaryAuthenticatorOptions.forEach(
    (option: SecondaryAuthenticatorOption, idx) => {
      indexMap.set(option.authenticatorType, idx);
    }
  );

  return authenticatorTypes.sort((a, b) => {
    if (indexMap.get(a) === undefined) return 1;
    if (indexMap.get(b) === undefined) return -1;
    return indexMap.get(a)! - indexMap.get(b)!;
  });
}

interface PendingFormState {
  identities: Set<IdentityType>;
  loginIDKeys: Set<LoginIDKeyType>;
  primaryAuthenticator: PrimaryAuthenticatorType;
  secondaryAuthenticationMode: SecondaryAuthenticationMode;
  secondaryAuthenticators: Set<SecondaryAuthenticatorType>;
  verificationClaims: VerificationClaimsConfig;
}

const defaultPendingFormState: PendingFormState = {
  identities: new Set<IdentityType>(["login_id"]),
  loginIDKeys: new Set<LoginIDKeyType>(["email"]),
  primaryAuthenticator: "password",
  secondaryAuthenticationMode: "if_exists",
  secondaryAuthenticators: new Set<SecondaryAuthenticatorType>(["totp"]),
  verificationClaims: {
    email: { enabled: true, required: true },
    phone_number: { enabled: true, required: true },
  },
};

interface FormState {
  pendingForm: PendingFormState;
}

function constructFormState(_config: PortalAPIAppConfig): FormState {
  return {
    pendingForm: defaultPendingFormState,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    config.authentication ??= {};
    config.authentication.identities = Array.from(
      currentState.pendingForm.identities
    );

    config.identity ??= {};
    config.identity.login_id ??= {};
    config.identity.login_id.keys = sortLoginIDKeyTypes(
      Array.from(currentState.pendingForm.loginIDKeys)
    ).map((t) => {
      return { type: t, key: t };
    });

    if (!currentState.pendingForm.identities.has("login_id")) {
      return;
    }

    config.authentication.primary_authenticators = [];
    if (currentState.pendingForm.primaryAuthenticator === "password") {
      config.authentication.primary_authenticators.push("password");
    }
    if (currentState.pendingForm.primaryAuthenticator === "oob") {
      if (currentState.pendingForm.loginIDKeys.has("email")) {
        config.authentication.primary_authenticators.push("oob_otp_email");
      }
      if (currentState.pendingForm.loginIDKeys.has("phone")) {
        config.authentication.primary_authenticators.push("oob_otp_sms");
      }
    }

    config.authentication.secondary_authentication_mode =
      currentState.pendingForm.secondaryAuthenticationMode;

    config.authentication.secondary_authenticators =
      sortSecondaryAuthenticatorTypes(
        Array.from(currentState.pendingForm.secondaryAuthenticators)
      );

    if (currentState.pendingForm.loginIDKeys.has("email")) {
      config.verification ??= {};
      config.verification.claims ??= {};
      config.verification.claims.email =
        currentState.pendingForm.verificationClaims.email;
    }

    if (currentState.pendingForm.loginIDKeys.has("phone")) {
      config.verification ??= {};
      config.verification.claims ??= {};
      config.verification.claims.phone_number =
        currentState.pendingForm.verificationClaims.phone_number;
    }
  });
}

interface UpgradeButtonProps {
  labelMessageID: string;
  headerMessageID: string;
  bodyMessageID: string;
}

const UpgradeButton: React.FC<UpgradeButtonProps> = function UpgradeButton(
  props: UpgradeButtonProps
) {
  const { labelMessageID, headerMessageID, bodyMessageID } = props;

  const tooltipContent = useMemo(
    () => (
      <div className={styles.upgradeTooltip}>
        <Text
          variant="xLarge"
          block={true}
          className={styles.upgradeTooltipHeader}
        >
          <FormattedMessage id={headerMessageID} />
        </Text>
        <Text variant="medium" block={true}>
          <FormattedMessage id={bodyMessageID} />
        </Text>
      </div>
    ),
    [headerMessageID, bodyMessageID]
  );

  return (
    <div>
      <TooltipHost
        tooltipProps={{
          onRenderContent: () => tooltipContent,
        }}
        directionalHint={DirectionalHint.bottomLeftEdge}
      >
        <Link className={styles.upgradeLink}>
          <FormattedMessage id={labelMessageID} />
        </Link>
      </TooltipHost>
    </div>
  );
};

interface IdentitiesButton {
  labelId: string;
  iconName: string;
  // button should have either identityType or loginIDType
  // has loginIDType implicitly means identityType == login_id
  identityType?: IdentityType;
  loginIDType?: LoginIDKeyType;
}

interface IdentitiesItemContentProps {
  form: AppConfigFormModel<FormState>;
  identityFeatureConfig?: IdentityFeatureConfig;
  btnItem: IdentitiesButton;
}

const IdentitiesItemContent: React.FC<IdentitiesItemContentProps> =
  function IdentitiesItemContent(props) {
    const {
      form: { state, setState },
      identityFeatureConfig,
      btnItem,
    } = props;

    const identityChecked = useMemo(() => {
      // check only if button item has loginIDType (e.g. email, phone and username)
      if (btnItem.loginIDType) {
        return state.pendingForm.loginIDKeys.has(btnItem.loginIDType);
      } else if (btnItem.identityType) {
        return state.pendingForm.identities.has(btnItem.identityType);
      }
      console.error(
        "IdentitiesButton should have either identityType or loginIDType"
      );
      return false;
    }, [state, btnItem]);

    const identityDisabled = useMemo(() => {
      if (btnItem.loginIDType === "phone") {
        return identityFeatureConfig?.login_id?.types?.phone?.disabled ?? false;
      }

      return false;
    }, [btnItem.loginIDType, identityFeatureConfig]);

    const upgradeButton = useMemo(() => {
      if (btnItem.loginIDType === "phone" && identityDisabled) {
        return (
          <UpgradeButton
            labelMessageID="Onboarding.upgrade.label"
            headerMessageID="Onboarding.upgrade.support-sms.title"
            bodyMessageID="Onboarding.upgrade.support-sms.desc"
          />
        );
      }
      return undefined;
    }, [btnItem.loginIDType, identityDisabled]);

    const onCheckedChange = useCallback(
      (checked?: boolean) => {
        const identities = new Set(state.pendingForm.identities);
        const loginIDKeys = new Set(state.pendingForm.loginIDKeys);
        if (btnItem.loginIDType) {
          if (checked) {
            loginIDKeys.add(btnItem.loginIDType);
          } else {
            loginIDKeys.delete(btnItem.loginIDType);
          }
        } else if (btnItem.identityType) {
          if (checked) {
            identities.add(btnItem.identityType);
          } else {
            identities.delete(btnItem.identityType);
          }
        } else {
          console.error(
            "IdentitiesButton should have either identityType or loginIDType"
          );
          return;
        }

        // check if there is any login id enabled
        // and update login_id in identities list
        if (loginIDKeys.size > 0) {
          identities.add("login_id");
        } else {
          identities.delete("login_id");
        }

        // if username is selected, the primary authenticator must be password
        let primaryAuthenticator = state.pendingForm.primaryAuthenticator;
        if (loginIDKeys.has("username")) {
          primaryAuthenticator = "password";
        }

        setState((prev) => ({
          ...prev,
          pendingForm: {
            ...prev.pendingForm,
            identities,
            loginIDKeys,
            primaryAuthenticator,
          },
        }));
      },
      [state, setState, btnItem]
    );

    const onItemClick = useCallback(
      (event: React.FormEvent) => {
        event.preventDefault();
        event.stopPropagation();
        const currentChecked = identityChecked;
        onCheckedChange(!currentChecked);
      },
      [identityChecked, onCheckedChange]
    );

    const onCheckboxChange = useCallback(
      (event?: React.FormEvent, checked?: boolean) => {
        event?.preventDefault();
        event?.stopPropagation();
        onCheckedChange(checked);
      },
      [onCheckedChange]
    );

    return (
      <div className={styles.identityListItem}>
        <div
          className={cn(styles.identityListItemContent, {
            [styles.readOnly]: identityDisabled,
          })}
          onClick={onItemClick}
        >
          <div className={styles.label}>
            <FontIcon iconName={btnItem.iconName} className={styles.icon} />
            <Text block={true} variant="medium">
              <FormattedMessage id={btnItem.labelId} />
            </Text>
          </div>
          <Checkbox
            className={styles.checkbox}
            checked={identityChecked}
            onChange={onCheckboxChange}
          />
        </div>
        {upgradeButton}
      </div>
    );
  };

interface IdentitiesListContentProps {
  form: AppConfigFormModel<FormState>;
  identityFeatureConfig?: IdentityFeatureConfig;
}

const IdentitiesListContent: React.FC<IdentitiesListContentProps> =
  function IdentitiesListContent(props) {
    const { form, identityFeatureConfig } = props;

    const showUsernameOnlyAlert = useMemo(
      () =>
        form.state.pendingForm.loginIDKeys.size === 1 &&
        form.state.pendingForm.loginIDKeys.has("username"),
      [form.state.pendingForm.loginIDKeys]
    );

    const showAnonymousOnlyAlert = useMemo(
      () =>
        form.state.pendingForm.identities.size === 1 &&
        form.state.pendingForm.identities.has("anonymous"),
      [form.state.pendingForm.identities]
    );

    return (
      <section className={styles.sections}>
        <Label className={styles.fieldLabel}>
          <FontIcon iconName="Contact" className={styles.icon} />
          <FormattedMessage id="Onboarding.identities.label" />
        </Label>
        <div className={styles.identityList}>
          {identitiesButtonItems.map((btn, idx) => {
            return (
              <IdentitiesItemContent
                form={form}
                identityFeatureConfig={identityFeatureConfig}
                btnItem={btn}
                key={`identity-item-${idx}`}
              />
            );
          })}
        </div>
        {showUsernameOnlyAlert && (
          <Text className={styles.alertText} block={true} variant="small">
            <FontIcon iconName="AlertSolid" className={styles.icon} />
            <FormattedMessage id="Onboarding.identities.username-only-alert" />
          </Text>
        )}
        {showAnonymousOnlyAlert && (
          <Text className={styles.alertText} block={true} variant="small">
            <FontIcon iconName="AlertSolid" className={styles.icon} />
            <FormattedMessage id="Onboarding.identities.anonymous-users-only-alert" />
          </Text>
        )}
      </section>
    );
  };

interface PrimaryAuthenticatorsContentProps {
  form: AppConfigFormModel<FormState>;
}

const PrimaryAuthenticatorsContent: React.FC<PrimaryAuthenticatorsContentProps> =
  function PrimaryAuthenticatorsContent(props) {
    const {
      form: { state, setState },
    } = props;

    const { renderToString } = useContext(Context);
    const options: IChoiceGroupOption[] = useMemo(
      () => [
        {
          key: "password",
          text: renderToString("Onboarding.primary-authenticators.password"),
        },
        {
          key: "oob",
          text: renderToString("Onboarding.primary-authenticators.oob"),
          disabled: state.pendingForm.loginIDKeys.has("username"),
        },
      ],
      [renderToString, state.pendingForm.loginIDKeys]
    );

    const onChange = useCallback(
      (_event, option?: IChoiceGroupOption) => {
        if (option?.key) {
          setState((prev) => ({
            ...prev,
            pendingForm: {
              ...prev.pendingForm,
              primaryAuthenticator: option.key as PrimaryAuthenticatorType,
            },
          }));
        }
      },
      [setState]
    );

    return (
      <section className={styles.sections}>
        <LabelWithTooltip
          className={styles.fieldLabel}
          labelId="Onboarding.primary-authenticators.title"
          tooltipHeaderId=""
          tooltipMessageId="Onboarding.primary-authenticators.tooltip-message"
          directionalHint={DirectionalHint.bottomLeftEdge}
          labelIIconProps={{ iconName: "AutoFillTemplate" }}
        />
        <ChoiceGroup
          selectedKey={state.pendingForm.primaryAuthenticator}
          options={options}
          onChange={onChange}
        />
      </section>
    );
  };

interface SecondaryAuthenticationModeContentProps {
  form: AppConfigFormModel<FormState>;
}

const SecondaryAuthenticationModeContent: React.FC<SecondaryAuthenticationModeContentProps> =
  function SecondaryAuthenticationModeContent(props) {
    const {
      form: { state, setState },
    } = props;

    const { renderToString } = useContext(Context);
    const options: IDropdownOption[] = useMemo(
      () => [
        {
          key: "required",
          text: renderToString(
            "Onboarding.secondary-authentication-mode.required"
          ),
        },
        {
          key: "if_exists",
          text: renderToString(
            "Onboarding.secondary-authentication-mode.if-exists"
          ),
        },
        {
          key: "if_requested",
          text: renderToString(
            "Onboarding.secondary-authentication-mode.if-requested"
          ),
        },
      ],
      [renderToString]
    );

    const onChange = useCallback(
      (_event, option?: IDropdownOption) => {
        if (option?.key) {
          setState((prev) => ({
            ...prev,
            pendingForm: {
              ...prev.pendingForm,
              secondaryAuthenticationMode:
                option.key as SecondaryAuthenticationMode,
            },
          }));
        }
      },
      [setState]
    );

    return (
      <section className={styles.sections}>
        <Label className={styles.fieldLabel}>
          <FontIcon iconName="Permissions" className={styles.icon} />
          <FormattedMessage id="Onboarding.secondary-authentication-mode.title" />
        </Label>
        <Dropdown
          options={options}
          selectedKey={state.pendingForm.secondaryAuthenticationMode}
          onChange={onChange}
        />
        <Text className={styles.helpText} block={true} variant="small">
          <FormattedMessage id="Onboarding.secondary-authentication-mode.desc" />
        </Text>
      </section>
    );
  };

interface SecondaryAuthenticatorOption {
  labelId: string;
  authenticatorType: SecondaryAuthenticatorType;
}

interface SecondaryAuthenticatorCheckboxProps {
  form: AppConfigFormModel<FormState>;
  authenticatorsFeatureConfig?: AuthenticatorsFeatureConfig;
  option: SecondaryAuthenticatorOption;
}

const SecondaryAuthenticatorCheckbox: React.FC<SecondaryAuthenticatorCheckboxProps> =
  function SecondaryAuthenticatorCheckbox(props) {
    const { renderToString } = useContext(Context);
    const {
      form: { state, setState },
      authenticatorsFeatureConfig,
      option,
    } = props;

    const getCheckedState = useCallback(
      (authenticatorType: SecondaryAuthenticatorType) =>
        state.pendingForm.secondaryAuthenticators.has(authenticatorType),
      [state]
    );

    const onChange = useCallback(
      (_event, checked?: boolean) => {
        const secondaryAuthenticators = new Set(
          state.pendingForm.secondaryAuthenticators
        );
        if (checked) {
          secondaryAuthenticators.add(option.authenticatorType);
        } else {
          secondaryAuthenticators.delete(option.authenticatorType);
        }

        setState((prev) => ({
          ...prev,
          pendingForm: {
            ...prev.pendingForm,
            secondaryAuthenticators,
          },
        }));
      },
      [state, setState, option]
    );

    const disabled = useMemo(() => {
      return (
        (option.authenticatorType === "oob_otp_sms" &&
          authenticatorsFeatureConfig?.oob_otp_sms?.disabled) ??
        false
      );
    }, [
      option.authenticatorType,
      authenticatorsFeatureConfig?.oob_otp_sms?.disabled,
    ]);

    return (
      <div className={styles.checkboxGroup}>
        <Checkbox
          className={styles.checkbox}
          checked={getCheckedState(option.authenticatorType)}
          label={renderToString(option.labelId)}
          onChange={onChange}
          disabled={disabled}
        />
        {disabled && (
          <UpgradeButton
            labelMessageID="Onboarding.upgrade.label"
            headerMessageID="Onboarding.upgrade.support-sms.title"
            bodyMessageID="Onboarding.upgrade.support-sms.desc"
          />
        )}
      </div>
    );
  };

interface SecondaryAuthenticatorsContentProps {
  form: AppConfigFormModel<FormState>;
  authenticatorsFeatureConfig?: AuthenticatorsFeatureConfig;
}

const SecondaryAuthenticatorsContent: React.FC<SecondaryAuthenticatorsContentProps> =
  function SecondaryAuthenticatorsContent(props) {
    const { form, authenticatorsFeatureConfig } = props;

    return (
      <section className={styles.sections}>
        <Label className={styles.fieldLabel}>
          <FontIcon iconName="PlayerSettings" className={styles.icon} />
          <FormattedMessage id="Onboarding.secondary-authenticators.title" />
        </Label>
        {secondaryAuthenticatorOptions.map((o, idx) => (
          <SecondaryAuthenticatorCheckbox
            authenticatorsFeatureConfig={authenticatorsFeatureConfig}
            key={`secondary-authenticator-${idx}`}
            form={form}
            option={o}
          />
        ))}
      </section>
    );
  };

interface VerificationContentProps {
  form: AppConfigFormModel<FormState>;
  labelIconName: string;
  labelId: string;
  //  email or phone_number
  claimName: keyof VerificationClaimsConfig;
}

const VerificationContent: React.FC<VerificationContentProps> =
  function VerificationContent(props) {
    const {
      form: { state, setState },
      labelIconName,
      labelId,
      claimName,
    } = props;

    const { renderToString } = useContext(Context);
    const options: IDropdownOption[] = useMemo(
      () => [
        {
          key: "required",
          text: renderToString("Onboarding.verification.required"),
        },
        {
          key: "optional",
          text: renderToString("Onboarding.verification.optional"),
        },
        {
          key: "disabled",
          text: renderToString("Onboarding.verification.disabled"),
        },
      ],
      [renderToString]
    );

    const selectedKey = useMemo(() => {
      if (state.pendingForm.verificationClaims[claimName]?.enabled) {
        if (state.pendingForm.verificationClaims[claimName]?.required) {
          return "required";
        }
        return "optional";
      }
      return "disabled";
    }, [state, claimName]);

    const onChange = useCallback(
      (_event, option?: IDropdownOption) => {
        let enabled = false;
        let required = false;
        switch (option?.key) {
          case "required":
            enabled = true;
            required = true;
            break;
          case "optional":
            enabled = true;
            required = false;
            break;
          case "disabled":
            enabled = false;
            required = false;
            break;
          default:
            return;
        }

        setState((prev) => ({
          ...prev,
          pendingForm: {
            ...prev.pendingForm,
            verificationClaims: {
              ...prev.pendingForm.verificationClaims,
              [claimName]: {
                required,
                enabled,
              },
            },
          },
        }));
      },
      [setState, claimName]
    );

    return (
      <section className={styles.sections}>
        <Label className={styles.fieldLabel}>
          <FontIcon iconName={labelIconName} className={styles.icon} />
          <FormattedMessage id={labelId} />
        </Label>
        <Dropdown
          options={options}
          selectedKey={selectedKey}
          onChange={onChange}
        />
        <Text className={styles.helpText} block={true} variant="small">
          <FormattedMessage id="Onboarding.verification.desc" />
        </Text>
      </section>
    );
  };

interface OnboardingConfigAppScreenFormProps {
  form: AppConfigFormModel<FormState>;
  featureConfig?: PortalAPIFeatureConfig;
}

const OnboardingConfigAppScreenForm: React.FC<OnboardingConfigAppScreenFormProps> =
  function OnboardingConfigAppScreenForm(props) {
    const { form, featureConfig } = props;

    const showPrimaryAuthenticators = useMemo(
      () => form.state.pendingForm.identities.has("login_id"),
      [form.state.pendingForm.identities]
    );

    // oauth and anonymous doesn't have 2fa
    const showSecondaryAuthenticationMode = useMemo(
      () => form.state.pendingForm.identities.has("login_id"),
      [form.state.pendingForm.identities]
    );

    // oauth and anonymous doesn't have 2fa
    const showSecondaryAuthenticators = useMemo(
      () => form.state.pendingForm.identities.has("login_id"),
      [form.state.pendingForm.identities]
    );

    const showEmailVerification = useMemo(
      () =>
        form.state.pendingForm.identities.has("login_id") &&
        form.state.pendingForm.loginIDKeys.has("email"),
      [form.state.pendingForm.identities, form.state.pendingForm.loginIDKeys]
    );

    const showPhoneVerification = useMemo(
      () =>
        form.state.pendingForm.identities.has("login_id") &&
        form.state.pendingForm.loginIDKeys.has("phone"),
      [form.state.pendingForm.identities, form.state.pendingForm.loginIDKeys]
    );

    return (
      <div>
        <Text className={styles.pageTitle} block={true} variant="xLarge">
          <FormattedMessage id="Onboarding.title" />
        </Text>
        <Text className={styles.pageDesc} block={true} variant="small">
          <FormattedMessage id="Onboarding.desc" />
        </Text>
        <IdentitiesListContent
          form={form}
          identityFeatureConfig={featureConfig?.identity}
        />
        {showPrimaryAuthenticators && (
          <PrimaryAuthenticatorsContent form={form} />
        )}
        {showSecondaryAuthenticationMode && (
          <SecondaryAuthenticationModeContent form={form} />
        )}
        {showSecondaryAuthenticators && (
          <SecondaryAuthenticatorsContent
            form={form}
            authenticatorsFeatureConfig={
              featureConfig?.authentication?.secondary_authenticators
            }
          />
        )}
        {showEmailVerification && (
          <VerificationContent
            form={form}
            claimName="email"
            labelIconName="Mail"
            labelId="Onboarding.verification.email-enabled.title"
          />
        )}
        {showPhoneVerification && (
          <VerificationContent
            form={form}
            claimName="phone_number"
            labelIconName="CellPhone"
            labelId="Onboarding.verification.phone-enabled.title"
          />
        )}
      </div>
    );
  };

interface OnboardingConfigAppScreenContentProps {
  featureConfig?: PortalAPIFeatureConfig;
}

const OnboardingConfigAppScreenContent: React.FC<OnboardingConfigAppScreenContentProps> =
  function OnboardingConfigAppScreenContent(props) {
    const { appID } = useParams();
    const navigate = useNavigate();
    const form = useAppConfigForm(
      appID,
      constructFormState,
      constructConfig,
      undefined,
      true
    );

    const {
      state: {
        pendingForm: { identities },
      },
      setCanSave,
    } = form;

    const { featureConfig } = props;

    useEffect(() => {
      if (form.isSubmitted) {
        navigate(`/project/${encodeURIComponent(appID)}/done`);
      }
    }, [form.isSubmitted, navigate, appID]);

    // Change form canSave state if selected identities is changed
    // at least one identity is needed
    useEffect(() => {
      const canSave = identities.size > 0;
      setCanSave(canSave);
    }, [identities, setCanSave]);

    if (form.isLoading) {
      return <ShowLoading />;
    }
    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    return (
      <OnboardingFormContainer form={form}>
        <OnboardingConfigAppScreenForm
          form={form}
          featureConfig={featureConfig}
        />
      </OnboardingFormContainer>
    );
  };

const OnboardingConfigAppScreen: React.FC =
  function OnboardingConfigAppScreen() {
    const { appID } = useParams();

    // NOTE: check if appID actually exist in authorized app list
    const form = useAppAndSecretConfigQuery(appID);

    const featureConfig = useAppFeatureConfigQuery(appID);

    if (form.loading || featureConfig.loading) {
      return <ShowLoading />;
    }

    const isInvalidAppID =
      form.error == null && form.effectiveAppConfig == null;
    if (isInvalidAppID) {
      return <Navigate to="/apps" replace={true} />;
    }

    if (form.error) {
      return (
        <ShowError
          error={form.error}
          onRetry={() => {
            form.refetch().finally(() => {});
          }}
        />
      );
    }

    if (featureConfig.error) {
      return (
        <ShowError
          error={featureConfig.error}
          onRetry={() => {
            featureConfig.refetch().finally(() => {});
          }}
        />
      );
    }

    return (
      <div className={styles.root}>
        <ScreenHeader />
        <OnboardingConfigAppScreenContent
          featureConfig={featureConfig.effectiveFeatureConfig ?? undefined}
        />
      </div>
    );
  };

export default OnboardingConfigAppScreen;
