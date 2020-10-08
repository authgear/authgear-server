import React, { useContext } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import cn from "classnames";

import PasswordPolicySettings from "./PasswordPolicySettings";
import ForgotPasswordSettings from "./ForgotPasswordSettings";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { usePivotNavigation } from "../../hook/usePivot";

import styles from "./PasswordsScreen.module.scss";

const PASSWORD_POLICY_PIVOT_KEY = "password_policy";
const FORGOT_PASSWORD_POLICY_KEY = "forgot_password";

const PasswordsScreen: React.FC = function PasswordsScreen() {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();
  const { selectedKey, onLinkClick } = usePivotNavigation();

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);
  const {
    loading,
    error,
    effectiveAppConfig,
    rawAppConfig,
    refetch,
  } = useAppConfigQuery(appID);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={cn(styles.root, { [styles.loading]: updatingAppConfig })}>
      {updateAppConfigError && <ShowError error={updateAppConfigError} />}
      <div className={styles.content}>
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
                updateAppConfig={updateAppConfig}
                updatingAppConfig={updatingAppConfig}
              />
            </PivotItem>
          </Pivot>
        </div>
      </div>
    </main>
  );
};

export default PasswordsScreen;
