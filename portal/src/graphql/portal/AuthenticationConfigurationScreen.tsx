import React from "react";
import { Pivot, PivotItem } from "@fluentui/react";

import { FormattedMessage, Context } from "@oursky/react-messageformat";

import AuthenticationLoginIDSettings from "./AuthenticationLoginIDSettings";
import AuthenticationAuthenticatorSettings from "./AuthenticationAuthenticatorSettings";

import styles from "./AuthenticationConfigurationScreen.module.scss";

const AuthenticationScreen: React.FC = function AuthenticationScreen() {
  const { renderToString } = React.useContext(Context);
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
              <AuthenticationLoginIDSettings />
            </PivotItem>
            <PivotItem
              headerText={renderToString(
                "AuthenticationScreen.authenticator.title"
              )}
            >
              <AuthenticationAuthenticatorSettings />
            </PivotItem>
          </Pivot>
        </div>
      </div>
    </div>
  );
};

export default AuthenticationScreen;
