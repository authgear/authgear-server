import React from "react";
import { useParams } from "react-router-dom";

import styles from "./AnonymousUsersConfigurationScreen.module.scss";
import { Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

const AnonymousUserConfigurationScreen: React.FC = function AnonymousUserConfigurationScreen() {
  const { appID } = useParams();

  return (
    <main className={styles.root}>
      <Text as="h1" className={styles.title}>
        <FormattedMessage id="AnonymousUsersConfigurationScreen.title" />
      </Text>
    </main>
  );
};

export default AnonymousUserConfigurationScreen;
