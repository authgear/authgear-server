import React, { useCallback, useContext, useMemo } from "react";
import { Navigate, useParams } from "react-router-dom";
import produce from "immer";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import {
  Checkbox,
  ChoiceGroup,
  Dropdown,
  FontIcon,
  IChoiceGroupOption,
  IDropdownOption,
  Label,
  Text,
} from "@fluentui/react";

import {
  IdentityType,
  LoginIDKeyType,
  PortalAPIAppConfig,
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
import { useAppConfigQuery } from "./query/appConfigQuery";
import OnboardingFormContainer from "./OnboardingFormContainer";
import styles from "./OnboardingConfigAppScreen.module.scss";

const primaryAuthenticatorTypes = ["password", "oob"] as const;
type PrimaryAuthenticatorType = typeof primaryAuthenticatorTypes[number];

interface PendingFormState {
  identities: Set<IdentityType>;
  loginIDKeys: Set<LoginIDKeyType>;
  primaryAuthenticator: PrimaryAuthenticatorType;
  secondaryAuthenticationMode: SecondaryAuthenticationMode;
  secondaryAuthenticators: Set<SecondaryAuthenticatorType>;
  verificationClaims: VerificationClaimsConfig;
}

interface FormState {
  pendingForm: PendingFormState;
}

function constructFormState(_config: PortalAPIAppConfig): FormState {
  return {
    pendingForm: {
      identities: new Set<IdentityType>(),
      loginIDKeys: new Set<LoginIDKeyType>(),
      primaryAuthenticator: "password",
      secondaryAuthenticationMode: "if_exists",
      secondaryAuthenticators: new Set<SecondaryAuthenticatorType>([
        "totp",
        "oob_otp_sms",
      ]),
      verificationClaims: {
        email: { enabled: true, required: true },
        phone_number: { enabled: true, required: true },
      },
    },
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {});
}

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
  btnItem: IdentitiesButton;
}

const IdentitiesItemContent: React.FC<IdentitiesItemContentProps> = function IdentitiesItemContent(
  props
) {
  const {
    form: { state, setState },
    btnItem,
  } = props;

  const getCheckedState = useCallback(() => {
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

      setState((prev) => ({
        ...prev,
        pendingForm: {
          ...prev.pendingForm,
          identities,
          loginIDKeys,
        },
      }));
    },
    [state, setState, btnItem]
  );

  const onItemClick = useCallback(
    (event: React.FormEvent) => {
      event.preventDefault();
      event.stopPropagation();
      const currentChecked = getCheckedState();
      onCheckedChange(!currentChecked);
    },
    [getCheckedState, onCheckedChange]
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
    <div className={styles.identityListItem} onClick={onItemClick}>
      <div className={styles.label}>
        <FontIcon iconName={btnItem.iconName} className={styles.icon} />
        <Text block={true} variant="medium">
          <FormattedMessage id={btnItem.labelId} />
        </Text>
      </div>
      <Checkbox
        className={styles.checkbox}
        checked={getCheckedState()}
        onChange={onCheckboxChange}
      />
    </div>
  );
};

interface IdentitiesListContentProps {
  form: AppConfigFormModel<FormState>;
}

const IdentitiesListContent: React.FC<IdentitiesListContentProps> = function IdentitiesListContent(
  props
) {
  const { form } = props;

  const identitiesButtonItems: IdentitiesButton[] = useMemo(
    () => [
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
    ],
    []
  );

  const showUsernameOnlyAlert = useMemo(
    () =>
      form.state.pendingForm.loginIDKeys.size === 1 &&
      form.state.pendingForm.loginIDKeys.has("username"),
    [form.state.pendingForm.loginIDKeys]
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
    </section>
  );
};

interface PrimaryAuthenticatorsContentProps {
  form: AppConfigFormModel<FormState>;
}

const PrimaryAuthenticatorsContent: React.FC<PrimaryAuthenticatorsContentProps> = function PrimaryAuthenticatorsContent(
  props
) {
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
      <Label className={styles.fieldLabel}>
        <FontIcon iconName="AutoFillTemplate" className={styles.icon} />
        <FormattedMessage id="Onboarding.primary-authenticators.title" />
      </Label>
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

const SecondaryAuthenticationModeContent: React.FC<SecondaryAuthenticationModeContentProps> = function SecondaryAuthenticationModeContent(
  props
) {
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
            secondaryAuthenticationMode: option.key as SecondaryAuthenticationMode,
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
      <p className={styles.helpText}>
        <FormattedMessage id="Onboarding.secondary-authentication-mode.desc" />
      </p>
    </section>
  );
};

