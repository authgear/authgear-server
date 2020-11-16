import React, { useCallback, useMemo, useState } from "react";
import cn from "classnames";
import deepEqual from "deep-equal";
import { Label } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

import CodeEditor from "../../CodeEditor";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ButtonWithLoading from "../../ButtonWithLoading";
import { ModifiedIndicatorPortal } from "../../ModifiedIndicatorPortal";
import {
  AppTemplatesUpdater,
  UpdateAppTemplatesData,
} from "./mutations/updateAppTemplatesMutation";
import {
  forgotPasswordEmailHtmlPath,
  forgotPasswordEmailTextPath,
  forgotPasswordSmsTextPath,
  getLocalizedTemplatePath,
  setUpdateTemplatesData,
  TemplateLocale,
  TemplateMap,
} from "../../templates";

import styles from "./ForgotPasswordTemplatesSettings.module.scss";

interface ForgotPasswordTemplatesSettingsProps {
  className?: string;
  templates: Record<string, string>;
  templateLocale: TemplateLocale;
  updateTemplates: AppTemplatesUpdater;
  updatingTemplates: boolean;
  resetForm: () => void;
}

interface ForgotPasswordTemplatesSettingsState {
  emailHtmlTemplate: string;
  emailPlainTextTemplate: string;
  smsTemplate: string;
}

function constructStateFromTemplates(
  templates: TemplateMap,
  templateLocale: TemplateLocale
): ForgotPasswordTemplatesSettingsState {
  return {
    emailHtmlTemplate:
      templates[
        getLocalizedTemplatePath(templateLocale, forgotPasswordEmailHtmlPath)
      ],
    emailPlainTextTemplate:
      templates[
        getLocalizedTemplatePath(templateLocale, forgotPasswordEmailTextPath)
      ],
    smsTemplate:
      templates[
        getLocalizedTemplatePath(templateLocale, forgotPasswordSmsTextPath)
      ],
  };
}

function constructUpdateTemplatesDataFromState(
  templateLocale: TemplateLocale,
  initialScreenState: ForgotPasswordTemplatesSettingsState,
  screenState: ForgotPasswordTemplatesSettingsState
): UpdateAppTemplatesData {
  const updateTemplatesData: UpdateAppTemplatesData = {};
  if (screenState.emailHtmlTemplate !== initialScreenState.emailHtmlTemplate) {
    setUpdateTemplatesData(
      updateTemplatesData,
      forgotPasswordEmailHtmlPath,
      templateLocale,
      screenState.emailHtmlTemplate
    );
  }
  if (
    screenState.emailPlainTextTemplate !==
    initialScreenState.emailPlainTextTemplate
  ) {
    setUpdateTemplatesData(
      updateTemplatesData,
      forgotPasswordEmailTextPath,
      templateLocale,
      screenState.emailPlainTextTemplate
    );
  }
  if (screenState.smsTemplate !== initialScreenState.smsTemplate) {
    setUpdateTemplatesData(
      updateTemplatesData,
      forgotPasswordSmsTextPath,
      templateLocale,
      screenState.smsTemplate
    );
  }

  return updateTemplatesData;
}

const ForgotPasswordTemplatesSettings: React.FC<ForgotPasswordTemplatesSettingsProps> = function ForgotPasswordTemplatesSettings(
  props: ForgotPasswordTemplatesSettingsProps
) {
  const {
    className,
    templates,
    templateLocale,
    updateTemplates,
    updatingTemplates,
    resetForm,
  } = props;

  const initialState = useMemo(() => {
    return constructStateFromTemplates(templates, templateLocale);
  }, [templates, templateLocale]);

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

  const onSmsTemplateChange = useCallback((_event, value?: string) => {
    if (value === undefined) {
      return;
    }
    setState((state) => ({
      ...state,
      smsTemplate: value,
    }));
  }, []);

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      const updateTemplatesData = constructUpdateTemplatesDataFromState(
        templateLocale,
        initialState,
        state
      );

      updateTemplates(updateTemplatesData).catch(() => {});
    },
    [templateLocale, initialState, state, updateTemplates]
  );

  return (
    <form className={cn(styles.form, className)} onSubmit={onFormSubmit}>
      <ModifiedIndicatorPortal
        resetForm={resetForm}
        isModified={isFormModified}
      />
      <NavigationBlockerDialog blockNavigation={isFormModified} />
      <Label className={styles.boldLabel}>
        <FormattedMessage id="ForgotPasswordTemplatesSettings.email.label" />
      </Label>

      <Label className={styles.label}>
        <FormattedMessage id="ForgotPasswordTemplatesSettings.email.html-email.label" />
      </Label>
      <CodeEditor
        className={styles.htmlCodeEditor}
        language="html"
        value={state.emailHtmlTemplate}
        onChange={onEmailHtmlTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="ForgotPasswordTemplatesSettings.email.plaintext-email.label" />
      </Label>
      <CodeEditor
        className={styles.plainTextCodeEditor}
        language="plaintext"
        value={state.emailPlainTextTemplate}
        onChange={onEmailPlainTextTemplateChange}
      />

      <Label className={styles.boldLabel}>
        <FormattedMessage id="ForgotPasswordTemplatesSettings.sms.label" />
      </Label>

      <Label className={styles.label}>
        <FormattedMessage id="ForgotPasswordTemplatesSettings.sms.sms-body.label" />
      </Label>
      <CodeEditor
        className={styles.plainTextCodeEditor}
        language="plaintext"
        value={state.smsTemplate}
        onChange={onSmsTemplateChange}
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

export default ForgotPasswordTemplatesSettings;
