import React from "react";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import cn from "classnames";

import { FormattedMessage, Context } from "@oursky/react-messageformat";

import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import AuthenticationLoginIDSettings from "./AuthenticationLoginIDSettings";
import AuthenticationAuthenticatorSettings from "./AuthenticationAuthenticatorSettings";

import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import { usePivot } from "../../hook/usePivot";

import styles from "./AuthenticationConfigurationScreen.module.scss";

const AuthenticationScreen: React.FC = function AuthenticationScreen() {
  const { renderToString } = React.useContext(Context);
  const { appID } = useParams();
  const { hash, onLinkClick } = usePivot();

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);

  const { loading, error, data, refetch } = useAppConfigQuery(appID);

  const { effectiveAppConfig, rawAppConfig } = React.useMemo(() => {
    const node = data?.node;
    return node?.__typename === "App"
      ? {
          effectiveAppConfig: node.effectiveAppConfig,
          rawAppConfig: node.rawAppConfig,
        }
      : {
          effectiveAppConfig: null,
          rawAppConfig: null,
        };
  }, [data]);

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
          <FormattedMessage id="AuthenticationScreen.title" />
        </Text>
        <div className={styles.tabsContainer}>
          <Pivot selectedKey={hash} onLinkClick={onLinkClick}>
            <PivotItem
              itemKey="login-id"
              headerText={renderToString("AuthenticationScreen.login-id.title")}
            >
              <AuthenticationLoginIDSettings
                effectiveAppConfig={effectiveAppConfig}
                rawAppConfig={rawAppConfig}
                updateAppConfig={updateAppConfig}
                updatingAppConfig={updatingAppConfig}
              />
            </PivotItem>
            <PivotItem
              itemKey="authenticator"
              headerText={renderToString(
                "AuthenticationScreen.authenticator.title"
              )}
            >
              <AuthenticationAuthenticatorSettings
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

export default AuthenticationScreen;
