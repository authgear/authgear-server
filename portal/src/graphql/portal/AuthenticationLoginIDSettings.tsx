import React from "react";

import styles from "./AuthenticationLoginIDSettings.module.scss";

interface Props {
  appConfig: Record<string, unknown> | null;
}

const AuthenticationLoginIDSettings: React.FC<Props> = function AuthenticationLoginIDSettings(
  props: Props
) {
  console.log(props.appConfig);
  return <div className={styles.root}></div>;
};

export default AuthenticationLoginIDSettings;
