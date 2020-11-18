import React, { useCallback, useMemo, useState } from "react";
import { FormattedMessage } from "@oursky/react-messageformat";
import { Label } from "@fluentui/react";
import deepEqual from "deep-equal";

import {
  AppTemplatesUpdater,
  UpdateAppTemplatesData,
} from "./mutations/updateAppTemplatesMutation";
import CodeEditor from "../../CodeEditor";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { ModifiedIndicatorPortal } from "../../ModifiedIndicatorPortal";
import {
  TemplateLocale,
  TemplateMap,
  getLocalizedTemplatePath,
  setUpdateTemplatesData,
  setupPrimaryOobEmailHtmlPath,
  setupPrimaryOobEmailTextPath,
  setupPrimaryOobSmsTextPath,
  authenticatePrimaryOobEmailHtmlPath,
  authenticatePrimaryOobEmailTextPath,
  authenticatePrimaryOobSmsTextPath,
} from "../../templates";

import styles from "./PasswordlessAuthenticatorTemplatesSettings.module.scss";

interface PasswordlessAuthenticatorTemplatesState {
  setupEmailHtmlTemplate: string;
  setupEmailPlainTextTemplate: string;
  setupSmsTemplate: string;
  authenticateEmailHtmlTemplate: string;
  authenticateEmailPlainTextTemplate: string;
  authenticateSmsTemplate: string;
}

interface PasswordlessAuthenticatorTemplatesSettingsProps {
  templates: TemplateMap;
  templateLocale: TemplateLocale;
  updateTemplates: AppTemplatesUpdater;
  updatingTemplates: boolean;
  resetForm: () => void;
}

function constructStateFromTemplates(
  templates: TemplateMap,
  templateLocale: TemplateLocale
): PasswordlessAuthenticatorTemplatesState {
  return {
    setupEmailHtmlTemplate:
      templates[
        getLocalizedTemplatePath(templateLocale, setupPrimaryOobEmailHtmlPath)
      ],
    setupEmailPlainTextTemplate:
      templates[
        getLocalizedTemplatePath(templateLocale, setupPrimaryOobEmailTextPath)
      ],
    setupSmsTemplate:
      templates[
        getLocalizedTemplatePath(templateLocale, setupPrimaryOobSmsTextPath)
      ],
    authenticateEmailHtmlTemplate:
      templates[
        getLocalizedTemplatePath(
          templateLocale,
          authenticatePrimaryOobEmailHtmlPath
        )
      ],
    authenticateEmailPlainTextTemplate:
      templates[
        getLocalizedTemplatePath(
          templateLocale,
          authenticatePrimaryOobEmailTextPath
        )
      ],
    authenticateSmsTemplate:
      templates[
        getLocalizedTemplatePath(
          templateLocale,
          authenticatePrimaryOobSmsTextPath
        )
      ],
  };
}

function constructUpdateTemplatesDataFromState(
  templateLocale: TemplateLocale,
  initialState: PasswordlessAuthenticatorTemplatesState,
  state: PasswordlessAuthenticatorTemplatesState
): UpdateAppTemplatesData {
  const templateUpdates: Partial<Record<string, string | null>> = {};
  if (state.setupEmailHtmlTemplate !== initialState.setupEmailHtmlTemplate) {
    setUpdateTemplatesData(
      templateUpdates,
      setupPrimaryOobEmailHtmlPath,
      templateLocale,
      state.setupEmailHtmlTemplate
    );
  }
  if (
    state.setupEmailPlainTextTemplate !==
    initialState.setupEmailPlainTextTemplate
  ) {
    setUpdateTemplatesData(
      templateUpdates,
      setupPrimaryOobEmailTextPath,
      templateLocale,
      state.setupEmailPlainTextTemplate
    );
  }
  if (state.setupSmsTemplate !== initialState.setupSmsTemplate) {
    setUpdateTemplatesData(
      templateUpdates,
      setupPrimaryOobSmsTextPath,
      templateLocale,
      state.setupSmsTemplate
    );
  }
  if (
    state.authenticateEmailHtmlTemplate !==
    initialState.authenticateEmailHtmlTemplate
  ) {
    setUpdateTemplatesData(
      templateUpdates,
      authenticatePrimaryOobEmailHtmlPath,
      templateLocale,
      state.authenticateEmailHtmlTemplate
    );
  }
  if (
    state.authenticateEmailPlainTextTemplate !==
    initialState.authenticateEmailPlainTextTemplate
  ) {
    setUpdateTemplatesData(
      templateUpdates,
      authenticatePrimaryOobEmailTextPath,
      templateLocale,
      state.authenticateEmailPlainTextTemplate
    );
  }
  if (state.authenticateSmsTemplate !== initialState.authenticateSmsTemplate) {
    setUpdateTemplatesData(
      templateUpdates,
      authenticatePrimaryOobSmsTextPath,
      templateLocale,
      state.authenticateSmsTemplate
    );
  }

  return templateUpdates;
}

const PasswordlessAuthenticatorTemplatesSettings: React.FC<PasswordlessAuthenticatorTemplatesSettingsProps> = function PasswordlessAuthenticatorTemplatesSettings(
  props: PasswordlessAuthenticatorTemplatesSettingsProps
) {
  const {
    templates,
    templateLocale,
    updateTemplates,
    updatingTemplates,
    resetForm,
  } = props;

  const initialState: PasswordlessAuthenticatorTemplatesState = useMemo(() => {
    return constructStateFromTemplates(templates, templateLocale);
  }, [templates, templateLocale]);

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

  const onFormSubmit = useCallback(() => {
    const templateUpdates = constructUpdateTemplatesDataFromState(
      templateLocale,
      initialState,
      state
    );
    updateTemplates(templateUpdates).catch(() => {});
  }, [templateLocale, state, initialState, updateTemplates]);

  return (
    <form className={styles.form} onSubmit={onFormSubmit}>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <ModifiedIndicatorPortal
        resetForm={resetForm}
        isModified={isFormModified}
      />

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
          type="submit"
          disabled={!isFormModified}
          loading={updatingTemplates}
          labelId="save"
          loadingLabelId="saving"
        />
      </div>
    </form>
  );
};

export default PasswordlessAuthenticatorTemplatesSettings;
