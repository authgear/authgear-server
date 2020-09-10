import React from "react";
import { useParams } from "react-router-dom";

import styles from "./SingleSignOnConfigurationScreen.module.scss";

const SingleSignOnConfigurationScreen: React.FC = function SingleSignOnConfigurationScreen() {
  const { appID } = useParams();

  return <main className={styles.root} role="main"></main>;
};

export default SingleSignOnConfigurationScreen;
