import React, { useContext } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import cn from "classnames";

import PasswordPolicySettings from "./PasswordPolicySettings";
import ForgotPasswordSettings from "./ForgotPasswordSettings";
import { ModifiedIndicatorWrapper } from "../../ModifiedIndicatorPortal";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { useUpdateAppTemplatesMutation } from "./mutations/updateAppTemplatesMutation";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useAppTemplatesQuery } from "./query/appTemplatesQuery";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { usePivotNavigation } from "../../hook/usePivot";
import { ForgotPasswordMessageTemplates } from "../../templates";

import styles from "./PasswordsScreen.module.scss";

type ForgotPasswordMessageTemplateKeys = typeof ForgotPasswordMessageTemplates[number];

const PASSWORD_POLICY_PIVOT_KEY = "password_policy";
const FORGOT_PASSWORD_POLICY_KEY = "forgot_password";

const PasswordsScreen: React.FC = function PasswordsScreen() {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();
  const { selectedKey, onLinkClick } = usePivotNavigation([
    PASSWORD_POLICY_PIVOT_KEY,
    FORGOT_PASSWORD_POLICY_KEY,
  ]);

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);
  const {
    updateAppTemplates,
    loading: updatingTemplates,
    error: updateTemplatesError,
  } = useUpdateAppTemplatesMutation<ForgotPasswordMessageTemplateKeys>(appID);

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
  } = useAppTemplatesQuery(appID, ...ForgotPasswordMessageTemplates);

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
                effectiveAppConfig={effectiveAppConfig}
                rawAppConfig={rawAppConfig}
                updateAppConfig={updateAppConfig}
                updatingAppConfig={updatingAppConfig}
              />
            </PivotItem>
            <PivotItem
              headerText={renderToString(
                "PasswordsScreen.forgot-password.title"
              )}
              itemKey={FORGOT_PASSWORD_POLICY_KEY}
            >
              <ForgotPasswordSettings
                effectiveAppConfig={effectiveAppConfig}
                rawAppConfig={rawAppConfig}
                templates={templates}
                updateAppConfig={updateAppConfig}
                updateTemplates={updateAppTemplates}
                updatingAppConfig={updatingAppConfig}
                updatingTemplates={updatingTemplates}
              />
            </PivotItem>
          </Pivot>
        </div>
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default PasswordsScreen;
