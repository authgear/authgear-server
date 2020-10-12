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
import {
  AppAndEmailSmsTemplatesQueryResult,
  useAppAndEmailSmsTemplatesQuery,
} from "./query/appAndEmailSmsTemplatesQuery";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import CodeEditor from "../../CodeEditor";
import { clearEmptyObject } from "../../util/misc";
import ButtonWithLoading from "../../ButtonWithLoading";
import NavigationBlockerDialog from "../../NavigationBlockerDialog";
import { PortalAPIAppConfig, PortalAPIEmailAndSmsTemplates } from "../../types";

import styles from "./PasswordlessAuthenticatorScreen.module.scss";

const EMAIL_HTML_TEMPLATE_NAME = "authenticate_secondary_oob_email.html";
const EMAIL_MJML_TEMPLATE_NAME = "authenticate_secondary_oob_email.mjml";
const EMAIL_TEXT_TEMPLATE_NAME = "authenticate_secondary_oob_email.txt";
const SMS_TEXT_TEMPLATE_NAME = "authenticate_secondary_oob_sms.txt";

interface PasswordlessAuthenticatorScreenState {
  emailHtmlTemplate: string;
  emailPlainTextTemplate: string;
  smsTemplate: string;
}

interface PasswordlessAuthenticatorProps {
  rawAppConfig: PortalAPIAppConfig | null;
  emailAndSmsTemplates: PortalAPIEmailAndSmsTemplates | null;
  updateAppAndEmailSmsTemplatesConfig: AppAndEmailSmsTemplatesConfigUpdater;
  refetchAppAndEmailSmsTemplatesConfig: AppAndEmailSmsTemplatesQueryResult["refetch"];
  updatingAppAndEmailSmsTemplateConfig: boolean;
}

const PasswordlessAuthenticator: React.FC<PasswordlessAuthenticatorProps> = function PasswordlessAuthenticator(
  props: PasswordlessAuthenticatorProps
) {
  const {
    rawAppConfig,
    emailAndSmsTemplates,
    updateAppAndEmailSmsTemplatesConfig,
    refetchAppAndEmailSmsTemplatesConfig,
    updatingAppAndEmailSmsTemplateConfig,
  } = props;

  const initialState: PasswordlessAuthenticatorScreenState = useMemo(() => {
    return {
      emailHtmlTemplate: emailAndSmsTemplates?.emailHtml ?? "",
      emailPlainTextTemplate: emailAndSmsTemplates?.emailText ?? "",
      smsTemplate: emailAndSmsTemplates?.smsText ?? "",
    };
  }, [emailAndSmsTemplates]);

  const [state, setState] = useState<PasswordlessAuthenticatorScreenState>(
    initialState
  );

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

    const newAppConfig = produce(rawAppConfig, (draftConfig) => {
      draftConfig.template = draftConfig.template ?? {};
      draftConfig.template.items = draftConfig.template.items ?? [];

      if (
        state.emailHtmlTemplate !== initialState.emailHtmlTemplate &&
        !draftConfig.template.items.some(
          (item) => item.type === EMAIL_HTML_TEMPLATE_NAME
        )
      ) {
        draftConfig.template.items.push({
          type: EMAIL_HTML_TEMPLATE_NAME,
          uri: `file:///templates/${EMAIL_HTML_TEMPLATE_NAME}`,
        });
      }

      if (
        state.emailPlainTextTemplate !== initialState.emailPlainTextTemplate &&
        !draftConfig.template.items.some(
          (item) => item.type === EMAIL_TEXT_TEMPLATE_NAME
        )
      ) {
        draftConfig.template.items.push({
          type: EMAIL_TEXT_TEMPLATE_NAME,
          uri: `file:///templates/${EMAIL_TEXT_TEMPLATE_NAME}`,
        });
      }

      if (
        state.smsTemplate !== initialState.smsTemplate &&
        !draftConfig.template.items.some(
          (item) => item.type === SMS_TEXT_TEMPLATE_NAME
        )
      ) {
        draftConfig.template.items.push({
          type: SMS_TEXT_TEMPLATE_NAME,
          uri: `file:///templates/${SMS_TEXT_TEMPLATE_NAME}`,
        });
      }

      clearEmptyObject(draftConfig);
    });

    // TODO: handle error
    updateAppAndEmailSmsTemplatesConfig(newAppConfig, {
      emailHtml:
        state.emailHtmlTemplate !== initialState.emailHtmlTemplate
          ? state.emailHtmlTemplate
          : undefined,
      emailText:
        state.emailPlainTextTemplate !== initialState.emailPlainTextTemplate
          ? state.emailPlainTextTemplate
          : undefined,
      smsText:
        state.smsTemplate !== initialState.smsTemplate
          ? state.smsTemplate
          : undefined,
    })
      .then(() => {
        refetchAppAndEmailSmsTemplatesConfig().catch(() => {});
      })
      .catch(() => {});
  }, [
    state,
    rawAppConfig,
    initialState,
    updateAppAndEmailSmsTemplatesConfig,
    refetchAppAndEmailSmsTemplatesConfig,
  ]);

  const onSmsTemplateChange = useCallback((_event, value?: string) => {
    if (value === undefined) {
      return;
    }
    setState((state) => ({
      ...state,
      smsTemplate: value,
    }));
  }, []);

  return (
    <div className={styles.form}>
      <Label className={styles.boldLabel}>
        <FormattedMessage id="PasswordsScreen.forgot-password.email.label" />
      </Label>

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorScreen.email.html-email.label" />
      </Label>
      <CodeEditor
        className={styles.htmlCodeEditor}
        language="html"
        value={state.emailHtmlTemplate}
        onChange={onEmailHtmlTemplateChange}
      />

      <Label className={styles.label}>
        <FormattedMessage id="PasswordlessAuthenticatorScreen.email.plaintext-email.label" />
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
        <FormattedMessage id="PasswordlessAuthenticatorScreen.sms.sms-body.label" />
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
          loading={updatingAppAndEmailSmsTemplateConfig}
          labelId="save"
          loadingLabelId="saving"
        />
      </div>

      <NavigationBlockerDialog blockNavigation={isFormModified} />
    </div>
  );
};

