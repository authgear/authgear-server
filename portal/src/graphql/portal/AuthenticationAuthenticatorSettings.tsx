import React from "react";

import styles from "./AuthenticationAuthenticatorSettings.module.scss";

interface Props {
  appConfig: Record<string, unknown> | null;
}

const AuthenticationAuthenticatorSettings: React.FC<Props> = function AuthenticationAuthenticatorSettings(
  props: Props
) {
  console.log(props.appConfig);
  return <div className={styles.root}></div>;
};

export default AuthenticationAuthenticatorSettings;
