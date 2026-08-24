import React from "react";
import { Text } from "@fluentui/react";
import { FormattedMessage } from "../../intl";
import styles from "./DynamicClientsEmptyView.module.css";

export const DynamicClientsEmptyView: React.VFC =
  function DynamicClientsEmptyView() {
    return (
      <div className={styles.container}>
        <Text
          variant="mediumPlus"
          className={styles.title}
          block={true}
          styles={{ root: { fontWeight: 600, color: "var(--gray-12)" } }}
        >
          <FormattedMessage id="DynamicClientsEmptyView.empty.title" />
        </Text>
        <Text
          variant="medium"
          className={styles.description}
          block={true}
          styles={{ root: { color: "var(--gray-11)" } }}
        >
          <FormattedMessage id="DynamicClientsEmptyView.empty.description" />
        </Text>
      </div>
    );
  };
