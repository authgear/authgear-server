import React, { useCallback, useMemo, useState } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Label } from "@fluentui/react";
import deepEqual from "deep-equal";

import { AppTemplatesUpdater } from "./mutations/updateAppTemplatesMutation";
import CodeEditor from "../../CodeEditor";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { ModifiedIndicatorPortal } from "../../ModifiedIndicatorPortal";
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

import styles from "./PasswordlessAuthenticatorTemplatesSettings.module.scss";

type PrimaryOOBMessageTemplateKeys =
  | typeof SetupPrimaryOOBMessageTemplates[number]
  | typeof AuthenticatePrimaryOOBMessageTemplates[number];

interface PasswordlessAuthenticatorTemplatesState {
  setupEmailHtmlTemplate: string;
  setupEmailPlainTextTemplate: string;
  setupSmsTemplate: string;
  authenticateEmailHtmlTemplate: string;
  authenticateEmailPlainTextTemplate: string;
  authenticateSmsTemplate: string;
}

interface PasswordlessAuthenticatorTemplatesSettingsProps {
  templates: Record<PrimaryOOBMessageTemplateKeys, string>;
  updateTemplates: AppTemplatesUpdater<PrimaryOOBMessageTemplateKeys>;
  updatingTemplates: boolean;
  resetForm: () => void;
}

const PasswordlessAuthenticatorTemplatesSettings: React.FC<PasswordlessAuthenticatorTemplatesSettingsProps> = function PasswordlessAuthenticatorTemplatesSettings(
  props: PasswordlessAuthenticatorTemplatesSettingsProps
) {
  const { templates, updateTemplates, updatingTemplates, resetForm } = props;

  const initialState: PasswordlessAuthenticatorTemplatesState = useMemo(() => {
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

  const [state, setState] = useState<PasswordlessAuthenticatorTemplatesState>(
    initialState
  );

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

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

  // eslint-disable-next-line complexity
  const onSaveButtonClicked = useCallback(() => {
    const updates: Partial<Record<
      PrimaryOOBMessageTemplateKeys,
      string | null
    >> = {};
    if (state.setupEmailHtmlTemplate !== initialState.setupEmailHtmlTemplate) {
      updates[TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_HTML] =
        state.setupEmailHtmlTemplate !== ""
          ? state.setupEmailHtmlTemplate
          : null;
    }
    if (
      state.setupEmailPlainTextTemplate !==
      initialState.setupEmailPlainTextTemplate
    ) {
      updates[TEMPLATE_SETUP_PRIMARY_OOB_EMAIL_TEXT] =
        state.setupEmailPlainTextTemplate !== ""
          ? state.setupEmailPlainTextTemplate
          : null;
    }
    if (state.setupSmsTemplate !== initialState.setupSmsTemplate) {
      updates[TEMPLATE_SETUP_PRIMARY_OOB_SMS_TEXT] =
        state.setupSmsTemplate !== "" ? state.setupSmsTemplate : null;
    }
    if (
      state.authenticateEmailHtmlTemplate !==
      initialState.authenticateEmailHtmlTemplate
    ) {
      updates[TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_HTML] =
        state.authenticateEmailHtmlTemplate !== ""
          ? state.authenticateEmailHtmlTemplate
          : null;
    }
    if (
      state.authenticateEmailPlainTextTemplate !==
      initialState.authenticateEmailPlainTextTemplate
    ) {
      updates[TEMPLATE_AUTHENTICATE_PRIMARY_OOB_EMAIL_TEXT] =
        state.authenticateEmailPlainTextTemplate !== ""
          ? state.authenticateEmailPlainTextTemplate
          : null;
    }
    if (
      state.authenticateSmsTemplate !== initialState.authenticateSmsTemplate
    ) {
      updates[TEMPLATE_AUTHENTICATE_PRIMARY_OOB_SMS_TEXT] =
        state.authenticateSmsTemplate !== ""
          ? state.authenticateSmsTemplate
          : null;
    }

    updateTemplates(updates).catch(() => {});
  }, [state, initialState, updateTemplates]);

  return (
    <div className={styles.form}>
      <Label className={styles.boldLabel}>
        <FormattedMessage id="PasswordlessAuthenticatorTemplatesSettings.first-time-setup.label" />
      </Label>

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorTemplatesSettings.first-time-setup.html-email.label" />
      </Label>
      <CodeEditor
        className={styles.htmlCodeEditor}
        language="html"
        value={state.setupEmailHtmlTemplate}
        onChange={onSetupEmailHtmlTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorTemplatesSettings.first-time-setup.plaintext-email.label" />
      </Label>
      <CodeEditor
        className={styles.plainTextCodeEditor}
        language="plaintext"
        value={state.setupEmailPlainTextTemplate}
        onChange={onSetupEmailPlainTextTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorTemplatesSettings.first-time-setup.sms-body.label" />
      </Label>
      <CodeEditor
        className={styles.plainTextCodeEditor}
        language="plaintext"
        value={state.setupSmsTemplate}
        onChange={onSetupSmsTemplateChange}
      />

      <Label className={styles.boldLabel}>
        <FormattedMessage id="PasswordlessAuthenticatorTemplatesSettings.subsequent-logins.label" />
      </Label>

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorTemplatesSettings.subsequent-logins.html-email.label" />
      </Label>
      <CodeEditor
        className={styles.htmlCodeEditor}
        language="html"
        value={state.authenticateEmailHtmlTemplate}
        onChange={onAuthenticateEmailHtmlTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorTemplatesSettings.subsequent-logins.plaintext-email.label" />
      </Label>
      <CodeEditor
        className={styles.plainTextCodeEditor}
        language="plaintext"
        value={state.authenticateEmailPlainTextTemplate}
        onChange={onAuthenticateEmailPlainTextTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorTemplatesSettings.subsequent-logins.sms-body.label" />
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
          loading={updatingTemplates}
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

export default PasswordlessAuthenticatorTemplatesSettings;
