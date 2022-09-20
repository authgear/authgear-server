import React, { ReactNode, useMemo, useCallback } from "react";
import { MessageBar, MessageBarType, Text } from "@fluentui/react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { FormattedMessage } from "@oursky/react-messageformat";
import {
  PortalAPIAppConfig,
  IdentityFeatureConfig,
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
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { makeValidationErrorMatchUnknownKindParseRule } from "../../error/parse";
import styles from "./LoginMethodConfigurationScreen.module.css";

const EXCLUSIVE_PRIMARY_AUTHENTICATOR_TYPES: PrimaryAuthenticatorType[] = [
  "password",
  "oob_otp_email",
  "oob_otp_sms",
];

const ERROR_RULES = [
  makeValidationErrorMatchUnknownKindParseRule(
    "const",
    /\/authentication\/identities/,
    "errors.validation.passkey"
  ),
];

interface FormState {
  identities: IdentityType[];
  primaryAuthenticators: PrimaryAuthenticatorType[];
  loginIDKeyConfigs: LoginIDKeyConfig[];
}

function loginIDIdentity(identities: IdentityType[]): boolean {
  return identities.includes("login_id");
}

function oauthIdentity(identities: IdentityType[]): boolean {
  // We intentionally do not check if "oauth" is present.
  // It is because it could be absent.
  return !loginIDIdentity(identities);
}

function loginIDOf(
  types: LoginIDKeyType[],
  loginIDKeyConfigs: LoginIDKeyConfig[]
): boolean {
  // We want the content and the order both be equal.
  if (types.length !== loginIDKeyConfigs.length) {
    return false;
  }

  for (let i = 0; i < types.length; ++i) {
    const typ = types[i];
    const config = loginIDKeyConfigs[i];
    if (config.type !== typ) {
      return false;
    }
  }

  return true;
}

function primaryAuthenticatorOf(
  types: PrimaryAuthenticatorType[],
  primaryAuthenticators: PrimaryAuthenticatorType[]
): boolean {
  const set1 = new Set(types);
  const set2 = new Set(primaryAuthenticators);

  const set3 = new Set<PrimaryAuthenticatorType>();
  for (const t of EXCLUSIVE_PRIMARY_AUTHENTICATOR_TYPES) {
    if (!set1.has(t)) {
      set3.add(t);
    }
  }

  // We want set2 >= set1 and the intersection of set2 and set3 is empty.

  // set2 >= set1
  for (const e1 of set1) {
    if (!set2.has(e1)) {
      return false;
    }
  }

  for (const e3 of set3) {
    if (set2.has(e3)) {
      return false;
    }
  }

  return true;
}

function setLoginIDIdentity(draft: FormState) {
  if (draft.identities.includes("login_id")) {
    return;
  }

  draft.identities.splice(0, 0, "login_id");
}

function setOAuthIdentity(draft: FormState) {
  const index = draft.identities.findIndex((t) => t === "login_id");
  if (index < 0) {
    return;
  }

  draft.identities.splice(index, 1);
}

function setLoginID(draft: FormState, types: LoginIDKeyType[]) {
  const out: LoginIDKeyConfig[] = [];

  for (const t of types) {
    const c = draft.loginIDKeyConfigs.find((c) => c.type === t);
    if (c != null) {
      out.push(c);
    } else {
      out.push({ type: t });
    }
  }

  draft.loginIDKeyConfigs = out;
}

function setPrimaryAuthenticator(
  draft: FormState,
  types: PrimaryAuthenticatorType[]
) {
  const others = draft.primaryAuthenticators.filter(
    (t) => !EXCLUSIVE_PRIMARY_AUTHENTICATOR_TYPES.includes(t)
  );
  draft.primaryAuthenticators = [...types, ...others];
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    identities: config.authentication?.identities ?? [],
    primaryAuthenticators: config.authentication?.primary_authenticators ?? [],
    loginIDKeyConfigs: config.identity?.login_id?.keys ?? [],
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.authentication ??= {};
    config.identity ??= {};
    config.identity.login_id ??= {};

    config.authentication.identities = currentState.identities;
    config.authentication.primary_authenticators =
      currentState.primaryAuthenticators;
    config.identity.login_id.keys = currentState.loginIDKeyConfigs;

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
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.choice.custom.description" />
      }
    />
  );
}

