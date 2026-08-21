import React, { useMemo } from "react";
import styles from "./CommandBarContainer.module.css";
import cn from "classnames";

export interface CommandBarContainerProps {
  className?: string;
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
      messageBar,
      headerPosition = "sticky",
      renderHeaderContent,
    } = props;

    const defaultHeaderContent = useMemo(() => {
      return <>{messageBar}</>;
    }, [messageBar]);

    const isHeaderEmpty = messageBar == null && renderHeaderContent == null;

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
