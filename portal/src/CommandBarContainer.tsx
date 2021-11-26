import React from "react";
import {
  CommandBar,
  ICommandBarItemProps,
  ProgressIndicator,
} from "@fluentui/react";
import styles from "./CommandBarContainer.module.scss";

const progressIndicatorStyles = {
  itemProgress: {
    padding: 0,
  },
};

export interface CommandBarContainerProps {
  className?: string;
  isLoading?: boolean;
  messageBar?: React.ReactNode;
  primaryItems?: ICommandBarItemProps[];
  secondaryItems?: ICommandBarItemProps[];
}

const CommandBarContainer: React.FC<CommandBarContainerProps> =
  function CommandBarContainer(props) {
    const { className, isLoading, primaryItems, secondaryItems, messageBar } =
      props;

    return (
      <div className={className}>
        <div className={styles.header}>
          <CommandBar
            className={styles.commandBar}
            items={primaryItems ?? []}
            farItems={secondaryItems}
          />
          {isLoading && (
            <ProgressIndicator
              className={styles.progressBar}
              styles={progressIndicatorStyles}
            />
          )}
          {messageBar}
        </div>
        {props.children}
      </div>
    );
  };

export default CommandBarContainer;
