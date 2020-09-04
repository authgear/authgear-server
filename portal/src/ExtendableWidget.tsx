import React from "react";
import { IconButton } from "@fluentui/react";
import cn from "classnames";

import styles from "./ExtendableWidget.module.scss";

interface ExtendableWidgetProps {
  HeaderComponent: React.ReactNode;
  extendable: boolean;
  children: React.ReactNode;
}

const ExtendableWidget: React.FC<ExtendableWidgetProps> = function ExtendableWidget(
  props: ExtendableWidgetProps
) {
  const { HeaderComponent } = props;

  const contentDivRef = React.createRef<HTMLDivElement>();
  const [extended, setExtended] = React.useState(false);
  const [contentHeight, setContentHeight] = React.useState("0");

  const onExtendClicked = React.useCallback(() => {
    const stateAftertoggle = !extended;
    setExtended(stateAftertoggle);
    if (!stateAftertoggle) {
      setContentHeight("0");
      return;
    }
    if (contentDivRef.current != null) {
      setContentHeight(`${contentDivRef.current.offsetHeight}px`);
      return;
    }
    setContentHeight("auto");
  }, [contentDivRef, extended]);

  React.useEffect(() => {
    if (!props.extendable && extended) {
      onExtendClicked();
    }
  }, [props.extendable, extended, onExtendClicked]);

  return (
    <div className={styles.root}>
      <div className={styles.header}>
        <div className={styles.propsHeader}>{HeaderComponent}</div>
        <IconButton
          className={cn(styles.downArrow, {
            [styles.downArrowExtended]: extended,
          })}
          onClick={onExtendClicked}
          disabled={!props.extendable}
          iconProps={{ iconName: "ChevronDown" }}
        />
      </div>
      <div
        className={styles.contentContainer}
        style={{ height: contentHeight }}
      >
        <div ref={contentDivRef} className={styles.content}>
          {props.children}
        </div>
      </div>
    </div>
  );
};

export default ExtendableWidget;
