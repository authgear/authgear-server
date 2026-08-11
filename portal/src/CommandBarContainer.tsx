import React, { useMemo } from "react";
import { Progress } from "@radix-ui/themes";
import styles from "./CommandBarContainer.module.css";
import cn from "classnames";

export interface CommandBarContainerProps {
  className?: string;
  isLoading?: boolean;
  messageBar?: React.ReactNode;
  children?: React.ReactNode;
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
      messageBar,
      headerPosition = "sticky",
      renderHeaderContent,
    } = props;

    const defaultHeaderContent = useMemo(() => {
      // Only mount the progress bar while loading; keeping a hidden bar
      // reserves height and leaves a thin white sticky strip above the
      // page title.
      return (
        <>
          {messageBar}
          {isLoading ? <Progress size="1" radius="none" /> : null}
        </>
      );
    }, [isLoading, messageBar]);

    const isHeaderEmpty =
      messageBar == null && !isLoading && renderHeaderContent == null;

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
          data-is-scrollable="true"
        >
          {props.children}
        </div>
      </>
    );
  };

export default CommandBarContainer;
