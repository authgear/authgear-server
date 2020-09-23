import React, { useMemo, useContext } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage, Context } from "@oursky/react-messageformat";
import { Pivot, PivotItem, Text } from "@fluentui/react";
import cn from "classnames";

import PasswordPolicySettings from "./PasswordPolicySettings";
import { useAppConfigQuery } from "./query/appConfigQuery";
import { useUpdateAppConfigMutation } from "./mutations/updateAppConfigMutation";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";

import styles from "./PasswordsScreen.module.scss";

const PasswordsScreen: React.FC = function PasswordsScreen() {
  const { renderToString } = useContext(Context);
  const { appID } = useParams();

  const {
    updateAppConfig,
    loading: updatingAppConfig,
    error: updateAppConfigError,
  } = useUpdateAppConfigMutation(appID);
  const { loading, error, data, refetch } = useAppConfigQuery(appID);
  const { effectiveAppConfig, rawAppConfig } = useMemo(() => {
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
          <FormattedMessage id="PasswordsScreen.title" />
        </Text>
        <div className={styles.tabsContainer}>
          <Pivot>
            <PivotItem
              alwaysRender={true}
              headerText={renderToString("PasswordsScreen.password-policy.title")}
            >
              <PasswordPolicySettings
                effectiveAppConfig={effectiveAppConfig}
                rawAppConfig={rawAppConfig}
                updateAppConfig={updateAppConfig}
                updatingAppConfig={updatingAppConfig}
              />
            </PivotItem>
            <PivotItem
              alwaysRender={true}
              headerText={renderToString("PasswordsScreen.forgot-password.title")}
            />
          </Pivot>
        </div>
      </div>
    </main>
  );
};

export default PasswordsScreen;
