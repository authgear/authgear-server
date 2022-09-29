import React, {
  ReactNode,
  useMemo,
  useCallback,
  useState,
  useEffect,
  useContext,
} from "react";
import cn from "classnames";
import {
  MessageBar,
  MessageBarType,
  Text,
  useTheme,
  IButtonProps,
} from "@fluentui/react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import {
  PortalAPIAppConfig,
  PortalAPIFeatureConfig,
  IdentityType,
  PrimaryAuthenticatorType,
  LoginIDKeyConfig,
  LoginIDKeyType,
} from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import ChoiceButton, { ChoiceButtonProps } from "../../ChoiceButton";
import Link from "../../Link";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import PriorityList from "../../PriorityList";
import WidgetDescription from "../../WidgetDescription";
import HorizontalDivider from "../../HorizontalDivider";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import PrimaryButton from "../../PrimaryButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { makeValidationErrorMatchUnknownKindParseRule } from "../../error/parse";
import styles from "./LoginMethodConfigurationScreen.module.css";

const ERROR_RULES = [
  makeValidationErrorMatchUnknownKindParseRule(
    "const",
    /\/authentication\/identities/,
    "errors.validation.passkey"
  ),
];

const IDENTITY_TYPES: IdentityType[] = ["login_id"];
const PRIMARY_AUTHENTICATOR_TYPES: PrimaryAuthenticatorType[] = [
  "password",
  "oob_otp_email",
  "oob_otp_sms",
];
const LOGIN_ID_KEY_CONFIGS: LoginIDKeyConfig[] = [
  { type: "email" },
  { type: "phone" },
  { type: "username" },
];

interface ControlOf<T> {
  isChecked: boolean;
  isDisabled: boolean;
  value: T;
}

// ControlList augments T with isChecked and isDisabled.
//
// controlListOf creates a list of ControlOf that this screen recognize.
//
// controlListPreserve turns a ControlList into plain list by preserving exotic values.
// This is useful for identities and primary_authenticators because
// "biometric", "anonymous", and "passkey" are exotic.
// They must be preserved.
//
// controlListUnwrap simply turns a ControlList into plain list.
//
// controlListIsEqualToPlainList determines whether a ControlList is equal to a plain list.
//
// controlListCheckWithPlainList checks a ControlList with a plain list.
type ControlList<T> = ControlOf<T>[];

function controlListOf<T>(
  eq: (a: T, b: T) => boolean,
  all: T[],
  current: T[]
): ControlList<T> {
  const out: ControlList<T> = [];

  for (const a of current) {
    const b = all.find((b) => eq(a, b));
    if (b != null) {
      out.push({
        isChecked: true,
        isDisabled: false,
        value: a,
      });
    }
  }

  for (const a of all) {
    const b = out.find((b) => eq(a, b.value));
    if (b == null) {
      out.push({
        isChecked: false,
        isDisabled: false,
        value: a,
      });
    }
  }
  return out;
}

function controlListIsEqualToPlainList<U, T>(
  eq: (u: U, t: T) => boolean,
  us: U[],
  ts: ControlList<T>
): boolean {
  const plains = ts.filter((t) => t.isChecked).map((t) => t.value);

  if (plains.length !== us.length) {
    return false;
  }

  for (let i = 0; i < us.length; i++) {
    const u = us[i];
    const t = plains[i];
    if (!eq(u, t)) {
      return false;
    }
  }

  return true;
}

function controlListPreserve<T>(
  eq: (a: T, b: T) => boolean,
  ts: ControlList<T>,
  plains: T[]
): T[] {
  const exotic = plains.filter((a) => {
    for (const t of ts) {
      if (eq(a, t.value)) {
        return false;
      }
    }
    return true;
  });

  return [...ts.filter((a) => a.isChecked).map((a) => a.value), ...exotic];
}

function controlListUnwrap<T>(ts: ControlList<T>): T[] {
  return ts.filter((a) => a.isChecked).map((a) => a.value);
}

function controlListCheckWithPlainList<U, T>(
  eq: (u: U, t: T) => boolean,
  us: U[],
  ts: ControlList<T>
): ControlList<T> {
  const checked: ControlList<T> = [];

  for (const u of us) {
    const t = ts.find((t) => eq(u, t.value));
    if (t != null) {
      checked.push({
        ...t,
        isChecked: true,
      });
    }
  }

  const unchecked: ControlList<T> = ts
    .filter((t) => {
      for (const u of us) {
        if (eq(u, t.value)) {
          return false;
        }
      }
      return true;
    })
    .map((t) => ({
      ...t,
      isChecked: false,
    }));

  const out: ControlList<T> = [...checked, ...unchecked];
  return out;
}

