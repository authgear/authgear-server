import React, { ReactNode, useMemo } from "react";
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
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import styles from "./LoginMethodConfigurationScreen.module.css";

const EXCLUSIVE_PRIMARY_AUTHENTICATOR_TYPES: PrimaryAuthenticatorType[] = [
  "password",
  "oob_otp_email",
  "oob_otp_sms",
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
  _currentState: FormState,
  _effectiveConfig: PortalAPIAppConfig
): PortalAPIAppConfig {
  return produce(config, (config) => {
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
  phonePasswordlessChecked: boolean;
  phoneEmailPasswordlessChecked: boolean;
}

function GroupPasswordless(props: GroupPasswordlessProps) {
  const {
    emailPasswordlessChecked,
    phonePasswordlessChecked,
    phoneEmailPasswordlessChecked,
  } = props;
  return (
    <MethodGroup
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.title" />
      }
    >
      <ChoiceEmailPasswordless checked={emailPasswordlessChecked} />
      <ChoicePhonePasswordless checked={phonePasswordlessChecked} />
      <ChoicePhoneEmailPasswordless checked={phoneEmailPasswordlessChecked} />
    </MethodGroup>
  );
}

interface GroupPasswordProps {
  emailPasswordChecked: boolean;
  phonePasswordChecked: boolean;
  phoneEmailPasswordChecked: boolean;
  usernamePasswordChecked: boolean;
}

function GroupPassword(props: GroupPasswordProps) {
  const {
    emailPasswordChecked,
    phonePasswordChecked,
    phoneEmailPasswordChecked,
    usernamePasswordChecked,
  } = props;
  return (
    <MethodGroup
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.title" />
      }
    >
      <ChoiceEmailPassword checked={emailPasswordChecked} />
      <ChoicePhonePassword checked={phonePasswordChecked} />
      <ChoicePhoneEmailPassword checked={phoneEmailPasswordChecked} />
      <ChoiceUsernamePassword checked={usernamePasswordChecked} />
    </MethodGroup>
  );
}

interface GroupOtherProps {
  oauthOnlyChecked: boolean;
  customChecked: boolean;
}

function GroupOther(props: GroupOtherProps) {
  const { oauthOnlyChecked, customChecked } = props;
  return (
    <MethodGroup
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.title" />
      }
    >
      <ChoiceOAuthOnly checked={oauthOnlyChecked} />
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

interface LoginMethodConfigurationContentProps {
  appID: string;
  form: AppConfigFormModel<FormState>;
  identityFeatureConfig?: IdentityFeatureConfig;
}

const LoginMethodConfigurationContent: React.VFC<LoginMethodConfigurationContentProps> =
  function LoginMethodConfigurationContent(props) {
    const { appID } = props;
    const { state } = props.form;

    const { identities, loginIDKeyConfigs, primaryAuthenticators } = state;

    const emailPasswordlessChecked = useMemo(() => {
      return (
        loginIDIdentity(identities) &&
        loginIDOf(["email"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["oob_otp_email"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const phonePasswordlessChecked = useMemo(() => {
      return (
        loginIDIdentity(identities) &&
        loginIDOf(["phone"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["oob_otp_sms"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const phoneEmailPasswordlessChecked = useMemo(() => {
      return (
        loginIDIdentity(identities) &&
        // Order is important.
        loginIDOf(["phone", "email"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(
          ["oob_otp_email", "oob_otp_sms"],
          primaryAuthenticators
        )
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const emailPasswordChecked = useMemo(() => {
      return (
        loginIDIdentity(identities) &&
        loginIDOf(["email"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const phonePasswordChecked = useMemo(() => {
      return (
        loginIDIdentity(identities) &&
        loginIDOf(["phone"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const phoneEmailPasswordChecked = useMemo(() => {
      return (
        loginIDIdentity(identities) &&
        // Order is important.
        loginIDOf(["phone", "email"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const usernamePasswordChecked = useMemo(() => {
      return (
        loginIDIdentity(identities) &&
        loginIDOf(["username"], loginIDKeyConfigs) &&
        primaryAuthenticatorOf(["password"], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

    const oauthOnlyChecked = useMemo(() => {
      return (
        oauthIdentity(identities) &&
        loginIDOf([], loginIDKeyConfigs) &&
        primaryAuthenticatorOf([], primaryAuthenticators)
      );
    }, [identities, loginIDKeyConfigs, primaryAuthenticators]);

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
            phonePasswordlessChecked={phonePasswordlessChecked}
            phoneEmailPasswordlessChecked={phoneEmailPasswordlessChecked}
          />
          <LinkToPasskey appID={appID} />
          <GroupPassword
            emailPasswordChecked={emailPasswordChecked}
            phonePasswordChecked={phonePasswordChecked}
            phoneEmailPasswordChecked={phoneEmailPasswordChecked}
            usernamePasswordChecked={usernamePasswordChecked}
          />
          <GroupOther
            oauthOnlyChecked={oauthOnlyChecked}
            customChecked={customChecked}
          />
        </Widget>
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
      <FormContainer form={form}>
        <LoginMethodConfigurationContent
          appID={appID}
          form={form}
          identityFeatureConfig={featureConfig.effectiveFeatureConfig?.identity}
        />
      </FormContainer>
    );
  };

export default LoginMethodConfigurationScreen;
