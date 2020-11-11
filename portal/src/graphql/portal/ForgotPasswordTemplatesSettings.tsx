import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import deepEqual from "deep-equal";
import { Label } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import CodeEditor from "../../CodeEditor";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import ButtonWithLoading from "../../ButtonWithLoading";
import { ModifiedIndicatorPortal } from "../../ModifiedIndicatorPortal";
import { AppTemplatesUpdater } from "./mutations/updateAppTemplatesMutation";
import {
  ForgotPasswordMessageTemplates,
  TEMPLATE_FORGOT_PASSWORD_EMAIL_HTML,
  TEMPLATE_FORGOT_PASSWORD_EMAIL_TEXT,
  TEMPLATE_FORGOT_PASSWORD_SMS_TEXT,
} from "../../templates";

import styles from "./ForgotPasswordTemplatesSettings.module.scss";

interface ForgotPasswordTemplatesSettingsProps {
  className?: string;
  templates: Record<ForgotPasswordMessageTemplateKeys, string>;
  updateTemplates: AppTemplatesUpdater<ForgotPasswordMessageTemplateKeys>;
  updatingTemplates: boolean;
  resetForm: () => void;
}

interface ForgotPasswordTemplatesSettingsState {
  emailHtmlTemplate: string;
  emailPlainTextTemplate: string;
  smsTemplate: string;
}

type ForgotPasswordMessageTemplateKeys = typeof ForgotPasswordMessageTemplates[number];

function constructStateFromTemplates(
  templates: Record<ForgotPasswordMessageTemplateKeys, string>
): ForgotPasswordTemplatesSettingsState {
  return {
    emailHtmlTemplate: templates[TEMPLATE_FORGOT_PASSWORD_EMAIL_HTML],
    emailPlainTextTemplate: templates[TEMPLATE_FORGOT_PASSWORD_EMAIL_TEXT],
    smsTemplate: templates[TEMPLATE_FORGOT_PASSWORD_SMS_TEXT],
  };
}

function constructUpdateTemplatesDataFromState(
  initialScreenState: ForgotPasswordTemplatesSettingsState,
  screenState: ForgotPasswordTemplatesSettingsState
): Partial<Record<ForgotPasswordMessageTemplateKeys, string | null>> {
  const updateTemplatesData: Partial<Record<
    ForgotPasswordMessageTemplateKeys,
    string | null
  >> = {};
  if (screenState.emailHtmlTemplate !== initialScreenState.emailHtmlTemplate) {
    updateTemplatesData[TEMPLATE_FORGOT_PASSWORD_EMAIL_HTML] =
      screenState.emailHtmlTemplate !== ""
        ? screenState.emailHtmlTemplate
        : null;
  }
  if (
    screenState.emailPlainTextTemplate !==
    initialScreenState.emailPlainTextTemplate
  ) {
    updateTemplatesData[TEMPLATE_FORGOT_PASSWORD_EMAIL_TEXT] =
      screenState.emailPlainTextTemplate !== ""
        ? screenState.emailPlainTextTemplate
        : null;
  }
  if (screenState.smsTemplate !== initialScreenState.smsTemplate) {
    updateTemplatesData[TEMPLATE_FORGOT_PASSWORD_SMS_TEXT] =
      screenState.smsTemplate !== "" ? screenState.smsTemplate : null;
  }

  return updateTemplatesData;
}

const ForgotPasswordTemplatesSettings: React.FC<ForgotPasswordTemplatesSettingsProps> = function ForgotPasswordTemplatesSettings(
  props: ForgotPasswordTemplatesSettingsProps
) {
  const {
    className,
    templates,
    updateTemplates,
    updatingTemplates,
    resetForm,
  } = props;

  const { renderToString } = useContext(Context);

  const initialState = useMemo(() => {
    return constructStateFromTemplates(templates);
  }, [templates]);

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
        initialState,
        state
      );

      updateTemplates(updateTemplatesData).catch(() => {});
    },
    [initialState, state, updateTemplates]
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
