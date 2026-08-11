import React, { useMemo } from "react";
import { CommandBar, ICommandBarItemProps } from "@fluentui/react";
import { Progress } from "@radix-ui/themes";
import styles from "./CommandBarContainer.module.css";
import cn from "classnames";

const commandBarStyles = {
  root: {
    // Align the first item with the screen title.
    padding: "0 14px",
  },
};

export interface CommandBarContainerProps {
  className?: string;
  isLoading?: boolean;
  messageBar?: React.ReactNode;
  primaryItems?: ICommandBarItemProps[];
  secondaryItems?: ICommandBarItemProps[];
  children?: React.ReactNode;
  hideCommandBar?: boolean;
  headerPosition?: "static" | "sticky";
  renderHeaderContent?: (
    defaultHeaderContent: React.ReactNode
  ) => React.ReactNode;
}

const CommandBarContainer: React.VFC<CommandBarContainerProps> =
  function CommandBarContainer(props) {
    const {
      className,
      isLoading,
      primaryItems,
      secondaryItems,
      messageBar,
      hideCommandBar,
      headerPosition = "sticky",
      renderHeaderContent,
    } = props;

    const defaultHeaderContent = useMemo(() => {
      // When the command bar is hidden, only mount the progress bar while
      // loading. Keeping a visibility:hidden indicator still reserves height and
      // leaves a thin white sticky strip above the page title.
      const progressIndicator =
        hideCommandBar === true ? (
          isLoading ? (
            <Progress size="1" radius="none" />
          ) : null
        ) : (
          <Progress
            size="1"
            radius="none"
            className={!isLoading ? styles.hidden : ""}
          />
        );

      return (
        <>
          {hideCommandBar === true ? null : (
            <CommandBar
              className={styles.commandBar}
              styles={commandBarStyles}
              items={primaryItems ?? []}
              farItems={secondaryItems}
            />
          )}
          {messageBar}
          {progressIndicator}
        </>
      );
    }, [hideCommandBar, isLoading, messageBar, primaryItems, secondaryItems]);

    const isHeaderEmpty =
      hideCommandBar === true &&
      messageBar == null &&
      !isLoading &&
      renderHeaderContent == null;

    return (
      <>
        {isHeaderEmpty ? null : (
          <div
            className={
              headerPosition === "sticky"
                ? styles.headerSticky
                : styles.headerStatic
            }
          >
            {renderHeaderContent
              ? renderHeaderContent(defaultHeaderContent)
              : defaultHeaderContent}
          </div>
        )}
        <div
          className={cn(styles.content, className)}
          // For DetailList to correctly know what to display
          // https://developer.microsoft.com/en-us/fluentui#/controls/web/detailslist
          data-is-scrollable="true"
        >
          {props.children}
        </div>
      </>
    );
  };

export default CommandBarContainer;