function controlListCheckWithPlainValue<U, T>(
  eq: (u: U, t: T) => boolean,
  u: U,
  isChecked: boolean,
  ts: ControlList<T>
): ControlList<T> {
  return ts.map((t) => {
    if (eq(u, t.value)) {
      return {
        ...t,
        isChecked,
      };
    }
    return t;
  });
}

function controlListSwap<T>(
  index1: number,
  index2: number,
  ts: ControlList<T>
): ControlList<T> {
  const newItems = [...ts];
  const thisItem = newItems[index1];
  const thatItem = newItems[index2];
  if (index1 < 0 || index2 < 0 || index1 >= ts.length || index2 >= ts.length) {
    return ts;
  }
  newItems[index1] = thatItem;
  newItems[index2] = thisItem;
  return newItems;
}

interface FormState {
  identitiesControl: ControlList<IdentityType>;
  primaryAuthenticatorsControl: ControlList<PrimaryAuthenticatorType>;
  loginIDKeyConfigsControl: ControlList<LoginIDKeyConfig>;
}

// eslint-disable-next-line complexity
function correctInitialFormState(state: FormState): void {
  // Uncheck "login_id" identity if no login ID is checked.
  const allLoginIDUnchecked = state.loginIDKeyConfigsControl.every(
    (a) => !a.isChecked
  );
  if (allLoginIDUnchecked) {
    for (const t of state.identitiesControl) {
      if (t.value === "login_id") {
        t.isChecked = false;
      }
    }
  }

  // Disable "oob_otp_sms" or "oob_otp_email" if the corresponding login ID is unchecked.
  // Note that we do NOT uncheck.
  for (const loginID of state.loginIDKeyConfigsControl) {
    for (const authenticator of state.primaryAuthenticatorsControl) {
      if (
        loginID.value.type === "email" &&
        authenticator.value === "oob_otp_email"
      ) {
        authenticator.isDisabled = !loginID.isChecked;
      }
      if (
        loginID.value.type === "phone" &&
        authenticator.value === "oob_otp_sms"
      ) {
        authenticator.isDisabled = !loginID.isChecked;
      }
    }
  }
}

// eslint-disable-next-line complexity
function correctCurrentFormState(state: FormState): void {
  // Check or uncheck "login_id" identity.
  const allLoginIDUnchecked = state.loginIDKeyConfigsControl.every(
    (a) => !a.isChecked
  );
  const someLoginIDChecked = !allLoginIDUnchecked;
  for (const t of state.identitiesControl) {
    if (t.value === "login_id") {
      t.isChecked = someLoginIDChecked;
    }
  }

  // Disable and unchecked "oob_otp_sms" or "oob_otp_email" if the corresponding login ID is unchecked.
  for (const loginID of state.loginIDKeyConfigsControl) {
    for (const authenticator of state.primaryAuthenticatorsControl) {
      if (
        loginID.value.type === "email" &&
        authenticator.value === "oob_otp_email"
      ) {
        authenticator.isDisabled = !loginID.isChecked;
        if (authenticator.isDisabled && authenticator.isChecked) {
          authenticator.isChecked = false;
        }
      }
      if (
        loginID.value.type === "phone" &&
        authenticator.value === "oob_otp_sms"
      ) {
        authenticator.isDisabled = !loginID.isChecked;
        if (authenticator.isDisabled && authenticator.isChecked) {
          authenticator.isChecked = false;
        }
      }
    }
  }
}

function loginIDIdentity(identities: ControlList<IdentityType>): boolean {
  return controlListIsEqualToPlainList(
    (a, b) => a === b,
    ["login_id"] as IdentityType[],
    identities
  );
}

function oauthIdentity(identities: ControlList<IdentityType>): boolean {
  return controlListIsEqualToPlainList(
    (a, b) => a === b,
    [] as IdentityType[],
    identities
  );
}

function loginIDOf(
  types: LoginIDKeyType[],
  loginIDKeyConfigs: ControlList<LoginIDKeyConfig>
): boolean {
  return controlListIsEqualToPlainList(
    (u, t) => {
      return u === t.type;
    },
    types,
    loginIDKeyConfigs
  );
}

function primaryAuthenticatorOf(
  types: PrimaryAuthenticatorType[],
  primaryAuthenticators: ControlList<PrimaryAuthenticatorType>
): boolean {
  return controlListIsEqualToPlainList(
    (u, t) => {
      return u === t;
    },
    types,
    primaryAuthenticators
  );
}

function setLoginIDIdentity(draft: FormState) {
  draft.identitiesControl = controlListCheckWithPlainList(
    (a, b) => a === b,
    ["login_id"] as IdentityType[],
    draft.identitiesControl
  );
}

function setOAuthIdentity(draft: FormState) {
  draft.identitiesControl = controlListCheckWithPlainList(
    (a, b) => a === b,
    [] as IdentityType[],
    draft.identitiesControl
  );
}

