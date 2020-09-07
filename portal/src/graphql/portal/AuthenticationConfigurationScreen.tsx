import React from "react";
import { useParams } from "react-router-dom";
import { gql, useQuery } from "@apollo/client";
import { Pivot, PivotItem } from "@fluentui/react";

import { FormattedMessage, Context } from "@oursky/react-messageformat";

import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import AuthenticationLoginIDSettings from "./AuthenticationLoginIDSettings";
import AuthenticationAuthenticatorSettings from "./AuthenticationAuthenticatorSettings";

import { client } from "../portal/apollo";
import { isPortalApiApp } from "../../util/types";
import { AppConfigQuery } from "./__generated__/AppConfigQuery";

import styles from "./AuthenticationConfigurationScreen.module.scss";

const authenticationScreenQuery = gql`
  query AppConfigQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        effectiveAppConfig
      }
    }
  }
`;

const AuthenticationScreen: React.FC = function AuthenticationScreen() {
  const { renderToString } = React.useContext(Context);
  const { appID } = useParams();

  const { loading, error, data, refetch } = useQuery<AppConfigQuery>(
    authenticationScreenQuery,
    {
      client,
      variables: {
        id: appID,
      },
    }
  );

  const appConfig = React.useMemo(() => {
    const node = data?.node;
    if (!isPortalApiApp(node)) {
      return null;
    }
    return node.effectiveAppConfig ?? null;
  }, [data]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return (
    <div className={styles.root}>
      <div className={styles.content}>
        <div className={styles.title}>
          <FormattedMessage id="AuthenticationScreen.title" />
        </div>
        <div className={styles.tabsContainer}>
          <Pivot>
            <PivotItem
              headerText={renderToString("AuthenticationScreen.login-id.title")}
            >
              <AuthenticationLoginIDSettings appConfig={appConfig} />
            </PivotItem>
            <PivotItem
              headerText={renderToString(
                "AuthenticationScreen.authenticator.title"
              )}
            >
              <AuthenticationAuthenticatorSettings appConfig={appConfig} />
            </PivotItem>
          </Pivot>
        </div>
      </div>
    </div>
  );
};

export default AuthenticationScreen;