const PasswordlessAuthenticatorScreen: React.FC = function PasswordlessAuthenticatorScreen() {
  const { appID } = useParams();

  const {
    updateAppAndEmailSmsTemplatesConfig,
    loading: updatingAppAndEmailSmsTemplateConfig,
    error: updateAppAndEmailSmsTemplateConfigError,
  } = useUpdateAppAndEmailSmsTemplatesConfigMutation(
    appID,
    `templates/${EMAIL_HTML_TEMPLATE_NAME}`,
    `templates/${EMAIL_MJML_TEMPLATE_NAME}`,
    `templates/${EMAIL_TEXT_TEMPLATE_NAME}`,
    `templates/${SMS_TEXT_TEMPLATE_NAME}`
  );
  const {
    emailAndSmsTemplates,
    rawAppConfig,
    loading,
    error,
    refetch,
  } = useAppAndEmailSmsTemplatesQuery(
    appID,
    `templates/${EMAIL_HTML_TEMPLATE_NAME}`,
    `templates/${EMAIL_MJML_TEMPLATE_NAME}`,
    `templates/${EMAIL_TEXT_TEMPLATE_NAME}`,
    `templates/${SMS_TEXT_TEMPLATE_NAME}`
  );

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main
      className={cn(styles.root, {
        [styles.loading]: updatingAppAndEmailSmsTemplateConfig,
      })}
    >
      {updateAppAndEmailSmsTemplateConfigError && (
        <ShowError error={updateAppAndEmailSmsTemplateConfigError} />
      )}
      <div className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="PasswordlessAuthenticatorScreen.title" />
        </Text>
        <PasswordlessAuthenticator
          emailAndSmsTemplates={emailAndSmsTemplates}
          rawAppConfig={rawAppConfig}
          updateAppAndEmailSmsTemplatesConfig={
            updateAppAndEmailSmsTemplatesConfig
          }
          refetchAppAndEmailSmsTemplatesConfig={refetch}
          updatingAppAndEmailSmsTemplateConfig={
            updatingAppAndEmailSmsTemplateConfig
          }
        />
      </div>
    </main>
  );
};

export default PasswordlessAuthenticatorScreen;
