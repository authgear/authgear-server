import React from "react";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import styles from "./DynamicClientsEmptyView.module.css";

export const DynamicClientsEmptyView: React.VFC =
  function DynamicClientsEmptyView() {
    return (
      <div className={styles.container}>
        <Text as="p" size="3" weight="medium" className={styles.title}>
          <FormattedMessage id="DynamicClientsEmptyView.empty.title" />
        </Text>
        <Text as="p" size="2" color="gray" className={styles.description}>
          <FormattedMessage id="DynamicClientsEmptyView.empty.description" />
        </Text>
      </div>
    );
  };