function setLoginID(draft: FormState, types: LoginIDKeyType[]) {
  draft.loginIDKeyConfigsControl = controlListCheckWithPlainList(
    (a, b) => a === b.type,
    types,
    draft.loginIDKeyConfigsControl
  );
}

function setPrimaryAuthenticator(
  draft: FormState,
  types: PrimaryAuthenticatorType[]
) {
  draft.primaryAuthenticatorsControl = controlListCheckWithPlainList(
    (a, b) => a === b,
    types,
    draft.primaryAuthenticatorsControl
  );
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  const identities = config.authentication?.identities ?? [];
  const primaryAuthenticators =
    config.authentication?.primary_authenticators ?? [];
  const loginIDKeyConfigs = config.identity?.login_id?.keys ?? [];

  const state = {
    identitiesControl: controlListOf(
      (a, b) => a === b,
      IDENTITY_TYPES,
      identities
    ),
    primaryAuthenticatorsControl: controlListOf(
      (a, b) => a === b,
      PRIMARY_AUTHENTICATOR_TYPES,
      primaryAuthenticators
    ),
    loginIDKeyConfigsControl: controlListOf(
      (a, b) => a.type === b.type,
      LOGIN_ID_KEY_CONFIGS,
      loginIDKeyConfigs
    ),
  };
  correctInitialFormState(state);
  return state;
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.authentication ??= {};
    config.identity ??= {};
    config.identity.login_id ??= {};

    config.authentication.identities = controlListPreserve(
      (a, b) => a === b,
      currentState.identitiesControl,
      effectiveConfig.authentication?.identities ?? []
    );
    config.authentication.primary_authenticators = controlListPreserve(
      (a, b) => a === b,
      currentState.primaryAuthenticatorsControl,
      effectiveConfig.authentication?.primary_authenticators ?? []
    );
    config.identity.login_id.keys = controlListUnwrap(
      currentState.loginIDKeyConfigsControl
    );

    clearEmptyObject(config);
  });
}

interface MethodGroupTitleProps {
  children?: ReactNode;
}

const FIELD_TITLE_STYLES = {
  root: {
    fontWeight: "600",
  },
};

function MethodGroupTitle(props: MethodGroupTitleProps) {
  const { children } = props;
  return (
    <Text as="h3" block={true} variant="medium" styles={FIELD_TITLE_STYLES}>
      {children}
    </Text>
  );
}

interface MethodGroupProps {
  title?: ReactNode;
  children?: ReactNode;
}

function MethodGroup(props: MethodGroupProps) {
  const { title, children } = props;
  return (
    <div className={styles.methodGroup}>
      <MethodGroupTitle>{title}</MethodGroupTitle>
      <div className={styles.methodGrid}>{children}</div>
    </div>
  );
}

interface MethodButtonProps {
  checked: boolean;
  text?: ReactNode;
  secondaryText?: ReactNode;
  onClick?: IButtonProps["onClick"];
}

function MethodButton(props: MethodButtonProps) {
  const { checked, text, secondaryText, onClick } = props;
  const theme = useTheme();
  const { themes } = useSystemConfig();

  if (!checked) {
    return null;
  }

  return (
    <div
      className={cn(styles.widget, styles.methodButton)}
      style={{
        backgroundColor: theme.palette.themePrimary,
      }}
    >
      <div className={styles.methodButtonText}>
        <Text
          className={styles.methodButtonTitle}
          block={true}
          variant="large"
          theme={themes.inverted}
        >
          {text}
        </Text>
        {secondaryText != null ? (
          <Text
            className={styles.methodButtonDescription}
            block={true}
            variant="medium"
            theme={themes.inverted}
          >
            {secondaryText}
          </Text>
        ) : null}
      </div>
      <PrimaryButton
        theme={themes.inverted}
        text={
          <FormattedMessage id="LoginMethodConfigurationScreen.method.change-method" />
        }
        onClick={onClick}
      />
    </div>
  );
}

interface MethodProps
  extends Omit<MethodButtonProps, "text" | "secondaryText"> {}

function MethodEmailPasswordless(props: MethodProps) {
  return (
    <MethodButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.email.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.email.description" />
      }
    />
  );
}

function MethodPhonePasswordless(props: MethodProps) {
  return (
    <MethodButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.phone.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.phone.description" />
      }
    />
  );
}

function MethodPhoneEmailPasswordless(props: MethodProps) {
  return (
    <MethodButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.all.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.all.description" />
      }
    />
  );
}

function MethodEmailPassword(props: MethodProps) {
  return (
    <MethodButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.email.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.email.description" />
      }
    />
  );
}

function MethodPhonePassword(props: MethodProps) {
  return (
    <MethodButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.phone.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.phone.description" />
      }
    />
  );
}

function MethodPhoneEmailPassword(props: MethodProps) {
  return (
    <MethodButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.no-username.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.no-username.description" />
      }
    />
  );
}

