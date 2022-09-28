import React, {
  ReactNode,
  useMemo,
  useCallback,
  useState,
  useContext,
} from "react";
import cn from "classnames";
import {
  MessageBar,
  MessageBarType,
  Text,
  useTheme,
  IButtonProps,
  FontIcon,
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
import ChoiceButton from "../../ChoiceButton";
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

type LoginMethodPasswordlessLoginID = "email" | "phone" | "phone-email";
type LoginMethodPasswordLoginID =
  | "email"
  | "phone"
  | "phone-email"
  | "username";

type LoginMethod =
  | `passwordless-${LoginMethodPasswordlessLoginID}`
  | `password-${LoginMethodPasswordLoginID}`
  | "oauth"
  | "custom";

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
function loginMethodFromFormState(formState: FormState): LoginMethod {
  const {
    identitiesControl,
    loginIDKeyConfigsControl,
    primaryAuthenticatorsControl,
  } = formState;

  if (
    loginIDIdentity(identitiesControl) &&
    loginIDOf(["email"], loginIDKeyConfigsControl) &&
    primaryAuthenticatorOf(["oob_otp_email"], primaryAuthenticatorsControl)
  ) {
    return "passwordless-email";
  }

  if (
    loginIDIdentity(identitiesControl) &&
    loginIDOf(["phone"], loginIDKeyConfigsControl) &&
    primaryAuthenticatorOf(["oob_otp_sms"], primaryAuthenticatorsControl)
  ) {
    return "passwordless-phone";
  }

  if (
    loginIDIdentity(identitiesControl) &&
    // Order is important.
    loginIDOf(["phone", "email"], loginIDKeyConfigsControl) &&
    primaryAuthenticatorOf(
      ["oob_otp_sms", "oob_otp_email"],
      primaryAuthenticatorsControl
    )
  ) {
    return "passwordless-phone-email";
  }

  if (
    loginIDIdentity(identitiesControl) &&
    loginIDOf(["email"], loginIDKeyConfigsControl) &&
    primaryAuthenticatorOf(["password"], primaryAuthenticatorsControl)
  ) {
    return "password-email";
  }

  if (
    loginIDIdentity(identitiesControl) &&
    loginIDOf(["phone"], loginIDKeyConfigsControl) &&
    primaryAuthenticatorOf(["password"], primaryAuthenticatorsControl)
  ) {
    return "password-phone";
  }

  if (
    loginIDIdentity(identitiesControl) &&
    // Order is important.
    loginIDOf(["phone", "email"], loginIDKeyConfigsControl) &&
    primaryAuthenticatorOf(["password"], primaryAuthenticatorsControl)
  ) {
    return "password-phone-email";
  }

  if (
    loginIDIdentity(identitiesControl) &&
    loginIDOf(["username"], loginIDKeyConfigsControl) &&
    primaryAuthenticatorOf(["password"], primaryAuthenticatorsControl)
  ) {
    return "password-username";
  }

  if (
    oauthIdentity(identitiesControl) &&
    loginIDOf([], loginIDKeyConfigsControl) &&
    primaryAuthenticatorOf([], primaryAuthenticatorsControl)
  ) {
    return "oauth";
  }

  return "custom";
}

function setLoginMethodToFormState(
  formState: FormState,
  loginMethod: LoginMethod
) {
  switch (loginMethod) {
    case "passwordless-email":
      setLoginIDIdentity(formState);
      setLoginID(formState, ["email"]);
      setPrimaryAuthenticator(formState, ["oob_otp_email"]);
      break;
    case "passwordless-phone":
      setLoginIDIdentity(formState);
      setLoginID(formState, ["phone"]);
      setPrimaryAuthenticator(formState, ["oob_otp_sms"]);
      break;
    case "passwordless-phone-email":
      setLoginIDIdentity(formState);
      setLoginID(formState, ["phone", "email"]);
      setPrimaryAuthenticator(formState, ["oob_otp_sms", "oob_otp_email"]);
      break;
    case "password-email":
      setLoginIDIdentity(formState);
      setLoginID(formState, ["email"]);
      setPrimaryAuthenticator(formState, ["password"]);
      break;
    case "password-phone":
      setLoginIDIdentity(formState);
      setLoginID(formState, ["phone"]);
      setPrimaryAuthenticator(formState, ["password"]);
      break;
    case "password-phone-email":
      setLoginIDIdentity(formState);
      setLoginID(formState, ["phone", "email"]);
      setPrimaryAuthenticator(formState, ["password"]);
      break;
    case "password-username":
      setLoginIDIdentity(formState);
      setLoginID(formState, ["username"]);
      setPrimaryAuthenticator(formState, ["password"]);
      break;
    case "oauth":
      setOAuthIdentity(formState);
      setLoginID(formState, []);
      setPrimaryAuthenticator(formState, []);
      break;
    case "custom":
      // No changes.
      break;
  }
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

interface WidgetSubtitleProps {
  children?: ReactNode;
}

const FIELD_TITLE_STYLES = {
  root: {
    fontWeight: "600",
  },
};

function WidgetSubtitle(props: WidgetSubtitleProps) {
  const { children } = props;
  return (
    <Text as="h3" block={true} variant="medium" styles={FIELD_TITLE_STYLES}>
      {children}
    </Text>
  );
}

interface WidgetSubsectionProps {
  children?: ReactNode;
}

function WidgetSubsection(props: WidgetSubsectionProps) {
  const { children } = props;
  return <div className={styles.widgetSubsection}>{children}</div>;
}

interface MethodButtonProps {
  text?: ReactNode;
  secondaryText?: ReactNode;
  onClick?: IButtonProps["onClick"];
}

function MethodButton(props: MethodButtonProps) {
  const { text, secondaryText, onClick } = props;
  const theme = useTheme();
  const { themes } = useSystemConfig();

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
          <FormattedMessage id="LoginMethodConfigurationScreen.change-method" />
        }
        onClick={onClick}
      />
    </div>
  );
}

