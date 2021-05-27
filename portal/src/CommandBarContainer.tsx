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

export interface FormModel {
  updateError: unknown;
  isDirty: boolean;
  isUpdating: boolean;
  reset: () => void;
  save: () => void;
}

export interface CommandBarContainerProps {
  className?: string;
  isLoading?: boolean;
  items?: ICommandBarItemProps[];
  farItems?: ICommandBarItemProps[];
  messageBar?: React.ReactNode;
}

const CommandBarContainer: React.FC<CommandBarContainerProps> =
  function CommandBarContainer(props) {
    const { className, isLoading, items, farItems, messageBar } = props;

    return (
      <div className={className}>
        <div className={styles.header}>
          <CommandBar
            className={styles.commandBar}
            items={items ?? []}
            farItems={farItems}
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
