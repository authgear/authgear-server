import React from "react";
import cn from "classnames";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import styles from "./ResourceListEmptyView.module.css";
import { CreateResourceButton } from "./CreateResourceButton";

export interface ResourceListEmptyViewProps {
  className?: string;
  onCreateClick: () => void;
}

export const ResourceListEmptyView: React.VFC<ResourceListEmptyViewProps> =
  function ResourceListEmptyView({ className, onCreateClick }) {
    return (
      <div className={cn(styles.container, className)}>
        <Text as="p" size="4" weight="bold" className={styles.title}>
          <FormattedMessage id="ResourceListEmptyView.title" />
        </Text>
        <Text as="p" size="2" color="gray" className={styles.description}>
          <FormattedMessage id="ResourceListEmptyView.description" />
        </Text>
        <div className={styles.button}>
          <CreateResourceButton onClick={onCreateClick} />
        </div>
      </div>
    );
  };
