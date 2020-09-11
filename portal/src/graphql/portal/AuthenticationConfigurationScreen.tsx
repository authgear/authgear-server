import React from "react";
import { useParams } from "react-router-dom";
import { Pivot, PivotItem, Text } from "@fluentui/react";

import { FormattedMessage, Context } from "@oursky/react-messageformat";

import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import AuthenticationLoginIDSettings from "./AuthenticationLoginIDSettings";
import AuthenticationAuthenticatorSettings from "./AuthenticationAuthenticatorSettings";

import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";

import styles from "./AuthenticationConfigurationScreen.module.scss";

const AuthenticationScreen: React.FC = function AuthenticationScreen() {
  const { renderToString } = React.useContext(Context);
  const { appID } = useParams();
  const { updateAppConfig } = useUpdateAppConfigMutation(appID);

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
    <main className={styles.root}>
      <div className={styles.content}>
        <Text as="h1" className={styles.title}>
          <FormattedMessage id="AuthenticationScreen.title" />
        </Text>
        <div className={styles.tabsContainer}>
          <Pivot>
            <PivotItem
              headerText={renderToString("AuthenticationScreen.login-id.title")}
            >
              <AuthenticationLoginIDSettings
                effectiveAppConfig={effectiveAppConfig}
                rawAppConfig={rawAppConfig}
                updateAppConfig={updateAppConfig}
              />
            </PivotItem>
            <PivotItem
              headerText={renderToString(
                "AuthenticationScreen.authenticator.title"
              )}
            >
              <AuthenticationAuthenticatorSettings
                effectiveAppConfig={effectiveAppConfig}
                rawAppConfig={rawAppConfig}
                updateAppConfig={updateAppConfig}
              />
            </PivotItem>
          </Pivot>
        </div>
      </div>
    </main>
  );
};

export default AuthenticationScreen;
