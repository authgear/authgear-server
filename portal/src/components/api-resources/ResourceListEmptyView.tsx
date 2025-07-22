import React, { useContext } from "react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { PrimaryButton, Text } from "@fluentui/react";
import styles from "./ResourceListEmptyView.module.css";

export const ResourceListEmptyView: React.VFC =
  function ResourceListEmptyView() {
    const { renderToString } = useContext(Context);

    return (
      <div className={styles.container}>
        <Text
          variant="mediumPlus"
          className={styles.title}
          block
          styles={{ root: { fontWeight: 600, color: "var(--gray-12)" } }}
        >
          <FormattedMessage id="ResourceListEmptyView.title" />
        </Text>
        <Text
          variant="medium"
          className={styles.description}
          block
          styles={{ root: { color: "var(--gray-11)" } }}
        >
          <FormattedMessage id="ResourceListEmptyView.description" />
        </Text>
        <PrimaryButton
          text={renderToString("ResourceListEmptyView.create-resource")}
          iconProps={{ iconName: "Add" }}
          onClick={() => {}}
        />
      </div>
    );
  };
