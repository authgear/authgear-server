import React from "react";
import { FormattedMessage } from "../../intl";

import { Text } from "@fluentui/react";
import styles from "./ResourceListEmptyView.module.css";
import { CreateResourceButton } from "./CreateResourceButton";

export const ResourceListEmptyView: React.VFC =
  function ResourceListEmptyView() {
    return (
      <div className={styles.container}>
        <Text
          variant="mediumPlus"
          className={styles.title}
          block={true}
          styles={{ root: { fontWeight: 600, color: "var(--gray-12)" } }}
        >
          <FormattedMessage id="ResourceListEmptyView.title" />
        </Text>
        <Text
          variant="medium"
          className={styles.description}
          block={true}
          styles={{ root: { color: "var(--gray-11)" } }}
        >
          <FormattedMessage id="ResourceListEmptyView.description" />
        </Text>
        <CreateResourceButton />
      </div>
    );
  };
