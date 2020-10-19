import React, { useCallback, useMemo, useState } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Label, Text } from "@fluentui/react";
import { useParams } from "react-router-dom";
import cn from "classnames";
import produce from "immer";
import deepEqual from "deep-equal";

import {
  AppAndEmailSmsTemplatesConfigUpdater,
  useUpdateAppAndEmailSmsTemplatesConfigMutation,
} from "./mutations/updateAppAndEmailSmsTemplatesMutation";
import { useAppAndEmailSmsTemplatesQuery } from "./query/appAndEmailSmsTemplatesQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import CodeEditor from "../../CodeEditor";
import { clearEmptyObject } from "../../util/misc";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";
import { PortalAPIAppConfig, PortalAPIEmailAndSmsTemplates } from "../../types";

import styles from "./PasswordlessAuthenticatorScreen.module.scss";

const SETUP_EMAIL_HTML_TEMPLATE_NAME = "setup_primary_oob_email.html";
const SETUP_EMAIL_TEXT_TEMPLATE_NAME = "setup_primary_oob_email.txt";
const SETUP_SMS_TEXT_TEMPLATE_NAME = "setup_primary_oob_sms.txt";
const AUTHENTICATE_EMAIL_HTML_TEMPLATE_NAME =
  "authenticate_secondary_oob_email.html";
const AUTHENTICATE_EMAIL_TEXT_TEMPLATE_NAME =
  "authenticate_secondary_oob_email.txt";
const AUTHENTICATE_SMS_TEXT_TEMPLATE_NAME =
  "authenticate_secondary_oob_sms.txt";

interface PasswordlessAuthenticatorScreenState {
  setupEmailHtmlTemplate: string;
  setupEmailPlainTextTemplate: string;
  setupSmsTemplate: string;
  authenticateEmailHtmlTemplate: string;
  authenticateEmailPlainTextTemplate: string;
  authenticateSmsTemplate: string;
}

interface PasswordlessAuthenticatorProps {
  rawAppConfig: PortalAPIAppConfig | null;
  setupEmailAndSmsTemplates: PortalAPIEmailAndSmsTemplates | null;
  authenticateEmailAndSmsTemplates: PortalAPIEmailAndSmsTemplates | null;
  updateAppAndSetupEmailSmsTemplatesConfig: AppAndEmailSmsTemplatesConfigUpdater;
  updateAppAndAuthenticateEmailSmsTemplatesConfig: AppAndEmailSmsTemplatesConfigUpdater;
  updatingAppAndSetupEmailSmsTemplateConfig: boolean;
  updatingAppAndAuthenticateEmailSmsTemplateConfig: boolean;
}