function MethodUsernamePassword(props: MethodProps) {
  return (
    <MethodButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.username.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.username.description" />
      }
    />
  );
}

function MethodOAuthOnly(props: MethodProps) {
  return (
    <MethodButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.choice.oauth.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.choice.oauth.description" />
      }
    />
  );
}

function MethodCustom(props: MethodProps) {
  return (
    <MethodButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.choice.custom.title" />
      }
    />
  );
}

interface ChoiceProps
  extends Omit<ChoiceButtonProps, "text" | "secondaryText"> {}

function ChoiceEmailPasswordless(props: ChoiceProps) {
  return (
    <ChoiceButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.email.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.email.description" />
      }
    />
  );
}

function ChoicePhonePasswordless(props: ChoiceProps) {
  return (
    <ChoiceButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.phone.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.phone.description" />
      }
    />
  );
}

function ChoicePhoneEmailPasswordless(props: ChoiceProps) {
  return (
    <ChoiceButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.all.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.all.description" />
      }
    />
  );
}

function ChoiceEmailPassword(props: ChoiceProps) {
  return (
    <ChoiceButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.email.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.email.description" />
      }
    />
  );
}

function ChoicePhonePassword(props: ChoiceProps) {
  return (
    <ChoiceButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.phone.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.phone.description" />
      }
    />
  );
}

function ChoicePhoneEmailPassword(props: ChoiceProps) {
  return (
    <ChoiceButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.no-username.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.no-username.description" />
      }
    />
  );
}

function ChoiceUsernamePassword(props: ChoiceProps) {
  return (
    <ChoiceButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.username.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.username.description" />
      }
    />
  );
}

function ChoiceOAuthOnly(props: ChoiceProps) {
  return (
    <ChoiceButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.choice.oauth.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.choice.oauth.description" />
      }
    />
  );
}

function ChoiceCustom(props: ChoiceProps) {
  return (
    <ChoiceButton
      {...props}
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.choice.custom.title" />
      }
    />
  );
}

interface PasskeyAndOAuthHintProps {
  appID: string;
}

function PasskeyAndOAuthHint(props: PasskeyAndOAuthHintProps) {
  const { appID } = props;
  return (
    <MessageBar className={styles.widget} messageBarType={MessageBarType.info}>
      <FormattedMessage
        id="LoginMethodConfigurationScreen.passkey-and-oauth"
        values={{
          passkey: `/project/${appID}/configuration/authentication/passkey`,
          oauth: `/project/${appID}/configuration/authentication/externa`,
        }}
      />
    </MessageBar>
  );
}

interface LinkToOAuthProps {
  appID: string;
  oauthOnlyChecked: boolean;
}

function LinkToOAuth(props: LinkToOAuthProps) {
  const { appID, oauthOnlyChecked } = props;
  if (!oauthOnlyChecked) {
    return null;
  }

  return (
    <Widget className={styles.widget}>
      <Link
        className={styles.oauthLink}
        to={`/project/${appID}/configuration/authentication/external-oauth`}
      >
        <FormattedMessage id="LoginMethodConfigurationScreen.oauth" />
      </Link>
    </Widget>
  );
}

interface CustomLoginMethodsProps {
  phoneLoginIDDisabled: boolean;
  customChecked: boolean;
  primaryAuthenticatorsControl: ControlList<PrimaryAuthenticatorType>;
  loginIDKeyConfigsControl: ControlList<LoginIDKeyConfig>;
  onChangeLoginIDChecked: (key: LoginIDKeyType, checked: boolean) => void;
  onSwapLoginID: (index1: number, index2: number) => void;
  onChangePrimaryAuthenticatorChecked: (
    key: PrimaryAuthenticatorType,
    checked: boolean
  ) => void;
  onSwapPrimaryAuthenticator: (index1: number, index2: number) => void;
}

