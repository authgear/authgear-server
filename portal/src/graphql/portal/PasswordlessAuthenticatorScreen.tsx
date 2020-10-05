import React, { useCallback, useMemo, useState } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Label, Text } from "@fluentui/react";
import { useParams } from "react-router-dom";
import cn from "classnames";
import produce from "immer";
import deepEqual from "deep-equal";

import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { useAppConfigQuery } from "./query/appConfigQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import CodeEditor from "../../CodeEditor";
import { PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";

import styles from "./PasswordlessAuthenticatorScreen.module.scss";

interface PasswordlessAuthenticatorScreenState {
  emailHtmlTemplate: string;
  emailPlainTextTemplate: string;
  smsTemplate: string;
}

function constructStateFromAppConfig(
  /* eslint-disable-next-line @typescript-eslint/no-unused-vars */
  _appConfig: PortalAPIAppConfig | null
): PasswordlessAuthenticatorScreenState {
  return {
    emailHtmlTemplate: "", // TODO: handle email template
    emailPlainTextTemplate: "", // TODO: handle email template
    smsTemplate: "", // TODO: handle sms template
  };
}

function constructAppConfigFromState(
  rawAppConfig: PortalAPIAppConfig,
  /* eslint-disable-next-line @typescript-eslint/no-unused-vars */
  _initialScreenState: PasswordlessAuthenticatorScreenState,
  /* eslint-disable-next-line @typescript-eslint/no-unused-vars */
  _screenState: PasswordlessAuthenticatorScreenState
): PortalAPIAppConfig {
  const newAppConfig = produce(rawAppConfig, (draftConfig) => {
    // TODO: update email template
    // TODO: update sms template

    clearEmptyObject(draftConfig);
  });

  return newAppConfig;
}

const PasswordlessAuthenticatorScreen: React.FC = function PasswordlessAuthenticatorScreen() {
  const { appID } = useParams();

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);
  const { loading, error, data, refetch } = useAppConfigQuery(appID);
  const { effectiveAppConfig, rawAppConfig } = useMemo(() => {
    const node = data?.node;
    return node?.__typename === "App"
      ? {
          effectiveAppConfig: node.effectiveAppConfig,
          rawAppConfig: node.rawAppConfig,
        }
      : {
          effectiveAppConfig: null,
          rawAppConfig: null,
        };
  }, [data]);

  const initialState = useMemo(() => {
    return constructStateFromAppConfig(effectiveAppConfig);
  }, [effectiveAppConfig]);

  const [state, setState] = useState(initialState);

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

  const onEmailHtmlTemplateChange = useCallback(
    (_event: unknown, value: string | undefined) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        emailHtmlTemplate: value,
      }));
    },
    []
  );

  const onEmailPlainTextTemplateChange = useCallback(
    (_event: unknown, value: string | undefined) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        emailPlainTextTemplate: value,
      }));
    },
    []
  );

  const onSaveButtonClicked = useCallback(() => {
    if (rawAppConfig == null) {
      return;
    }

    const newAppConfig = constructAppConfigFromState(
      rawAppConfig,
      initialState,
      state
    );

    // TODO: handle error
    updateAppConfig(newAppConfig)
      .then(() => {
        // TODO: remove this alert after implementing templates saving
        alert("SMS and email templates cannot be saved currently");
      })
      .catch(() => {});
  }, [state, rawAppConfig, updateAppConfig, initialState]);

  const onSmsTemplateChange = useCallback((_event, value?: string) => {
    if (value === undefined) {
      return;
    }
    setState((state) => ({
      ...state,
      smsTemplate: value,
    }));
  }, []);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={cn(styles.root, { [styles.loading]: updatingAppConfig })}>
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      <div className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="PasswordlessAuthenticatorScreen.title" />
        </Text>
        <div className={styles.form}>
          <Label className={styles.boldLabel}>
            <FormattedMessage id="PasswordsScreen.forgot-password.email.label" />
          </Label>

          <Label className={styles.label}>
            <FormattedMessage id="PasswordlessAuthenticatorScreen.email.styled-content.label" />
          </Label>
          <CodeEditor
            className={styles.htmlCodeEditor}
            language="html"
            value={state.emailHtmlTemplate}
            onChange={onEmailHtmlTemplateChange}
          />

          <Label className={styles.label}>
            <FormattedMessage id="PasswordlessAuthenticatorScreen.email.plain-content.label" />
          </Label>
          <CodeEditor
            className={styles.plainTextCodeEditor}
            language="plaintext"
            value={state.emailPlainTextTemplate}
            onChange={onEmailPlainTextTemplateChange}
          />

          <Label className={styles.boldLabel}>
            <FormattedMessage id="PasswordlessAuthenticatorScreen.sms.label" />
          </Label>

          <Label className={styles.label}>
            <FormattedMessage id="PasswordlessAuthenticatorScreen.sms.content.label" />
          </Label>
          <CodeEditor
            className={styles.plainTextCodeEditor}
            language="plaintext"
            value={state.smsTemplate}
            onChange={onSmsTemplateChange}
          />

          <div className={styles.saveButtonContainer}>
            <ButtonWithLoading
              disabled={!isFormModified}
              onClick={onSaveButtonClicked}
              loading={updatingAppConfig}
              labelId="save"
              loadingLabelId="saving"
            />
          </div>
        </div>
      </div>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
    </main>
  );
};

export default PasswordlessAuthenticatorScreen;