interface ChosenMethodProps {
  loginMethod: LoginMethod;
  onClick: MethodButtonProps["onClick"];
}

function ChosenMethod(props: ChosenMethodProps) {
  const { loginMethod, onClick } = props;
  const title = `LoginMethodConfigurationScreen.login-method.title.${loginMethod}`;
  const description =
    loginMethod === "custom"
      ? undefined
      : `LoginMethodConfigurationScreen.login-method.description.${loginMethod}`;
  return (
    <MethodButton
      text={<FormattedMessage id={title} />}
      secondaryText={
        description == null ? undefined : <FormattedMessage id={description} />
      }
      onClick={onClick}
    />
  );
}

function MatrixAdd() {
  return <FontIcon className={styles.matrixAdd} iconName="Add" />;
}

function MatrixOr() {
  return (
    <div className={styles.matrixOrContainer}>
      <HorizontalDivider className={styles.matrixOrDivider} />
      <Text variant="medium">
        <FormattedMessage id="LoginMethodConfigurationScreen.matrix.or" />
      </Text>
      <HorizontalDivider className={styles.matrixOrDivider} />
    </div>
  );
}

function MatrixColumnNoChoice() {
  const theme = useTheme();
  return (
    <div className={styles.matrixColumnNoChoice}>
      <Text
        block={true}
        variant="medium"
        styles={{
          root: {
            color: theme.semanticColors.disabledBodyText,
            fontWeight: "600",
          },
        }}
      >
        <FormattedMessage id="LoginMethodConfigurationScreen.matrix.no-login-id-choices" />
      </Text>
    </div>
  );
}

function MatrixColumnInapplicableChoice() {
  const theme = useTheme();
  return (
    <div
      className={styles.matrixColumnNoChoice}
      style={{
        backgroundColor: theme.semanticColors.infoBackground,
      }}
    >
      <Text
        block={true}
        variant="medium"
        styles={{
          root: {
            color: theme.semanticColors.disabledBodyText,
            fontWeight: "600",
          },
        }}
      >
        <FormattedMessage id="LoginMethodConfigurationScreen.matrix.inapplicable-login-id-choices" />
      </Text>
    </div>
  );
}

interface MatrixColumnsProps {
  children?: ReactNode;
}

function MatrixColumns(props: MatrixColumnsProps) {
  const { children } = props;
  return <div className={styles.matrixColumns}>{children}</div>;
}

interface MatrixColumnProps {
  children?: ReactNode;
}

function MatrixColumn(props: MatrixColumnProps) {
  const { children } = props;
  return <div className={styles.matrixColumn}>{children}</div>;
}

interface MatrixFirstLevelChoiceProps {
  checked: boolean;
  loginMethod: LoginMethod;
  text: ReactNode;
  secondaryText: ReactNode;
  onClick?: (loginMethod: LoginMethod) => void;
}