interface SecondaryAuthenticatorOption {
  labelId: string;
  authenticatorType: SecondaryAuthenticatorType;
}

interface SecondaryAuthenticatorCheckboxProps {
  form: AppConfigFormModel<FormState>;
  option: SecondaryAuthenticatorOption;
}

const SecondaryAuthenticatorCheckbox: React.FC<SecondaryAuthenticatorCheckboxProps> = function SecondaryAuthenticatorCheckbox(
  props
) {
  const { renderToString } = useContext(Context);
  const {
    form: { state, setState },
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

  return (
    <Checkbox
      className={styles.checkboxGroup}
      checked={getCheckedState(option.authenticatorType)}
      label={renderToString(option.labelId)}
      onChange={onChange}
    />
  );
};

interface SecondaryAuthenticatorsContentProps {
  form: AppConfigFormModel<FormState>;
}

const SecondaryAuthenticatorsContent: React.FC<SecondaryAuthenticatorsContentProps> = function SecondaryAuthenticatorsContent(
  props
) {
  const { form } = props;

  const options: SecondaryAuthenticatorOption[] = useMemo(
    () => [
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
    ],
    []
  );

  return (
    <section className={styles.sections}>
      <Label className={styles.fieldLabel}>
        <FontIcon iconName="PlayerSettings" className={styles.icon} />
        <FormattedMessage id="Onboarding.secondary-authenticators.title" />
      </Label>
      {options.map((o, idx) => (
        <SecondaryAuthenticatorCheckbox
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

const VerificationContent: React.FC<VerificationContentProps> = function VerificationContent(
  props
) {
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

  const getSelectedKey = useCallback(() => {
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
        selectedKey={getSelectedKey()}
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
}

const OnboardingConfigAppScreenForm: React.FC<OnboardingConfigAppScreenFormProps> = function OnboardingConfigAppScreenForm(
  props
) {
  const { form } = props;
  return (
    <div>
      <Text className={styles.pageTitle} block={true} variant="xLarge">
        <FormattedMessage id="Onboarding.title" />
      </Text>
      <Text className={styles.pageDesc} block={true} variant="small">
        <FormattedMessage id="Onboarding.desc" />
      </Text>
      <IdentitiesListContent form={form} />
      <PrimaryAuthenticatorsContent form={form} />
      <SecondaryAuthenticationModeContent form={form} />
      <SecondaryAuthenticatorsContent form={form} />
      <VerificationContent
        form={form}
        claimName="email"
        labelIconName="Mail"
        labelId="Onboarding.verification.email-enabled.title"
      />
      <VerificationContent
        form={form}
        claimName="phone_number"
        labelIconName="CellPhone"
        labelId="Onboarding.verification.phone-enabled.title"
      />
    </div>
  );
};

const OnboardingConfigAppScreenContent: React.FC = function OnboardingConfigAppScreenContent() {
  const { appID } = useParams();
  const form = useAppConfigForm(appID, constructFormState, constructConfig);

  if (form.isLoading) {
    return <ShowLoading />;
  }
  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <OnboardingFormContainer form={form}>
      <OnboardingConfigAppScreenForm form={form} />
    </OnboardingFormContainer>
  );
};

const OnboardingConfigAppScreen: React.FC = function OnboardingConfigAppScreen() {
  const { appID } = useParams();

  // NOTE: check if appID actually exist in authorized app list
  const { effectiveAppConfig, loading, error } = useAppConfigQuery(appID);
  if (loading) {
    return <ShowLoading />;
  }
  const isInvalidAppID = error == null && effectiveAppConfig == null;
  if (isInvalidAppID) {
    return <Navigate to="/apps" replace={true} />;
  }

  return (
    <div className={styles.root}>
      <ScreenHeader />
      <OnboardingConfigAppScreenContent />
    </div>
  );
};

export default OnboardingConfigAppScreen;