function CustomLoginMethods(props: CustomLoginMethodsProps) {
  const {
    phoneLoginIDDisabled,
    customChecked,
    loginIDKeyConfigsControl,
    primaryAuthenticatorsControl,
    onChangeLoginIDChecked: onChangeLoginIDCheckedProp,
    onSwapLoginID: onSwapLoginIDProp,
    onChangePrimaryAuthenticatorChecked:
      onChangePrimaryAuthenticatorCheckedProp,
    onSwapPrimaryAuthenticator: onSwapPrimaryAuthenticatorProp,
  } = props;

  const { renderToString } = useContext(Context);

  const {
    semanticColors: { disabledText },
  } = useTheme();

  const [extended, setExtended] = useState(false);
  const onToggleButtonClick = useCallback(() => {
    setExtended((prev) => !prev);
  }, []);

  const loginIDs = useMemo(() => {
    // FIXME: handle feature disabled
    return loginIDKeyConfigsControl.map((a) => {
      let disabled = a.isDisabled;
      if (a.value.type === "phone") {
        disabled = disabled || phoneLoginIDDisabled;
      }
      return {
        key: a.value.type,
        checked: a.isChecked,
        disabled,
        content: (
          <Text
            variant="small"
            block={true}
            styles={{
              root: {
                color: disabled ? disabledText : undefined,
              },
            }}
          >
            <FormattedMessage id={"LoginIDKeyType." + a.value.type} />
          </Text>
        ),
      };
    });
  }, [loginIDKeyConfigsControl, phoneLoginIDDisabled, disabledText]);

  const onChangeLoginIDChecked = useCallback(
    (key: string, checked: boolean) => {
      onChangeLoginIDCheckedProp(key as LoginIDKeyType, checked);
    },
    [onChangeLoginIDCheckedProp]
  );

  const onSwapLoginID = useCallback(
    (index1: number, index2: number) => {
      onSwapLoginIDProp(index1, index2);
    },
    [onSwapLoginIDProp]
  );

  const authenticators = useMemo(() => {
    return primaryAuthenticatorsControl.map((a) => {
      let disabled = a.isDisabled;
      if (a.value === "oob_otp_sms") {
        disabled = disabled || phoneLoginIDDisabled;
      }
      return {
        key: a.value,
        checked: a.isChecked,
        disabled,
        content: (
          <Text
            variant="small"
            block={true}
            styles={{
              root: {
                color: disabled ? disabledText : undefined,
              },
            }}
          >
            <FormattedMessage id={"PrimaryAuthenticatorType." + a.value} />
          </Text>
        ),
      };
    });
  }, [primaryAuthenticatorsControl, phoneLoginIDDisabled, disabledText]);

  const onChangePrimaryAuthenticatorChecked = useCallback(
    (key: string, checked: boolean) => {
      onChangePrimaryAuthenticatorCheckedProp(
        key as PrimaryAuthenticatorType,
        checked
      );
    },
    [onChangePrimaryAuthenticatorCheckedProp]
  );

  const onSwapPrimaryAuthenticator = useCallback(
    (index1: number, index2: number) => {
      onSwapPrimaryAuthenticatorProp(index1, index2);
    },
    [onSwapPrimaryAuthenticatorProp]
  );

  if (!customChecked) {
    return null;
  }

  return (
    <Widget
      className={styles.widget}
      showToggleButton={true}
      extended={extended}
      onToggleButtonClick={onToggleButtonClick}
    >
      <WidgetTitle>
        <FormattedMessage id="LoginMethodConfigurationScreen.custom-login-methods.title" />
      </WidgetTitle>
      {phoneLoginIDDisabled ? (
        <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
      ) : null}
      <div className={styles.methodGroup}>
        <MethodGroupTitle>
          <FormattedMessage id="LoginMethodConfigurationScreen.custom-login-methods.login-id.title" />
        </MethodGroupTitle>
        <WidgetDescription>
          <FormattedMessage id="LoginMethodConfigurationScreen.custom-login-methods.login-id.description" />
        </WidgetDescription>
      </div>
      <PriorityList
        items={loginIDs}
        checkedColumnLabel={renderToString("activate")}
        keyColumnLabel={renderToString(
          "LoginMethodConfigurationScreen.custom-login-methods.login-id.title"
        )}
        onChangeChecked={onChangeLoginIDChecked}
        onSwap={onSwapLoginID}
      />
      <HorizontalDivider />
      <div className={styles.methodGroup}>
        <MethodGroupTitle>
          <FormattedMessage id="LoginMethodConfigurationScreen.custom-login-methods.authenticator.title" />
        </MethodGroupTitle>
        <WidgetDescription>
          <FormattedMessage id="LoginMethodConfigurationScreen.custom-login-methods.authenticator.description" />
        </WidgetDescription>
      </div>
      <PriorityList
        items={authenticators}
        checkedColumnLabel={renderToString("activate")}
        keyColumnLabel={renderToString(
          "LoginMethodConfigurationScreen.custom-login-methods.authenticator.title"
        )}
        onChangeChecked={onChangePrimaryAuthenticatorChecked}
        onSwap={onSwapPrimaryAuthenticator}
      />
    </Widget>
  );
}

interface LoginMethodConfigurationContentProps {
  appID: string;
  form: AppConfigFormModel<FormState>;
  featureConfig?: PortalAPIFeatureConfig;
}