function MatrixFirstLevelChoice(props: MatrixFirstLevelChoiceProps) {
  const {
    checked,
    loginMethod,
    onClick: onClickProp,
    text,
    secondaryText,
  } = props;
  const onClick = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      if (!checked) {
        onClickProp?.(loginMethod);
      }
    },
    [checked, onClickProp, loginMethod]
  );
  return (
    <ChoiceButton
      className={styles.matrixChoice}
      checked={checked}
      text={text}
      secondaryText={secondaryText}
      onClick={onClick}
    />
  );
}

interface MatrixSecondLevelChoiceProps {
  className?: string;
  currentValue: LoginMethod;
  choiceValue: LoginMethod;
  disabled?: boolean;
  onClick?: (loginMethod: LoginMethod) => void;
}

function MatrixSecondLevelChoice(props: MatrixSecondLevelChoiceProps) {
  const {
    className,
    currentValue,
    choiceValue,
    disabled,
    onClick: onClickProp,
  } = props;
  const checked = currentValue === choiceValue;
  const onClick = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      if (currentValue !== choiceValue) {
        onClickProp?.(choiceValue);
      }
    },
    [currentValue, choiceValue, onClickProp]
  );
  return (
    <ChoiceButton
      className={className}
      checked={checked}
      disabled={disabled}
      text={
        <FormattedMessage
          id={
            "LoginMethodConfigurationScreen.matrix.login-method.title." +
            choiceValue
          }
        />
      }
      secondaryText={
        choiceValue === "custom" ? undefined : (
          <FormattedMessage
            id={
              "LoginMethodConfigurationScreen.login-method.description." +
              choiceValue
            }
          />
        )
      }
      onClick={onClick}
    />
  );
}

interface MatrixProps {
  loginMethod: LoginMethod;
  phoneLoginIDDisabled: boolean;
  onChangeLoginMethod: (loginMethod: LoginMethod) => void;
}

function Matrix(props: MatrixProps) {
  const { phoneLoginIDDisabled, onChangeLoginMethod } = props;
  const [loginMethod, setLoginMethod] = useState(props.loginMethod);
  const passwordlessChecked = useMemo(
    () => loginMethod.startsWith("passwordless-"),
    [loginMethod]
  );
  const passwordChecked = useMemo(
    () => loginMethod.startsWith("password-"),
    [loginMethod]
  );
  const oauthChecked = loginMethod === "oauth";
  const onClick = useCallback((loginMethod) => {
    setLoginMethod(loginMethod);
  }, []);
  const onClickConfirm = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      onChangeLoginMethod(loginMethod);
    },
    [loginMethod, onChangeLoginMethod]
  );
  return (
    <div className={styles.matrix}>
      <WidgetDescription>
        <FormattedMessage id="LoginMethodConfigurationScreen.matrix.description" />
      </WidgetDescription>
      <MatrixColumns>
        <MatrixColumn>
          <MatrixFirstLevelChoice
            checked={passwordlessChecked}
            loginMethod="passwordless-email"
            text={
              <FormattedMessage id="LoginMethodConfigurationScreen.matrix.authenticator.title.passwordless" />
            }
            secondaryText={
              <FormattedMessage id="LoginMethodConfigurationScreen.matrix.authenticator.description.passwordless" />
            }
            onClick={onClick}
          />
          <MatrixFirstLevelChoice
            checked={passwordChecked}
            loginMethod="password-email"
            text={
              <FormattedMessage id="LoginMethodConfigurationScreen.matrix.authenticator.title.password" />
            }
            secondaryText={
              <FormattedMessage id="LoginMethodConfigurationScreen.matrix.authenticator.description.password" />
            }
            onClick={onClick}
          />
          <MatrixSecondLevelChoice
            className={styles.matrixChoice}
            currentValue={loginMethod}
            choiceValue="oauth"
            onClick={onClick}
          />
        </MatrixColumn>
        <MatrixAdd />
        {passwordlessChecked ? (
          <MatrixColumn>
            <MatrixSecondLevelChoice
              className={styles.matrixChoice}
              currentValue={loginMethod}
              choiceValue={"passwordless-email"}
              onClick={onClick}
            />
            <MatrixSecondLevelChoice
              className={styles.matrixChoice}
              currentValue={loginMethod}
              choiceValue={"passwordless-phone"}
              disabled={phoneLoginIDDisabled}
              onClick={onClick}
            />
            <MatrixSecondLevelChoice
              className={styles.matrixChoice}
              currentValue={loginMethod}
              choiceValue={"passwordless-phone-email"}
              disabled={phoneLoginIDDisabled}
              onClick={onClick}
            />
          </MatrixColumn>
        ) : passwordChecked ? (
          <MatrixColumn>
            <MatrixSecondLevelChoice
              className={styles.matrixChoice}
              currentValue={loginMethod}
              choiceValue={"password-email"}
              onClick={onClick}
            />
            <MatrixSecondLevelChoice
              className={styles.matrixChoice}
              currentValue={loginMethod}
              choiceValue={"password-phone"}
              disabled={phoneLoginIDDisabled}
              onClick={onClick}
            />
            <MatrixSecondLevelChoice
              className={styles.matrixChoice}
              currentValue={loginMethod}
              choiceValue={"password-phone-email"}
              disabled={phoneLoginIDDisabled}
              onClick={onClick}
            />
            <MatrixSecondLevelChoice
              className={styles.matrixChoice}
              currentValue={loginMethod}
              choiceValue={"password-username"}
              onClick={onClick}
            />
          </MatrixColumn>
        ) : oauthChecked ? (
          <MatrixColumnNoChoice />
        ) : (
          <MatrixColumnInapplicableChoice />
        )}
      </MatrixColumns>
      <MatrixOr />
      <MatrixSecondLevelChoice
        currentValue={loginMethod}
        choiceValue="custom"
        onClick={onClick}
      />
      <PrimaryButton
        className={styles.matrixConfirmButton}
        text={<FormattedMessage id="next" />}
        onClick={onClickConfirm}
      />
    </div>
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
}

