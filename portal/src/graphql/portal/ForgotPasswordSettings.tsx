import React, { useMemo, useContext, useState, useCallback } from "react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { TextField, Label } from "@fluentui/react";
import cn from "classnames";
import deepEqual from "deep-equal";
import produce from "immer";

import CodeEditor from "../../CodeEditor";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { setFieldIfChanged, clearEmptyObject } from "../../util/misc";
import { PortalAPIAppConfig, PortalAPIApp } from "../../types";

import styles from "./ForgotPasswordSettings.module.scss";
import TodoButtonWrapper from "../../TodoButtonWrapper";

interface ForgotPasswordSettingsProps {
  className?: string;
  effectiveAppConfig: PortalAPIAppConfig | null;
  rawAppConfig: PortalAPIAppConfig | null;
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>;
  updatingAppConfig: boolean;
}

interface ForgotPasswordSettingsState {
  emailHtmlTemplate: string;
  emailPlainTextTemplate: string;
  smsTemplate: string;
  resetCodeExpirySeconds: number;
}

function constructStateFromAppConfig(
  appConfig: PortalAPIAppConfig | null
): ForgotPasswordSettingsState {
  const forgot_password = appConfig?.forgot_password;

  return {
    emailHtmlTemplate: "", // TODO: handle email template
    emailPlainTextTemplate: "", // TODO: handle email template
    smsTemplate: "", // TODO: handle sms template
    resetCodeExpirySeconds: forgot_password?.reset_code_expiry_seconds ?? 0,
  };
}

function constructAppConfigFromState(
  rawAppConfig: PortalAPIAppConfig,
  initialScreenState: ForgotPasswordSettingsState,
  screenState: ForgotPasswordSettingsState
): PortalAPIAppConfig {
  const newAppConfig = produce(rawAppConfig, (draftConfig) => {
    draftConfig.forgot_password = draftConfig.forgot_password ?? {};

    const forgotPassword = draftConfig.forgot_password;

    setFieldIfChanged(
      forgotPassword,
      "reset_code_expiry_seconds",
      initialScreenState.resetCodeExpirySeconds,
      screenState.resetCodeExpirySeconds
    );

    // TODO: update email template
    // TODO: update sms template

    clearEmptyObject(draftConfig);
  });

  return newAppConfig;
}

const ForgotPasswordSettings: React.FC<ForgotPasswordSettingsProps> = function ForgotPasswordSettings(
  props
) {
  const {
    className,
    effectiveAppConfig,
    rawAppConfig,
    updateAppConfig,
    updatingAppConfig,
  } = props;

  const { renderToString } = useContext(Context);

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

      const newAppConfig = constructAppConfigFromState(
        rawAppConfig,
        initialState,
        state
      );

      // TODO: handle error
      updateAppConfig(newAppConfig).catch(() => {});
    },
    [state, rawAppConfig, updateAppConfig, initialState]
  );

  return (
    <form className={cn(styles.root, className)} onSubmit={onFormSubmit}>
      <Label className={styles.boldLabel}>
        <FormattedMessage id="PasswordsScreen.forgot-password.email.label" />
      </Label>

      <Label className={styles.label}>
        <FormattedMessage id="PasswordsScreen.forgot-password.email.styled-content.label" />
      </Label>
      <CodeEditor
        className={styles.htmlCodeEditor}
        language="html"
        value={state.emailHtmlTemplate}
        onChange={onEmailHtmlTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="PasswordsScreen.forgot-password.email.plain-content.label" />
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
        <FormattedMessage id="PasswordsScreen.forgot-password.sms.content.label" />
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
        <TodoButtonWrapper>
          <ButtonWithLoading
            type="submit"
            disabled={!isFormModified}
            loading={updatingAppConfig}
            labelId="save"
            loadingLabelId="saving"
          />
        </TodoButtonWrapper>
      </div>
      <NavigationBlockerDialog blockNavigation={isFormModified} />
    </form>
  );
};

export default ForgotPasswordSettings;