const LoginMethodConfigurationContent: React.VFC<LoginMethodConfigurationContentProps> =
  // eslint-disable-next-line complexity
  function LoginMethodConfigurationContent(props) {
    const { appID, featureConfig } = props;
    const { state, setState } = props.form;

    const {
      identitiesControl,
      primaryAuthenticatorsControl,
      loginIDKeyConfigsControl,
    } = state;

    const [isChoosingMethod, setIsChoosingMethod] = useState(false);
    const [customChecked, setCustomChecked] = useState(false);

    const phoneLoginIDDisabled =
      featureConfig?.identity?.login_id?.types?.phone?.disabled ?? false;

    const phonePasswordlessDisabled = phoneLoginIDDisabled;
    const phoneEmailPasswordlessDisabled = phoneLoginIDDisabled;
    const phonePasswordDisabled = phoneLoginIDDisabled;
    const phoneEmailPasswordDisabled = phoneLoginIDDisabled;

    const onClickChooseLoginMethod = useCallback((e) => {
      e.preventDefault();
      e.stopPropagation();
      setIsChoosingMethod(true);
    }, []);

    const onCustomClick = useCallback((e) => {
      e.preventDefault();
      e.stopPropagation();
      setIsChoosingMethod(false);
      setCustomChecked(true);
    }, []);

    const emailPasswordlessChecked = useMemo(() => {
      return (
        loginIDIdentity(identitiesControl) &&
        loginIDOf(["email"], loginIDKeyConfigsControl) &&
        primaryAuthenticatorOf(["oob_otp_email"], primaryAuthenticatorsControl)
      );
    }, [
      identitiesControl,
      loginIDKeyConfigsControl,
      primaryAuthenticatorsControl,
    ]);

    const onEmailPasswordlessClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        setIsChoosingMethod(false);
        setCustomChecked(false);
        setState((prev) =>
          produce(prev, (prev) => {
            setLoginIDIdentity(prev);
            setLoginID(prev, ["email"]);
            setPrimaryAuthenticator(prev, ["oob_otp_email"]);
          })
        );
      },
      [setState]
    );

    const phonePasswordlessChecked = useMemo(() => {
      return (
        loginIDIdentity(identitiesControl) &&
        loginIDOf(["phone"], loginIDKeyConfigsControl) &&
        primaryAuthenticatorOf(["oob_otp_sms"], primaryAuthenticatorsControl)
      );
    }, [
      identitiesControl,
      loginIDKeyConfigsControl,
      primaryAuthenticatorsControl,
    ]);

    const onPhonePasswordlessClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        setIsChoosingMethod(false);
        setCustomChecked(false);
        setState((prev) =>
          produce(prev, (prev) => {
            setLoginIDIdentity(prev);
            setLoginID(prev, ["phone"]);
            setPrimaryAuthenticator(prev, ["oob_otp_sms"]);
          })
        );
      },
      [setState]
    );

    const phoneEmailPasswordlessChecked = useMemo(() => {
      return (
        loginIDIdentity(identitiesControl) &&
        // Order is important.
        loginIDOf(["phone", "email"], loginIDKeyConfigsControl) &&
        primaryAuthenticatorOf(
          ["oob_otp_sms", "oob_otp_email"],
          primaryAuthenticatorsControl
        )
      );
    }, [
      identitiesControl,
      loginIDKeyConfigsControl,
      primaryAuthenticatorsControl,
    ]);

    const onPhoneEmailPasswordlessClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        setIsChoosingMethod(false);
        setCustomChecked(false);
        setState((prev) =>
          produce(prev, (prev) => {
            setLoginIDIdentity(prev);
            setLoginID(prev, ["phone", "email"]);
            setPrimaryAuthenticator(prev, ["oob_otp_sms", "oob_otp_email"]);
          })
        );
      },
      [setState]
    );

    const emailPasswordChecked = useMemo(() => {
      return (
        loginIDIdentity(identitiesControl) &&
        loginIDOf(["email"], loginIDKeyConfigsControl) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticatorsControl)
      );
    }, [
      identitiesControl,
      loginIDKeyConfigsControl,
      primaryAuthenticatorsControl,
    ]);

    const onEmailPasswordClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        setIsChoosingMethod(false);
        setCustomChecked(false);
        setState((prev) =>
          produce(prev, (prev) => {
            setLoginIDIdentity(prev);
            setLoginID(prev, ["email"]);
            setPrimaryAuthenticator(prev, ["password"]);
          })
        );
      },
      [setState]
    );

    const phonePasswordChecked = useMemo(() => {
      return (
        loginIDIdentity(identitiesControl) &&
        loginIDOf(["phone"], loginIDKeyConfigsControl) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticatorsControl)
      );
    }, [
      identitiesControl,
      loginIDKeyConfigsControl,
      primaryAuthenticatorsControl,
    ]);

    const onPhonePasswordClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        setIsChoosingMethod(false);
        setCustomChecked(false);
        setState((prev) =>
          produce(prev, (prev) => {
            setLoginIDIdentity(prev);
            setLoginID(prev, ["phone"]);
            setPrimaryAuthenticator(prev, ["password"]);
          })
        );
      },
      [setState]
    );

    const phoneEmailPasswordChecked = useMemo(() => {
      return (
        loginIDIdentity(identitiesControl) &&
        // Order is important.
        loginIDOf(["phone", "email"], loginIDKeyConfigsControl) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticatorsControl)
      );
    }, [
      identitiesControl,
      loginIDKeyConfigsControl,
      primaryAuthenticatorsControl,
    ]);

    const onPhoneEmailPasswordClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        setIsChoosingMethod(false);
        setCustomChecked(false);
        setState((prev) =>
          produce(prev, (prev) => {
            setLoginIDIdentity(prev);
            setLoginID(prev, ["phone", "email"]);
            setPrimaryAuthenticator(prev, ["password"]);
          })
        );
      },
      [setState]
    );

    const usernamePasswordChecked = useMemo(() => {
      return (
        loginIDIdentity(identitiesControl) &&
        loginIDOf(["username"], loginIDKeyConfigsControl) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticatorsControl)
      );
    }, [
      identitiesControl,
      loginIDKeyConfigsControl,
      primaryAuthenticatorsControl,
    ]);

    const onUsernamePasswordClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        setIsChoosingMethod(false);
        setCustomChecked(false);
        setState((prev) =>
          produce(prev, (prev) => {
            setLoginIDIdentity(prev);
            setLoginID(prev, ["username"]);
            setPrimaryAuthenticator(prev, ["password"]);
          })
        );
      },
      [setState]
    );

    const oauthOnlyChecked = useMemo(() => {
      return (
        oauthIdentity(identitiesControl) &&
        loginIDOf([], loginIDKeyConfigsControl) &&
        primaryAuthenticatorOf([], primaryAuthenticatorsControl)
      );
    }, [
      identitiesControl,
      loginIDKeyConfigsControl,
      primaryAuthenticatorsControl,
    ]);

    const onOAuthOnlyClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
        setIsChoosingMethod(false);
        setCustomChecked(false);
        setState((prev) =>
          produce(prev, (prev) => {
            setOAuthIdentity(prev);
            setLoginID(prev, []);
            setPrimaryAuthenticator(prev, []);
          })
        );
      },
      [setState]
    );

    const onChangeLoginIDChecked = useCallback(
      (typ: LoginIDKeyType, checked: boolean) => {
        setState((prev) =>
          produce(prev, (prev) => {
            prev.loginIDKeyConfigsControl = controlListCheckWithPlainValue(
              (a, b) => a === b.type,
              typ,
              checked,
              prev.loginIDKeyConfigsControl
            );
            correctCurrentFormState(prev);
          })
        );
      },
      [setState]
    );

    const onSwapLoginID = useCallback(
      (index1: number, index2: number) => {
        setState((prev) =>
          produce(prev, (prev) => {
            prev.loginIDKeyConfigsControl = controlListSwap(
              index1,
              index2,
              prev.loginIDKeyConfigsControl
            );
          })
        );
      },
      [setState]
    );

    const onChangePrimaryAuthenticatorChecked = useCallback(
      (typ: PrimaryAuthenticatorType, checked: boolean) => {
        setState((prev) =>
          produce(prev, (prev) => {
            prev.primaryAuthenticatorsControl = controlListCheckWithPlainValue(
              (a, b) => a === b,
              typ,
              checked,
              prev.primaryAuthenticatorsControl
            );
          })
        );
      },
      [setState]
    );

    const onSwapPrimaryAuthenticator = useCallback(
      (index1: number, index2: number) => {
        setState((prev) =>
          produce(prev, (prev) => {
            prev.primaryAuthenticatorsControl = controlListSwap(
              index1,
              index2,
              prev.primaryAuthenticatorsControl
            );
          })
        );
      },
      [setState]
    );

    useEffect(() => {
      if (
        !emailPasswordlessChecked &&
        !phonePasswordlessChecked &&
        !phoneEmailPasswordlessChecked &&
        !emailPasswordChecked &&
        !phonePasswordChecked &&
        !phoneEmailPasswordChecked &&
        !usernamePasswordChecked &&
        !oauthOnlyChecked
      ) {
        setCustomChecked(true);
      }
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="LoginMethodConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="LoginMethodConfigurationScreen.description" />
        </ScreenDescription>
        <PasskeyAndOAuthHint appID={appID} />
        {!isChoosingMethod ? (
          <>
            <MethodEmailPasswordless
              checked={Boolean(!customChecked && emailPasswordlessChecked)}
              onClick={onClickChooseLoginMethod}
            />
            <MethodPhonePasswordless
              checked={Boolean(!customChecked && phonePasswordlessChecked)}
              onClick={onClickChooseLoginMethod}
            />
            <MethodPhoneEmailPasswordless
              checked={Boolean(!customChecked && phoneEmailPasswordlessChecked)}
              onClick={onClickChooseLoginMethod}
            />
            <MethodEmailPassword
              checked={Boolean(!customChecked && emailPasswordChecked)}
              onClick={onClickChooseLoginMethod}
            />
            <MethodPhonePassword
              checked={Boolean(!customChecked && phonePasswordChecked)}
              onClick={onClickChooseLoginMethod}
            />
            <MethodPhoneEmailPassword
              checked={Boolean(!customChecked && phoneEmailPasswordChecked)}
              onClick={onClickChooseLoginMethod}
            />
            <MethodUsernamePassword
              checked={Boolean(!customChecked && usernamePasswordChecked)}
              onClick={onClickChooseLoginMethod}
            />
            <MethodOAuthOnly
              checked={Boolean(!customChecked && oauthOnlyChecked)}
              onClick={onClickChooseLoginMethod}
            />
            <MethodCustom
              checked={customChecked}
              onClick={onClickChooseLoginMethod}
            />
          </>
        ) : null}
        {isChoosingMethod ? (
          <Widget className={styles.widget}>
            <WidgetTitle>
              <FormattedMessage id="LoginMethodConfigurationScreen.method.title" />
            </WidgetTitle>
            {phoneLoginIDDisabled ? (
              <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
            ) : null}
            <MethodGroup
              title={
                <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.title" />
              }
            >
              <ChoiceEmailPasswordless
                checked={Boolean(!customChecked && emailPasswordlessChecked)}
                onClick={onEmailPasswordlessClick}
              />
              <ChoicePhonePasswordless
                checked={Boolean(!customChecked && phonePasswordlessChecked)}
                disabled={phonePasswordlessDisabled}
                onClick={onPhonePasswordlessClick}
              />
              <ChoicePhoneEmailPasswordless
                checked={Boolean(
                  !customChecked && phoneEmailPasswordlessChecked
                )}
                disabled={phoneEmailPasswordlessDisabled}
                onClick={onPhoneEmailPasswordlessClick}
              />
            </MethodGroup>
            <MethodGroup
              title={
                <FormattedMessage id="LoginMethodConfigurationScreen.method.password.title" />
              }
            >
              <ChoiceEmailPassword
                checked={Boolean(!customChecked && emailPasswordChecked)}
                onClick={onEmailPasswordClick}
              />
              <ChoicePhonePassword
                checked={Boolean(!customChecked && phonePasswordChecked)}
                disabled={phonePasswordDisabled}
                onClick={onPhonePasswordClick}
              />
              <ChoicePhoneEmailPassword
                checked={Boolean(!customChecked && phoneEmailPasswordChecked)}
                disabled={phoneEmailPasswordDisabled}
                onClick={onPhoneEmailPasswordClick}
              />
              <ChoiceUsernamePassword
                checked={Boolean(!customChecked && usernamePasswordChecked)}
                onClick={onUsernamePasswordClick}
              />
            </MethodGroup>
            <MethodGroup
              title={
                <FormattedMessage id="LoginMethodConfigurationScreen.method.other.title" />
              }
            >
              <ChoiceOAuthOnly
                checked={Boolean(!customChecked && oauthOnlyChecked)}
                onClick={onOAuthOnlyClick}
              />
              <ChoiceCustom checked={customChecked} onClick={onCustomClick} />
            </MethodGroup>
          </Widget>
        ) : null}
        {!isChoosingMethod ? (
          <>
            <LinkToOAuth
              appID={appID}
              oauthOnlyChecked={Boolean(!customChecked && oauthOnlyChecked)}
            />
            <CustomLoginMethods
              phoneLoginIDDisabled={phoneLoginIDDisabled}
              customChecked={customChecked}
              primaryAuthenticatorsControl={primaryAuthenticatorsControl}
              loginIDKeyConfigsControl={loginIDKeyConfigsControl}
              onChangeLoginIDChecked={onChangeLoginIDChecked}
              onSwapLoginID={onSwapLoginID}
              onChangePrimaryAuthenticatorChecked={
                onChangePrimaryAuthenticatorChecked
              }
              onSwapPrimaryAuthenticator={onSwapPrimaryAuthenticator}
            />
          </>
        ) : null}
      </ScreenContent>
    );
  };

const LoginMethodConfigurationScreen: React.VFC =
  function LoginMethodConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });

    const featureConfig = useAppFeatureConfigQuery(appID);

    if (form.isLoading || featureConfig.loading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    if (featureConfig.error) {
      return (
        <ShowError
          error={featureConfig.error}
          onRetry={featureConfig.refetch}
        />
      );
    }

    return (
      <FormContainer form={form} errorRules={ERROR_RULES}>
        <LoginMethodConfigurationContent
          appID={appID}
          form={form}
          featureConfig={featureConfig.effectiveFeatureConfig}
        />
      </FormContainer>
    );
  };

export default LoginMethodConfigurationScreen;
