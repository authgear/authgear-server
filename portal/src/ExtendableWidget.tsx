import React, {
  useState,
  useCallback,
  useEffect,
  useMemo,
  useContext,
} from "react";
import { IconButton, DefaultEffects } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import cn from "classnames";

import styles from "./ExtendableWidget.module.scss";

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

  const [extended, setExtended] = useState(initiallyExtended);

  const { renderToString } = useContext(Context);

  const onExtendClicked = useCallback(() => {
    const stateAftertoggle = !extended;
    setExtended(stateAftertoggle);
  }, [extended]);

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
      <div className={styles.contentContainer}>
        {extended && (
          <div className={cn(styles.content, { [styles.readOnly]: readOnly })}>
            {children}
          </div>
        )}
      </div>
    </div>
  );
};

export default ExtendableWidget;
