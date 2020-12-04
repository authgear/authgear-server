import React, { useCallback, useState } from "react";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import cn from "classnames";

import { FormattedMessage, Context } from "@oursky/react-messageformat";

import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import AuthenticationAuthenticatorSettings from "./AuthenticationAuthenticatorSettings";
import { ModifiedIndicatorWrapper } from "../../ModifiedIndicatorPortal";

import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { usePivotNavigation } from "../../hook/usePivot";

import styles from "./AuthenticationConfigurationScreen.module.scss";

const LOGIN_ID_PIVOT_KEY = "login-id";
const AUTHENTICATOR_PIVOT_KEY = "authenticator";

const AuthenticationScreen: React.FC = function AuthenticationScreen() {
  const { renderToString } = React.useContext(Context);
  const { appID } = useParams();

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
    resetError: resetUpdateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  const { selectedKey, onLinkClick } = usePivotNavigation(
    [LOGIN_ID_PIVOT_KEY, AUTHENTICATOR_PIVOT_KEY],
    resetUpdateAppConfigError
  );

  const {
    loading,
    error,
    effectiveAppConfig,
    rawAppConfig,
    refetch,
  } = useAppConfigQuery(appID);

  const [remountIdentifier, setRemountIdentifier] = useState(0);

  const resetForm = useCallback(() => {
    setRemountIdentifier((prev) => prev + 1);
    resetUpdateAppConfigError();
  }, [resetUpdateAppConfigError]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <main className={cn(styles.root, { [styles.loading]: updatingAppConfig })}>
      <ModifiedIndicatorWrapper className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="AuthenticationScreen.title" />
        </Text>
        <div className={styles.tabsContainer}>
          <Pivot selectedKey={selectedKey} onLinkClick={onLinkClick}>
            <PivotItem
              itemKey={LOGIN_ID_PIVOT_KEY}
              headerText={renderToString("AuthenticationScreen.login-id.title")}
            >
              <div />
            </PivotItem>
            <PivotItem
              itemKey={AUTHENTICATOR_PIVOT_KEY}
              headerText={renderToString(
                "AuthenticationScreen.authenticator.title"
              )}
            >
              <AuthenticationAuthenticatorSettings
                key={remountIdentifier}
                effectiveAppConfig={effectiveAppConfig}
                rawAppConfig={rawAppConfig}
                updateAppConfig={updateAppConfig}
                updatingAppConfig={updatingAppConfig}
                updateAppConfigError={updateAppConfigError}
                resetForm={resetForm}
              />
            </PivotItem>
          </Pivot>
        </div>
      </ModifiedIndicatorWrapper>
    </main>
  );
};

export default AuthenticationScreen;