function LinkToOAuth(props: LinkToOAuthProps) {
  const { appID } = props;

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
      <WidgetSubsection>
        <WidgetSubtitle>
          <FormattedMessage id="LoginMethodConfigurationScreen.custom-login-methods.login-id.title" />
        </WidgetSubtitle>
        <WidgetDescription>
          <FormattedMessage id="LoginMethodConfigurationScreen.custom-login-methods.login-id.description" />
        </WidgetDescription>
      </WidgetSubsection>
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
      <WidgetSubsection>
        <WidgetSubtitle>
          <FormattedMessage id="LoginMethodConfigurationScreen.custom-login-methods.authenticator.title" />
        </WidgetSubtitle>
        <WidgetDescription>
          <FormattedMessage id="LoginMethodConfigurationScreen.custom-login-methods.authenticator.description" />
        </WidgetDescription>
      </WidgetSubsection>
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

    const { primaryAuthenticatorsControl, loginIDKeyConfigsControl } = state;

    const [loginMethod, setLoginMethod] = useState(() =>
      loginMethodFromFormState(state)
    );

    const [isChoosingMethod, setIsChoosingMethod] = useState(false);

    const phoneLoginIDDisabled =
      featureConfig?.identity?.login_id?.types?.phone?.disabled ?? false;

    const onClickChooseLoginMethod = useCallback((e) => {
      e.preventDefault();
      e.stopPropagation();
      setIsChoosingMethod(true);
    }, []);

    const onChangeLoginMethod = useCallback(
      (loginMethod: LoginMethod) => {
        setIsChoosingMethod(false);
        setLoginMethod(loginMethod);
        setState((prev) =>
          produce(prev, (prev) => {
            setLoginMethodToFormState(prev, loginMethod);
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
            <ChosenMethod
              loginMethod={loginMethod}
              onClick={onClickChooseLoginMethod}
            />
            {loginMethod === "oauth" ? <LinkToOAuth appID={appID} /> : null}
            {loginMethod === "custom" ? (
              <CustomLoginMethods
                phoneLoginIDDisabled={phoneLoginIDDisabled}
                primaryAuthenticatorsControl={primaryAuthenticatorsControl}
                loginIDKeyConfigsControl={loginIDKeyConfigsControl}
                onChangeLoginIDChecked={onChangeLoginIDChecked}
                onSwapLoginID={onSwapLoginID}
                onChangePrimaryAuthenticatorChecked={
                  onChangePrimaryAuthenticatorChecked
                }
                onSwapPrimaryAuthenticator={onSwapPrimaryAuthenticator}
              />
            ) : null}
          </>
        ) : (
          <Widget className={styles.widget}>
            {phoneLoginIDDisabled ? (
              <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
            ) : null}
            <Matrix
              loginMethod={loginMethod}
              onChangeLoginMethod={onChangeLoginMethod}
              phoneLoginIDDisabled={phoneLoginIDDisabled}
            />
          </Widget>
        )}
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
