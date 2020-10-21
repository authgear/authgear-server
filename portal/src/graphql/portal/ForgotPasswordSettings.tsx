import React, { useMemo, useContext, useState, useCallback } from "react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { TextField, Label } from "@fluentui/react";
import cn from "classnames";
import deepEqual from "deep-equal";
import produce from "immer";

import { UpdateAppTemplatesData } from "./mutations/updateAppTemplatesMutation";
import CodeEditor from "../../CodeEditor";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { ModifiedIndicatorPortal } from "../../ModifiedIndicatorPortal";
import { setFieldIfChanged, clearEmptyObject } from "../../util/misc";
import { PortalAPIApp, PortalAPIAppConfig } from "../../types";
import {
  ForgotPasswordMessageTemplates,
  TEMPLATE_FORGOT_PASSWORD_EMAIL_HTML,
  TEMPLATE_FORGOT_PASSWORD_EMAIL_TEXT,
  TEMPLATE_FORGOT_PASSWORD_SMS_TEXT,
} from "../../templates";

import styles from "./ForgotPasswordSettings.module.scss";

type ForgotPasswordMessageTemplateKeys = typeof ForgotPasswordMessageTemplates[number];

interface ForgotPasswordSettingsProps {
  className?: string;
  effectiveAppConfig: PortalAPIAppConfig | null;
  rawAppConfig: PortalAPIAppConfig | null;
  templates: Record<ForgotPasswordMessageTemplateKeys, string>;
  updateAppConfigAndTemplates: (
    appConfig: PortalAPIAppConfig,
    updateTemplatesData: UpdateAppTemplatesData<
      ForgotPasswordMessageTemplateKeys
    >
  ) => Promise<PortalAPIApp | null>;
  updatingAppConfigAndTemplates: boolean;
}

interface ForgotPasswordSettingsState {
  emailHtmlTemplate: string;
  emailPlainTextTemplate: string;
  smsTemplate: string;
  resetCodeExpirySeconds: number;
}

function constructStateFromAppConfigAndTemplates(
  appConfig: PortalAPIAppConfig | null,
  templates: Record<ForgotPasswordMessageTemplateKeys, string>
): ForgotPasswordSettingsState {
  const forgot_password = appConfig?.forgot_password;

  return {
    emailHtmlTemplate: templates[TEMPLATE_FORGOT_PASSWORD_EMAIL_HTML],
    emailPlainTextTemplate: templates[TEMPLATE_FORGOT_PASSWORD_EMAIL_TEXT],
    smsTemplate: templates[TEMPLATE_FORGOT_PASSWORD_SMS_TEXT],
    resetCodeExpirySeconds: forgot_password?.reset_code_expiry_seconds ?? 0,
  };
}

function constructAppConfigAndUpdateTemplatesDataFromState(
  rawAppConfig: PortalAPIAppConfig,
  initialScreenState: ForgotPasswordSettingsState,
  screenState: ForgotPasswordSettingsState
): {
  appConfig: PortalAPIAppConfig;
  updateTemplatesData: Partial<
    Record<ForgotPasswordMessageTemplateKeys, string | null>
  >;
} {
  const newAppConfig = produce(rawAppConfig, (draftConfig) => {
    draftConfig.forgot_password = draftConfig.forgot_password ?? {};

    const forgotPassword = draftConfig.forgot_password;

    setFieldIfChanged(
      forgotPassword,
      "reset_code_expiry_seconds",
      initialScreenState.resetCodeExpirySeconds,
      screenState.resetCodeExpirySeconds
    );

    clearEmptyObject(draftConfig);
  });

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

  return {
    appConfig: newAppConfig,
    updateTemplatesData,
  };
}

const ForgotPasswordSettings: React.FC<ForgotPasswordSettingsProps> = function ForgotPasswordSettings(
  props
) {
  const {
    className,
    effectiveAppConfig,
    rawAppConfig,
    templates,
    updateAppConfigAndTemplates,
    updatingAppConfigAndTemplates,
  } = props;

  const { renderToString } = useContext(Context);

  const initialState = useMemo(() => {
    return constructStateFromAppConfigAndTemplates(
      effectiveAppConfig,
      templates
    );
  }, [effectiveAppConfig, templates]);

  const [state, setState] = useState(initialState);

  const isFormModified = useMemo(() => {
    return !deepEqual(initialState, state, { strict: true });
  }, [initialState, state]);

  const resetForm = useCallback(() => {
    setState(initialState);
  }, [initialState]);

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

  const onResetCodeExpirySecondsChange = useCallback(
    (_event, value?: string) => {
      if (value === undefined) {
        return;
      }
      setState((state) => ({
        ...state,
        resetCodeExpirySeconds: parseInt(value, 10),
      }));
    },
    []
  );

  const onFormSubmit = useCallback(
    (ev: React.SyntheticEvent<HTMLElement>) => {
      ev.preventDefault();
      ev.stopPropagation();

      if (rawAppConfig == null) {
        return;
      }

      const {
        appConfig: newAppConfig,
        updateTemplatesData,
      } = constructAppConfigAndUpdateTemplatesDataFromState(
        rawAppConfig,
        initialState,
        state
      );

      updateAppConfigAndTemplates(
        newAppConfig,
        updateTemplatesData
      ).catch(() => {});
    },
    [rawAppConfig, initialState, state, updateAppConfigAndTemplates]
  );

  return (
    <form className={cn(styles.root, className)} onSubmit={onFormSubmit}>
      <ModifiedIndicatorPortal
        resetForm={resetForm}
        isModified={isFormModified}
      />
      <Label className={styles.boldLabel}>
        <FormattedMessage id="PasswordsScreen.forgot-password.email.label" />
      </Label>

      <Label className={styles.label}>
        <FormattedMessage id="PasswordsScreen.forgot-password.email.html-email.label" />
      </Label>
      <CodeEditor
        className={styles.htmlCodeEditor}
        language="html"
        value={state.emailHtmlTemplate}
        onChange={onEmailHtmlTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="PasswordsScreen.forgot-password.email.plaintext-email.label" />
      </Label>
      <CodeEditor
        className={styles.plainTextCodeEditor}
        language="plaintext"
        value={state.emailPlainTextTemplate}
        onChange={onEmailPlainTextTemplateChange}
      />

      <Label className={styles.boldLabel}>
        <FormattedMessage id="PasswordsScreen.forgot-password.sms.label" />
      </Label>

      <Label className={styles.label}>
        <FormattedMessage id="PasswordsScreen.forgot-password.sms.sms-body.label" />
      </Label>
      <CodeEditor
        className={styles.plainTextCodeEditor}
        language="plaintext"
        value={state.smsTemplate}
        onChange={onSmsTemplateChange}
      />

      <TextField
        className={styles.textField}
        type="number"
        min="0"
        step="1"
        label={renderToString(
          "PasswordsScreen.forgot-password.time-to-invalid-reset-code.label"
        )}
        value={`${state.resetCodeExpirySeconds}`}
        onChange={onResetCodeExpirySecondsChange}
      />

      <div className={styles.saveButtonContainer}>
        <ButtonWithLoading
          type="submit"
          disabled={!isFormModified}
          loading={updatingAppConfigAndTemplates}
          labelId="save"
          loadingLabelId="saving"
        />
      </div>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
    </form>
  );
};

export default ForgotPasswordSettings;
