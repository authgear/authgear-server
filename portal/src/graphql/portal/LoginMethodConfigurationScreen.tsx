import React, { ReactNode } from "react";
import { MessageBar, MessageBarType, Text } from "@fluentui/react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { FormattedMessage } from "@oursky/react-messageformat";
import { PortalAPIAppConfig, IdentityFeatureConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import ChoiceButton from "../../ChoiceButton";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import styles from "./LoginMethodConfigurationScreen.module.css";

interface FormState {}

function constructFormState(_config: PortalAPIAppConfig): FormState {
  return {};
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

function ChoiceEmailPasswordless() {
  return (
    <ChoiceButton
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.email.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.email.description" />
      }
    />
  );
}

function ChoicePhonePasswordless() {
  return (
    <ChoiceButton
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.phone.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.phone.description" />
      }
    />
  );
}

function ChoicePhoneEmailPasswordless() {
  return (
    <ChoiceButton
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.all.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.choice.all.description" />
      }
    />
  );
}

function ChoiceEmailPassword() {
  return (
    <ChoiceButton
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.email.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.email.description" />
      }
    />
  );
}

function ChoicePhonePassword() {
  return (
    <ChoiceButton
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.phone.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.phone.description" />
      }
    />
  );
}

function ChoiceNoUsernamePassword() {
  return (
    <ChoiceButton
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.no-username.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.no-username.description" />
      }
    />
  );
}

function ChoiceUsernamePassword() {
  return (
    <ChoiceButton
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.username.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.choice.username.description" />
      }
    />
  );
}

function ChoiceOAuthOnly() {
  return (
    <ChoiceButton
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.choice.oauth.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.choice.oauth.description" />
      }
    />
  );
}

function ChoiceCustom() {
  return (
    <ChoiceButton
      text={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.choice.custom.title" />
      }
      secondaryText={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.choice.custom.description" />
      }
    />
  );
}

function GroupPasswordless() {
  return (
    <MethodGroup
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.passwordless.title" />
      }
    >
      <ChoiceEmailPasswordless />
      <ChoicePhonePasswordless />
      <ChoicePhoneEmailPasswordless />
    </MethodGroup>
  );
}

function GroupPassword() {
  return (
    <MethodGroup
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.password.title" />
      }
    >
      <ChoiceEmailPassword />
      <ChoicePhonePassword />
      <ChoiceNoUsernamePassword />
      <ChoiceUsernamePassword />
    </MethodGroup>
  );
}

function GroupOther() {
  return (
    <MethodGroup
      title={
        <FormattedMessage id="LoginMethodConfigurationScreen.method.other.title" />
      }
    >
      <ChoiceOAuthOnly />
      <ChoiceCustom />
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
          <GroupPasswordless />
          <LinkToPasskey appID={appID} />
          <GroupPassword />
          <GroupOther />
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