interface GroupPasswordlessProps {
  emailPasswordlessChecked: boolean;
  onEmailPasswordlessClick: ChoiceProps["onClick"];
  phonePasswordlessChecked: boolean;
  onPhonePasswordlessClick: ChoiceProps["onClick"];
  phoneEmailPasswordlessChecked: boolean;
  onPhoneEmailPasswordlessClick: ChoiceProps["onClick"];
}

function GroupPasswordless(props: GroupPasswordlessProps) {
  const {
    emailPasswordlessChecked,
    onEmailPasswordlessClick,
    phonePasswordlessChecked,
    onPhonePasswordlessClick,
    phoneEmailPasswordlessChecked,
    onPhoneEmailPasswordlessClick,
  } = props;
  return (
    <MethodGroup
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.title" />
      }
    >
      <ChoiceEmailPasswordless
        checked={emailPasswordlessChecked}
        onClick={onEmailPasswordlessClick}
      />
      <ChoicePhonePasswordless
        checked={phonePasswordlessChecked}
        onClick={onPhonePasswordlessClick}
      />
      <ChoicePhoneEmailPasswordless
        checked={phoneEmailPasswordlessChecked}
        onClick={onPhoneEmailPasswordlessClick}
      />
    </MethodGroup>
  );
}

interface GroupPasswordProps {
  emailPasswordChecked: boolean;
  onEmailPasswordClick: ChoiceProps["onClick"];
  phonePasswordChecked: boolean;
  onPhonePasswordClick: ChoiceProps["onClick"];
  phoneEmailPasswordChecked: boolean;
  onPhoneEmailPasswordClick: ChoiceProps["onClick"];
  usernamePasswordChecked: boolean;
  onUsernamePasswordClick: ChoiceProps["onClick"];
}

function GroupPassword(props: GroupPasswordProps) {
  const {
    emailPasswordChecked,
    onEmailPasswordClick,
    phonePasswordChecked,
    onPhonePasswordClick,
    phoneEmailPasswordChecked,
    onPhoneEmailPasswordClick,
    usernamePasswordChecked,
    onUsernamePasswordClick,
  } = props;
  return (
    <MethodGroup
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.title" />
      }
    >
      <ChoiceEmailPassword
        checked={emailPasswordChecked}
        onClick={onEmailPasswordClick}
      />
      <ChoicePhonePassword
        checked={phonePasswordChecked}
        onClick={onPhonePasswordClick}
      />
      <ChoicePhoneEmailPassword
        checked={phoneEmailPasswordChecked}
        onClick={onPhoneEmailPasswordClick}
      />
      <ChoiceUsernamePassword
        checked={usernamePasswordChecked}
        onClick={onUsernamePasswordClick}
      />
    </MethodGroup>
  );
}

interface GroupOtherProps {
  oauthOnlyChecked: boolean;
  onOAuthOnlyClick: ChoiceProps["onClick"];
  customChecked: boolean;
}

function GroupOther(props: GroupOtherProps) {
  const { oauthOnlyChecked, onOAuthOnlyClick, customChecked } = props;
  return (
    <MethodGroup
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.title" />
      }
    >
      <ChoiceOAuthOnly checked={oauthOnlyChecked} onClick={onOAuthOnlyClick} />
      <ChoiceCustom checked={customChecked} />
    </MethodGroup>
  );
}

interface LinkToPasskeyProps {
  appID: string;
}

