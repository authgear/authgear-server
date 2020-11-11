import React, { useCallback, useContext, useState } from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { ModifiedIndicatorWrapper } from "../../ModifiedIndicatorPortal";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ForgotPasswordTemplatesSettings from "./ForgotPasswordTemplatesSettings";
import PasswordlessAuthenticatorTemplatesSettings from "./PasswordlessAuthenticatorTemplatesSettings";
import { useAppTemplatesQuery } from "./query/appTemplatesQuery";
import {
  UpdateAppTemplatesData,
  useUpdateAppTemplatesMutation,
} from "./mutations/updateAppTemplatesMutation";
import { usePivotNavigation } from "../../hook/usePivot";
import {
  AuthenticatePrimaryOOBMessageTemplatePaths,
  DEFAULT_TEMPLATE_LOCALE,
  ForgotPasswordMessageTemplatePaths,
  SetupPrimaryOOBMessageTemplatePaths,
} from "../../templates";

import styles from "./TemplatesConfigurationScreen.module.scss";

const FORGOT_PASSWORD_PIVOT_KEY = "forgot_password";
const PASSWORDLESS_AUTHENTICATOR_PIVOT_KEY = "passwordless_authenticator";

const TemplatesConfigurationScreen: React.FC = function TemplatesConfigurationScreen() {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();

  const [remountIdentifier, setRemountIdentifier] = useState(0);
  const [templateLocale] = useState(DEFAULT_TEMPLATE_LOCALE);

  const {
    templates,
    loading: loadingTemplates,
    error: loadTemplatesError,
    refetch: refetchTemplates,
  } = useAppTemplatesQuery(
    appID,
    templateLocale,
    ...ForgotPasswordMessageTemplatePaths,
    ...SetupPrimaryOOBMessageTemplatePaths,
    ...AuthenticatePrimaryOOBMessageTemplatePaths
  );

  const {
    updateAppTemplates,
    loading: updatingTemplates,
    error: updateTemplatesError,
    resetError: resetUpdateTemplatesError,
  } = useUpdateAppTemplatesMutation(
    appID,
    templateLocale,
    ...ForgotPasswordMessageTemplatePaths,
    ...SetupPrimaryOOBMessageTemplatePaths,
    ...AuthenticatePrimaryOOBMessageTemplatePaths
  );

  const resetError = useCallback(() => {
    resetUpdateTemplatesError();
  }, [resetUpdateTemplatesError]);

  const resetForm = useCallback(() => {
    setRemountIdentifier((prev) => prev + 1);
    resetError();
  }, [resetError]);

  const { selectedKey, onLinkClick } = usePivotNavigation(
    [FORGOT_PASSWORD_PIVOT_KEY, PASSWORDLESS_AUTHENTICATOR_PIVOT_KEY],
    resetError
  );

  const updateTemplatesAndRemountChildren = useCallback(
    async (updateTemplatesData: UpdateAppTemplatesData) => {
      const app = await updateAppTemplates(updateTemplatesData);
      setRemountIdentifier((prev) => prev + 1);
      return app;
    },
    [updateAppTemplates]
  );

  if (loadingTemplates) {
    return <ShowLoading />;
  }

  if (loadTemplatesError != null) {
    return <ShowError error={loadTemplatesError} onRetry={refetchTemplates} />;
  }

  return (
    <main
      className={cn(styles.root, {
        [styles.loading]: updatingTemplates,
      })}
    >
      {updateTemplatesError && <ShowError error={updateTemplatesError} />}
      <ModifiedIndicatorWrapper className={styles.screen}>
        <Text className={styles.screenHeaderText} as="h1">
          <FormattedMessage id="TemplatesConfigurationScreen.title" />
        </Text>
        <Pivot onLinkClick={onLinkClick} selectedKey={selectedKey}>
          <PivotItem
            headerText={renderToString(
              "TemplatesConfigurationScreen.forgot-password.title"
            )}
            itemKey={FORGOT_PASSWORD_PIVOT_KEY}
          >
            <ForgotPasswordTemplatesSettings
              key={remountIdentifier}
              templates={templates}
              templateLocale={templateLocale}
              updateTemplates={updateTemplatesAndRemountChildren}
              updatingTemplates={updatingTemplates}
              resetForm={resetForm}
            />
          </PivotItem>
          <PivotItem
            headerText={renderToString(
              "TemplatesConfigurationScreen.passwordless-authenticator.title"
            )}
            itemKey={PASSWORDLESS_AUTHENTICATOR_PIVOT_KEY}
          >
            <PasswordlessAuthenticatorTemplatesSettings
              key={remountIdentifier}
              templates={templates}
              templateLocale={templateLocale}
              updateTemplates={updateTemplatesAndRemountChildren}
              updatingTemplates={updatingTemplates}
              resetForm={resetForm}
            />
          </PivotItem>
        </Pivot>
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default TemplatesConfigurationScreen;
