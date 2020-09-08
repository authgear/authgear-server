import React from "react";
import { PortalAPIAppConfig } from "../../types";

import styles from "./AuthenticationAuthenticatorSettings.module.scss";

interface Props {
  appConfig: PortalAPIAppConfig | null;
}

const AuthenticationAuthenticatorSettings: React.FC<Props> = function AuthenticationAuthenticatorSettings(
  props: Props
) {
  console.log(props.appConfig);
  return <div className={styles.root}></div>;
};

export default AuthenticationAuthenticatorSettings;