const PasswordlessAuthenticator: React.FC<PasswordlessAuthenticatorProps> = function PasswordlessAuthenticator(
  props: PasswordlessAuthenticatorProps
) {
  const {
    rawAppConfig,
    setupEmailAndSmsTemplates,
    authenticateEmailAndSmsTemplates,
    updateAppAndSetupEmailSmsTemplatesConfig,
    updateAppAndAuthenticateEmailSmsTemplatesConfig,
    updatingAppAndSetupEmailSmsTemplateConfig,
    updatingAppAndAuthenticateEmailSmsTemplateConfig,
  } = props;

  const initialState: PasswordlessAuthenticatorScreenState = useMemo(() => {
    return {
      setupEmailHtmlTemplate: setupEmailAndSmsTemplates?.emailHtml ?? "",
      setupEmailPlainTextTemplate: setupEmailAndSmsTemplates?.emailText ?? "",
      setupSmsTemplate: setupEmailAndSmsTemplates?.smsText ?? "",
      authenticateEmailHtmlTemplate:
        authenticateEmailAndSmsTemplates?.emailHtml ?? "",
      authenticateEmailPlainTextTemplate:
        authenticateEmailAndSmsTemplates?.emailText ?? "",
      authenticateSmsTemplate: authenticateEmailAndSmsTemplates?.smsText ?? "",
    };
  }, [setupEmailAndSmsTemplates, authenticateEmailAndSmsTemplates]);

  const [state, setState] = useState<PasswordlessAuthenticatorScreenState>(
    initialState
  );

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

  const resetForm = useCallback(() => {
    setState(initialState);
  }, [initialState]);

  const onSetupEmailHtmlTemplateChange = useCallback(
    (_event: unknown, value: string | undefined) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        setupEmailHtmlTemplate: value,
      }));
    },
    []
  );

  const onSetupEmailPlainTextTemplateChange = useCallback(
    (_event: unknown, value: string | undefined) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        setupEmailPlainTextTemplate: value,
      }));
    },
    []
  );

  const onSetupSmsTemplateChange = useCallback((_event, value?: string) => {
    if (value === undefined) {
      return;
    }
    setState((state) => ({
      ...state,
      setupSmsTemplate: value,
    }));
  }, []);

  const onAuthenticateEmailHtmlTemplateChange = useCallback(
    (_event: unknown, value: string | undefined) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        authenticateEmailHtmlTemplate: value,
      }));
    },
    []
  );

  const onAuthenticateEmailPlainTextTemplateChange = useCallback(
    (_event: unknown, value: string | undefined) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        authenticateEmailPlainTextTemplate: value,
      }));
    },
    []
  );

  const onAuthenticateSmsTemplateChange = useCallback(
    (_event, value?: string) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        authenticateSmsTemplate: value,
      }));
    },
    []
  );

  const onSaveButtonClicked = useCallback(() => {
    if (rawAppConfig == null) {
      return;
    }

    // eslint-disable-next-line complexity
    const newAppConfig = produce(rawAppConfig, (draftConfig) => {
      draftConfig.template = draftConfig.template ?? {};
      draftConfig.template.items = draftConfig.template.items ?? [];

      if (
        state.setupEmailHtmlTemplate !== initialState.setupEmailHtmlTemplate &&
        !draftConfig.template.items.some(
          (item) => item.type === SETUP_EMAIL_HTML_TEMPLATE_NAME
        )
      ) {
        draftConfig.template.items.push({
          type: SETUP_EMAIL_HTML_TEMPLATE_NAME,
          uri: `file:///templates/${SETUP_EMAIL_HTML_TEMPLATE_NAME}`,
        });
      }

      if (
        state.setupEmailPlainTextTemplate !==
          initialState.setupEmailPlainTextTemplate &&
        !draftConfig.template.items.some(
          (item) => item.type === SETUP_EMAIL_TEXT_TEMPLATE_NAME
        )
      ) {
        draftConfig.template.items.push({
          type: SETUP_EMAIL_TEXT_TEMPLATE_NAME,
          uri: `file:///templates/${SETUP_EMAIL_TEXT_TEMPLATE_NAME}`,
        });
      }

      if (
        state.setupSmsTemplate !== initialState.setupSmsTemplate &&
        !draftConfig.template.items.some(
          (item) => item.type === SETUP_SMS_TEXT_TEMPLATE_NAME
        )
      ) {
        draftConfig.template.items.push({
          type: SETUP_SMS_TEXT_TEMPLATE_NAME,
          uri: `file:///templates/${SETUP_SMS_TEXT_TEMPLATE_NAME}`,
        });
      }

      if (
        state.authenticateEmailHtmlTemplate !==
          initialState.authenticateEmailHtmlTemplate &&
        !draftConfig.template.items.some(
          (item) => item.type === AUTHENTICATE_EMAIL_HTML_TEMPLATE_NAME
        )
      ) {
        draftConfig.template.items.push({
          type: AUTHENTICATE_EMAIL_HTML_TEMPLATE_NAME,
          uri: `file:///templates/${AUTHENTICATE_EMAIL_HTML_TEMPLATE_NAME}`,
        });
      }

      if (
        state.authenticateEmailPlainTextTemplate !==
          initialState.authenticateEmailPlainTextTemplate &&
        !draftConfig.template.items.some(
          (item) => item.type === AUTHENTICATE_EMAIL_TEXT_TEMPLATE_NAME
        )
      ) {
        draftConfig.template.items.push({
          type: AUTHENTICATE_EMAIL_TEXT_TEMPLATE_NAME,
          uri: `file:///templates/${AUTHENTICATE_EMAIL_TEXT_TEMPLATE_NAME}`,
        });
      }

      if (
        state.authenticateSmsTemplate !==
          initialState.authenticateSmsTemplate &&
        !draftConfig.template.items.some(
          (item) => item.type === AUTHENTICATE_SMS_TEXT_TEMPLATE_NAME
        )
      ) {
        draftConfig.template.items.push({
          type: AUTHENTICATE_SMS_TEXT_TEMPLATE_NAME,
          uri: `file:///templates/${AUTHENTICATE_SMS_TEXT_TEMPLATE_NAME}`,
        });
      }

      clearEmptyObject(draftConfig);
    });

    updateAppAndSetupEmailSmsTemplatesConfig(newAppConfig, {
      emailHtml:
        state.setupEmailHtmlTemplate !== initialState.setupEmailHtmlTemplate
          ? state.setupEmailHtmlTemplate
          : undefined,
      emailText:
        state.setupEmailPlainTextTemplate !==
        initialState.setupEmailPlainTextTemplate
          ? state.setupEmailPlainTextTemplate
          : undefined,
      smsText:
        state.setupSmsTemplate !== initialState.setupSmsTemplate
          ? state.setupSmsTemplate
          : undefined,
    }).catch(() => {});

    updateAppAndAuthenticateEmailSmsTemplatesConfig(newAppConfig, {
      emailHtml:
        state.authenticateEmailHtmlTemplate !==
        initialState.authenticateEmailHtmlTemplate
          ? state.authenticateEmailHtmlTemplate
          : undefined,
      emailText:
        state.authenticateEmailPlainTextTemplate !==
        initialState.authenticateEmailPlainTextTemplate
          ? state.authenticateEmailPlainTextTemplate
          : undefined,
      smsText:
        state.authenticateSmsTemplate !== initialState.authenticateSmsTemplate
          ? state.authenticateSmsTemplate
          : undefined,
    }).catch(() => {});
  }, [
    state,
    rawAppConfig,
    initialState,
    updateAppAndSetupEmailSmsTemplatesConfig,
    updateAppAndAuthenticateEmailSmsTemplatesConfig,
  ]);

  return (
    <div className={styles.form}>
      <Label className={styles.boldLabel}>
        <FormattedMessage id="PasswordlessAuthenticatorScreen.first-time-setup.label" />
      </Label>

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorScreen.first-time-setup.html-email.label" />
      </Label>
      <CodeEditor
        className={styles.htmlCodeEditor}
        language="html"
        value={state.setupEmailHtmlTemplate}
        onChange={onSetupEmailHtmlTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorScreen.first-time-setup.plaintext-email.label" />
      </Label>
      <CodeEditor
        className={styles.plainTextCodeEditor}
        language="plaintext"
        value={state.setupEmailPlainTextTemplate}
        onChange={onSetupEmailPlainTextTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorScreen.first-time-setup.sms-body.label" />
      </Label>
      <CodeEditor
        className={styles.plainTextCodeEditor}
        language="plaintext"
        value={state.setupSmsTemplate}
        onChange={onSetupSmsTemplateChange}
      />

      <Label className={styles.boldLabel}>
        <FormattedMessage id="PasswordlessAuthenticatorScreen.subsequent-logins.label" />
      </Label>

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorScreen.subsequent-logins.html-email.label" />
      </Label>
      <CodeEditor
        className={styles.htmlCodeEditor}
        language="html"
        value={state.authenticateEmailHtmlTemplate}
        onChange={onAuthenticateEmailHtmlTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorScreen.subsequent-logins.plaintext-email.label" />
      </Label>
      <CodeEditor
        className={styles.plainTextCodeEditor}
        language="plaintext"
        value={state.authenticateEmailPlainTextTemplate}
        onChange={onAuthenticateEmailPlainTextTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorScreen.subsequent-logins.sms-body.label" />
      </Label>
      <CodeEditor
        className={styles.plainTextCodeEditor}
        language="plaintext"
        value={state.authenticateSmsTemplate}
        onChange={onAuthenticateSmsTemplateChange}
      />

      <div className={styles.saveButtonContainer}>
        <ButtonWithLoading
          disabled={!isFormModified}
          onClick={onSaveButtonClicked}
          loading={
            updatingAppAndSetupEmailSmsTemplateConfig ||
            updatingAppAndAuthenticateEmailSmsTemplateConfig
          }
          labelId="save"
          loadingLabelId="saving"
        />
      </div>

      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <ModifiedIndicatorPortal
        resetForm={resetForm}
        isModified={isFormModified}
      />
    </div>
  );
};

const PasswordlessAuthenticatorScreen: React.FC = function PasswordlessAuthenticatorScreen() {
  const { appID } = useParams();

  const {
    updateAppAndEmailSmsTemplatesConfig: updateAppAndSetupEmailSmsTemplatesConfig,
    loading: updatingAppAndSetupEmailSmsTemplateConfig,
    error: updateAppAndSetupEmailSmsTemplateConfigError,
  } = useUpdateAppAndEmailSmsTemplatesConfigMutation(
    appID,
    `templates/${SETUP_EMAIL_HTML_TEMPLATE_NAME}`,
    `templates/${SETUP_EMAIL_TEXT_TEMPLATE_NAME}`,
    `templates/${SETUP_SMS_TEXT_TEMPLATE_NAME}`
  );
  const {
    updateAppAndEmailSmsTemplatesConfig: updateAppAndAuthenticateEmailSmsTemplatesConfig,
    loading: updatingAppAndAuthenticateEmailSmsTemplateConfig,
    error: updateAppAndAuthenticateEmailSmsTemplateConfigError,
  } = useUpdateAppAndEmailSmsTemplatesConfigMutation(
    appID,
    `templates/${AUTHENTICATE_EMAIL_HTML_TEMPLATE_NAME}`,
    `templates/${AUTHENTICATE_EMAIL_TEXT_TEMPLATE_NAME}`,
    `templates/${AUTHENTICATE_SMS_TEXT_TEMPLATE_NAME}`
  );

  const {
    emailAndSmsTemplates: setupEmailAndSmsTemplates,
    rawAppConfig,
    loading: loadingSetupEmailAndSmsTemplates,
    error: setupEmailAndSmsTemplatesError,
    refetch: refetchSetupEmailAndSmsTemplates,
  } = useAppAndEmailSmsTemplatesQuery(
    appID,
    `templates/${SETUP_EMAIL_HTML_TEMPLATE_NAME}`,
    `templates/${SETUP_EMAIL_TEXT_TEMPLATE_NAME}`,
    `templates/${SETUP_SMS_TEXT_TEMPLATE_NAME}`
  );
  const {
    emailAndSmsTemplates: authenticateEmailAndSmsTemplates,
    loading: loadingAuthenticateEmailAndSmsTemplates,
    error: authenticateEmailAndSmsTemplatesError,
    refetch: refetchAuthenticateEmailAndSmsTemplates,
  } = useAppAndEmailSmsTemplatesQuery(
    appID,
    `templates/${AUTHENTICATE_EMAIL_HTML_TEMPLATE_NAME}`,
    `templates/${AUTHENTICATE_EMAIL_TEXT_TEMPLATE_NAME}`,
    `templates/${AUTHENTICATE_SMS_TEXT_TEMPLATE_NAME}`
  );

  if (
    loadingAuthenticateEmailAndSmsTemplates ||
    loadingSetupEmailAndSmsTemplates
  ) {
    return <ShowLoading />;
  }

  if (setupEmailAndSmsTemplatesError != null) {
    return (
      <ShowError
        error={setupEmailAndSmsTemplatesError}
        onRetry={refetchSetupEmailAndSmsTemplates}
      />
    );
  }
  if (authenticateEmailAndSmsTemplatesError != null) {
    return (
      <ShowError
        error={authenticateEmailAndSmsTemplatesError}
        onRetry={refetchAuthenticateEmailAndSmsTemplates}
      />
    );
  }

  return (
    <main
      className={cn(styles.root, {
        [styles.loading]: updatingAppAndAuthenticateEmailSmsTemplateConfig,
      })}
    >
      {updateAppAndAuthenticateEmailSmsTemplateConfigError && (
        <ShowError
          error={updateAppAndAuthenticateEmailSmsTemplateConfigError}
        />
      )}
      {updateAppAndSetupEmailSmsTemplateConfigError && (
        <ShowError error={updateAppAndSetupEmailSmsTemplateConfigError} />
      )}
      <ModifiedIndicatorWrapper className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="PasswordlessAuthenticatorScreen.title" />
        </Text>
        <PasswordlessAuthenticator
          setupEmailAndSmsTemplates={setupEmailAndSmsTemplates}
          authenticateEmailAndSmsTemplates={authenticateEmailAndSmsTemplates}
          rawAppConfig={rawAppConfig}
          updateAppAndSetupEmailSmsTemplatesConfig={
            updateAppAndSetupEmailSmsTemplatesConfig
          }
          updateAppAndAuthenticateEmailSmsTemplatesConfig={
            updateAppAndAuthenticateEmailSmsTemplatesConfig
          }
          updatingAppAndSetupEmailSmsTemplateConfig={
            updatingAppAndSetupEmailSmsTemplateConfig
          }
          updatingAppAndAuthenticateEmailSmsTemplateConfig={
            updatingAppAndAuthenticateEmailSmsTemplateConfig
          }
        />
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default PasswordlessAuthenticatorScreen;