function LinkToPasskey(props: LinkToPasskeyProps) {
  const { appID } = props;
  return (
    <MessageBar messageBarType={MessageBarType.info}>
      <FormattedMessage
        id="LoginMethodConfigurationScreen.passkey"
        values={{
          to: `/project/${appID}/configuration/authentication/passkey`,
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

interface LoginMethodConfigurationContentProps {
  appID: string;
  form: AppConfigFormModel<FormState>;
  identityFeatureConfig?: IdentityFeatureConfig;
}

const LoginMethodConfigurationContent: React.VFC<LoginMethodConfigurationContentProps> =
  function LoginMethodConfigurationContent(props) {
    const { appID } = props;
    const { state, setState } = props.form;

    const { identities, loginIDKeyConfigs, primaryAuthenticators } = state;

    const emailPasswordlessChecked = useMemo(() => {
      return (
        loginIDIdentity(identities) &&
        loginIDOf(["email"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["oob_otp_email"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const onEmailPasswordlessClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
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
        loginIDIdentity(identities) &&
        loginIDOf(["phone"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["oob_otp_sms"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const onPhonePasswordlessClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
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
        loginIDIdentity(identities) &&
        // Order is important.
        loginIDOf(["phone", "email"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(
          ["oob_otp_sms", "oob_otp_email"],
          primaryAuthenticators
        )
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const onPhoneEmailPasswordlessClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
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
        loginIDIdentity(identities) &&
        loginIDOf(["email"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const onEmailPasswordClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
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
        loginIDIdentity(identities) &&
        loginIDOf(["phone"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const onPhonePasswordClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
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
        loginIDIdentity(identities) &&
        // Order is important.
        loginIDOf(["phone", "email"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const onPhoneEmailPasswordClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
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
        loginIDIdentity(identities) &&
        loginIDOf(["username"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const onUsernamePasswordClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
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
        oauthIdentity(identities) &&
        loginIDOf([], loginIDKeyConfigs) &&
        primaryAuthenticatorOf([], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const onOAuthOnlyClick = useCallback(
      (e) => {
        e.preventDefault();
        e.stopPropagation();
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

    const customChecked = useMemo(() => {
      return (
        !emailPasswordlessChecked &&
        !phonePasswordlessChecked &&
        !phoneEmailPasswordlessChecked &&
        !emailPasswordChecked &&
        !phonePasswordChecked &&
        !phoneEmailPasswordChecked &&
        !usernamePasswordChecked &&
        !oauthOnlyChecked
      );
    }, [
      emailPasswordlessChecked,
      phonePasswordlessChecked,
      phoneEmailPasswordlessChecked,
      emailPasswordChecked,
      phonePasswordChecked,
      phoneEmailPasswordChecked,
      usernamePasswordChecked,
      oauthOnlyChecked,
    ]);

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="LoginMethodConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="LoginMethodConfigurationScreen.description" />
        </ScreenDescription>
        <Widget className={styles.widget}>
          <WidgetTitle>
            <FormattedMessage id="LoginMethodConfigurationScreen.method.title" />
          </WidgetTitle>
          <GroupPasswordless
            emailPasswordlessChecked={emailPasswordlessChecked}
            onEmailPasswordlessClick={onEmailPasswordlessClick}
            phonePasswordlessChecked={phonePasswordlessChecked}
            onPhonePasswordlessClick={onPhonePasswordlessClick}
            phoneEmailPasswordlessChecked={phoneEmailPasswordlessChecked}
            onPhoneEmailPasswordlessClick={onPhoneEmailPasswordlessClick}
          />
          <LinkToPasskey appID={appID} />
          <GroupPassword
            emailPasswordChecked={emailPasswordChecked}
            onEmailPasswordClick={onEmailPasswordClick}
            phonePasswordChecked={phonePasswordChecked}
            onPhonePasswordClick={onPhonePasswordClick}
            phoneEmailPasswordChecked={phoneEmailPasswordChecked}
            onPhoneEmailPasswordClick={onPhoneEmailPasswordClick}
            usernamePasswordChecked={usernamePasswordChecked}
            onUsernamePasswordClick={onUsernamePasswordClick}
          />
          <GroupOther
            oauthOnlyChecked={oauthOnlyChecked}
            onOAuthOnlyClick={onOAuthOnlyClick}
            customChecked={customChecked}
          />
        </Widget>
        <LinkToOAuth appID={appID} oauthOnlyChecked={oauthOnlyChecked} />
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
          identityFeatureConfig={featureConfig.effectiveFeatureConfig?.identity}
        />
      </FormContainer>
    );
  };

export default LoginMethodConfigurationScreen;
