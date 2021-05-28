import React, { useMemo, useContext } from "react";
import { IconButton, DefaultEffects } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";
import cn from "classnames";

import styles from "./ExtendableWidget.module.scss";

interface ExtendableWidgetProps {
  HeaderComponent: React.ReactNode;
  extended: boolean;
  onExtendClicked: () => void;
  extendButtonDisabled: boolean;
  readOnly?: boolean;
  extendButtonAriaLabelId: string;
  children: React.ReactNode;
  className?: string;
}

const ICON_PROPS = {
  iconName: "ChevronDown",
};

const ExtendableWidget: React.FC<ExtendableWidgetProps> =
  function ExtendableWidget(props: ExtendableWidgetProps) {
    const {
      className,
      HeaderComponent,
      extended,
      onExtendClicked,
      extendButtonDisabled,
      readOnly,
      children,
      extendButtonAriaLabelId,
    } = props;

    const { renderToString } = useContext(Context);

    const buttonAriaLabel = useMemo(
      () => renderToString(extendButtonAriaLabelId),
      [extendButtonAriaLabelId, renderToString]
    );

    const buttonStyles = {
      icon: {
        transition: "transform 200ms ease",
        transform: extended ? "rotate(-180deg)" : undefined,
      },
    };

    return (
      <div
        className={className}
        style={{ boxShadow: DefaultEffects.elevation4 }}
      >
        <div className={styles.header}>
          <div className={styles.propsHeader}>{HeaderComponent}</div>
          <IconButton
            styles={buttonStyles}
            ariaLabel={buttonAriaLabel}
            onClick={onExtendClicked}
            disabled={extendButtonDisabled}
            iconProps={ICON_PROPS}
          />
        </div>
        <div className={styles.contentContainer}>
          {extended && (
            <div
              className={cn(styles.content, { [styles.readOnly]: readOnly })}
            >
              {children}
            </div>
          )}
        </div>
      </div>
    );
  };

export default ExtendableWidget;
