import React, { useCallback, useContext, useState } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import cn from "classnames";

import PasswordPolicySettings from "./PasswordPolicySettings";
import ForgotPasswordSettings from "./ForgotPasswordSettings";
import { ModifiedIndicatorWrapper } from "../../ModifiedIndicatorPortal";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import {
  UpdateAppTemplatesData,
  useUpdateAppTemplatesMutation,
} from "./mutations/updateAppTemplatesMutation";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useAppTemplatesQuery } from "./query/appTemplatesQuery";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { usePivotNavigation } from "../../hook/usePivot";
import { ForgotPasswordMessageTemplates } from "../../templates";
import { PortalAPIAppConfig } from "../../types";

import styles from "./PasswordsScreen.module.scss";

type ForgotPasswordMessageTemplateKeys = typeof ForgotPasswordMessageTemplates[number];

const PASSWORD_POLICY_PIVOT_KEY = "password_policy";
const FORGOT_PASSWORD_POLICY_KEY = "forgot_password";

const PasswordsScreen: React.FC = function PasswordsScreen() {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();

  const [remountIdentifier, setRemountIdentifier] = useState(0);

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
    resetError: resetUpdateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  const {
    updateAppTemplates,
    loading: updatingTemplates,
    error: updateTemplatesError,
    resetError: resetUpdateTemplatesError,
  } = useUpdateAppTemplatesMutation<ForgotPasswordMessageTemplateKeys>(
    appID,
    ...ForgotPasswordMessageTemplates
  );

  const resetError = useCallback(() => {
    resetUpdateAppConfigError();
    resetUpdateTemplatesError();
  }, [resetUpdateAppConfigError, resetUpdateTemplatesError]);

  const resetForm = useCallback(() => {
    setRemountIdentifier((prev) => prev + 1);
    resetError();
  }, [resetError]);

  const { selectedKey, onLinkClick } = usePivotNavigation(
    [PASSWORD_POLICY_PIVOT_KEY, FORGOT_PASSWORD_POLICY_KEY],
    resetError
  );

  const {
    loading: loadingAppConfig,
    error: loadAppConfigError,
    effectiveAppConfig,
    rawAppConfig,
    refetch,
  } = useAppConfigQuery(appID);
  const {
    templates,
    loading: loadingTemplates,
    error: loadTemplatesError,
    refetch: refetchTemplates,
  } = useAppTemplatesQuery<ForgotPasswordMessageTemplateKeys>(
    appID,
    ...ForgotPasswordMessageTemplates
  );

  const updateAppConfigAndRemountChildren = useCallback(
    async (appConfig: PortalAPIAppConfig) => {
      const app = await updateAppConfig(appConfig);
      setRemountIdentifier((prev) => prev + 1);
      return app;
    },
    [updateAppConfig]
  );

  const updateAppConfigAndTemplatesAndRemountChildren = useCallback(
    async (
      appConfig: PortalAPIAppConfig,
      updateTemplatesData: UpdateAppTemplatesData<
        ForgotPasswordMessageTemplateKeys
      >
    ) => {
      await updateAppConfig(appConfig);
      const app = await updateAppTemplates(updateTemplatesData);
      setRemountIdentifier((prev) => prev + 1);
      return app;
    },
    [updateAppConfig, updateAppTemplates]
  );

  if (loadingAppConfig || loadingTemplates) {
    return <ShowLoading />;
  }

  if (loadAppConfigError != null) {
    return <ShowError error={loadAppConfigError} onRetry={refetch} />;
  }
  if (loadTemplatesError != null) {
    return <ShowError error={loadTemplatesError} onRetry={refetchTemplates} />;
  }

  return (
    <main
      className={cn(styles.root, {
        [styles.loading]: updatingAppConfig || updatingTemplates,
      })}
    >
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      {updateTemplatesError && <ShowError error={updateTemplatesError} />}
      <ModifiedIndicatorWrapper className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="PasswordsScreen.title" />
        </Text>
        <div className={styles.tabsContainer}>
          <Pivot onLinkClick={onLinkClick} selectedKey={selectedKey}>
            <PivotItem
              headerText={renderToString(
                "PasswordsScreen.password-policy.title"
              )}
              itemKey={PASSWORD_POLICY_PIVOT_KEY}
            >
              <PasswordPolicySettings
                key={remountIdentifier}
                effectiveAppConfig={effectiveAppConfig}
                rawAppConfig={rawAppConfig}
                updateAppConfig={updateAppConfigAndRemountChildren}
                updatingAppConfig={updatingAppConfig}
                resetForm={resetForm}
              />
            </PivotItem>
            <PivotItem
              headerText={renderToString(
                "PasswordsScreen.forgot-password.title"
              )}
              itemKey={FORGOT_PASSWORD_POLICY_KEY}
            >
              <ForgotPasswordSettings
                key={remountIdentifier}
                effectiveAppConfig={effectiveAppConfig}
                rawAppConfig={rawAppConfig}
                templates={templates}
                updateAppConfigAndTemplates={
                  updateAppConfigAndTemplatesAndRemountChildren
                }
                updatingAppConfigAndTemplates={
                  updatingAppConfig || updatingTemplates
                }
                resetForm={resetForm}
              />
            </PivotItem>
          </Pivot>
        </div>
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default PasswordsScreen;
