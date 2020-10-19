import React, { useCallback, useMemo, useState } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Label, Text } from "@fluentui/react";
import { useParams } from "react-router-dom";
import cn from "classnames";
import deepEqual from "deep-equal";

import {
  AppTemplatesUpdater,
  useUpdateAppTemplatesMutation,
} from "./mutations/updateAppTemplatesMutation";
import { useAppTemplatesQuery } from "./query/appTemplatesQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import CodeEditor from "../../CodeEditor";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import {
  ModifiedIndicatorPortal,
  ModifiedIndicatorWrapper,
} from "../../ModifiedIndicatorPortal";

import styles from "./PasswordlessAuthenticatorScreen.module.scss";
import {
  SetupPrimaryOOBMessageTemplates,
  AuthenticatePrimaryOOBMessageTemplates,
  TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML,
  TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TEXT,
  TEMPLATE_AUTHENTICATE_PRIMARY_OOB_SMS_TEXT,
  TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_HTML,
  TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_TEXT,
  TEMPLATE_SETUP_PRIMARY_OOB_SMS_TEXT,
} from "../../templates";

type PrimaryOOBMessageTemplates =
  | typeof SetupPrimaryOOBMessageTemplates[number]
  | typeof AuthenticatePrimaryOOBMessageTemplates[number];

interface PasswordlessAuthenticatorScreenState {
  setupEmailHtmlTemplate: string;
  setupEmailPlainTextTemplate: string;
  setupSmsTemplate: string;
  authenticateEmailHtmlTemplate: string;
  authenticateEmailPlainTextTemplate: string;
  authenticateSmsTemplate: string;
}

interface PasswordlessAuthenticatorProps {
  templates: Record<PrimaryOOBMessageTemplates, string>;
  updateTemplates: AppTemplatesUpdater<PrimaryOOBMessageTemplates>;
  isUpdatingTemplates: boolean;
}

const PasswordlessAuthenticator: React.FC<PasswordlessAuthenticatorProps> = function PasswordlessAuthenticator(
  props: PasswordlessAuthenticatorProps
) {
  const { templates, updateTemplates, isUpdatingTemplates } = props;

  const initialState: PasswordlessAuthenticatorScreenState = useMemo(() => {
    return {
      setupEmailHtmlTemplate: templates[TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_HTML],
      setupEmailPlainTextTemplate:
        templates[TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_TEXT],
      setupSmsTemplate: templates[TEMPLATE_SETUP_PRIMARY_OOB_SMS_TEXT],
      authenticateEmailHtmlTemplate:
        templates[TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML],
      authenticateEmailPlainTextTemplate:
        templates[TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TEXT],
      authenticateSmsTemplate:
        templates[TEMPLATE_AUTHENTICATE_PRIMARY_OOB_SMS_TEXT],
    };
  }, [templates]);

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
    const updates: Partial<Record<PrimaryOOBMessageTemplates, string>> = {};
    if (state.setupEmailHtmlTemplate !== initialState.setupEmailHtmlTemplate) {
      updates[TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_HTML] =
        state.setupEmailHtmlTemplate;
    }
    if (
      state.setupEmailPlainTextTemplate !==
      initialState.setupEmailPlainTextTemplate
    ) {
      updates[TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_TEXT] =
        state.setupEmailPlainTextTemplate;
    }
    if (state.setupSmsTemplate !== initialState.setupSmsTemplate) {
      updates[TEMPLATE_SETUP_PRIMARY_OOB_SMS_TEXT] = state.setupSmsTemplate;
    }
    if (
      state.authenticateEmailHtmlTemplate !==
      initialState.authenticateEmailHtmlTemplate
    ) {
      updates[TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML] =
        state.authenticateEmailHtmlTemplate;
    }
    if (
      state.authenticateEmailPlainTextTemplate !==
      initialState.authenticateEmailPlainTextTemplate
    ) {
      updates[TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TEXT] =
        state.authenticateEmailPlainTextTemplate;
    }
    if (
      state.authenticateSmsTemplate !== initialState.authenticateSmsTemplate
    ) {
      updates[TEMPLATE_AUTHENTICATE_PRIMARY_OOB_SMS_TEXT] =
        state.authenticateSmsTemplate;
    }

    updateTemplates(updates).catch(() => {});
  }, [state, initialState, updateTemplates]);

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
          loading={isUpdatingTemplates}
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
    updateAppTemplates,
    loading: isUpdatingTemplates,
    error: updateTemplatesError,
  } = useUpdateAppTemplatesMutation<PrimaryOOBMessageTemplates>(appID);

  const {
    templates,
    loading: isLoadingTemplates,
    error: loadTemplatesError,
    refetch: refetchTemplates,
  } = useAppTemplatesQuery(
    appID,
    ...SetupPrimaryOOBMessageTemplates,
    ...AuthenticatePrimaryOOBMessageTemplates
  );

  if (isLoadingTemplates) {
    return <ShowLoading />;
  }

  if (loadTemplatesError) {
    return <ShowError error={loadTemplatesError} onRetry={refetchTemplates} />;
  }

  return (
    <main
      className={cn(styles.root, {
        [styles.loading]: isUpdatingTemplates,
      })}
    >
      {updateTemplatesError && <ShowError error={updateTemplatesError} />}
      <ModifiedIndicatorWrapper className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="PasswordlessAuthenticatorScreen.title" />
        </Text>
        <PasswordlessAuthenticator
          templates={templates}
          updateTemplates={updateAppTemplates}
          isUpdatingTemplates={isUpdatingTemplates}
        />
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default PasswordlessAuthenticatorScreen;
