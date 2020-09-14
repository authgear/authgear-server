import React, {
  createRef,
  useState,
  useCallback,
  useEffect,
  useMemo,
  useContext,
} from "react";
import { IconButton, DefaultEffects } from "@fluentui/react";
import cn from "classnames";

import styles from "./ExtendableWidget.module.scss";
import { Context } from "@oursky/react-messageformat";

interface ExtendableWidgetProps {
  HeaderComponent: React.ReactNode;
  initiallyExtended: boolean;
  extendable: boolean;
  readOnly?: boolean;
  extendButtonAriaLabelId: string;
  children: React.ReactNode;
  className?: string;
}

const ICON_PROPS = {
  iconName: "ChevronDown",
};

const ExtendableWidget: React.FC<ExtendableWidgetProps> = function ExtendableWidget(
  props: ExtendableWidgetProps
) {
  const {
    className,
    HeaderComponent,
    initiallyExtended,
    extendable,
    readOnly,
    children,
    extendButtonAriaLabelId,
  } = props;

  const contentDivRef = createRef<HTMLDivElement>();
  const [extended, setExtended] = useState(initiallyExtended);
  const [contentHeight, setContentHeight] = useState(
    initiallyExtended ? "auto" : "0"
  );

  const { renderToString } = useContext(Context);

  const onExtendClicked = useCallback(() => {
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

  // Collapse when extendable becomes false.
  useEffect(() => {
    if (!extendable && extended) {
      onExtendClicked();
    }
  }, [extendable, extended, onExtendClicked]);

  const buttonAriaLabel = useMemo(
    () => renderToString(extendButtonAriaLabelId),
    [extendButtonAriaLabelId, renderToString]
  );

  return (
    <div className={className} style={{ boxShadow: DefaultEffects.elevation4 }}>
      <div className={styles.header}>
        <div className={styles.propsHeader}>{HeaderComponent}</div>
        <IconButton
          className={cn(styles.downArrow, {
            [styles.downArrowExtended]: extended,
          })}
          ariaLabel={buttonAriaLabel}
          onClick={onExtendClicked}
          disabled={!extendable}
          iconProps={ICON_PROPS}
        />
      </div>
      <div
        className={styles.contentContainer}
        style={{ height: contentHeight }}
      >
        <div
          ref={contentDivRef}
          className={cn(styles.content, { [styles.readOnly]: readOnly })}
        >
          {children}
        </div>
      </div>
    </div>
  );
};

export default ExtendableWidget;
