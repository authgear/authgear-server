import React, { useState, useCallback, useRef, useMemo } from "react";
import { IconButton, DefaultEffects } from "@fluentui/react";
import cn from "classnames";

import styles from "./Widget.module.css";

interface WidgetProps {
  className?: string;
  contentLayout?: "flex-column" | "grid";
  children?: React.ReactNode;
  extended?: boolean;
  showToggleButton?: boolean;
  toggleButtonDisabled?: boolean;
  onToggleButtonClick?: () => void;
}

const ICON_PROPS = {
  iconName: "ChevronDown",
};

// 16px top padding + 32px icon button height + 16px bottom padding
// WidgetTitle's line height is related to this value.
const COLLAPSED_HEIGHT = 64;

const Widget: React.VFC<WidgetProps> = function Widget(props: WidgetProps) {
  const {
    className,
    contentLayout = "flex-column",
    children,
    extended = true,
    showToggleButton = false,
    toggleButtonDisabled = false,
    onToggleButtonClick,
  } = props;
  const [measuredHeight, setMeasureHeight] = useState<number | null>(null);
  const divRef = useRef<HTMLDivElement | null>(null);

  const onClick = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      // We used to measure the height with useLayoutEffect.
      // However, it could happen that the height is updated after useLayoutEffect.
      // Therefore we set measured height just before we expand/collapse the widget.
      if (divRef.current instanceof HTMLDivElement) {
        setMeasureHeight(divRef.current.getBoundingClientRect().height);
      }
      onToggleButtonClick?.();
    },
    [onToggleButtonClick]
  );

  const buttonStyles = useMemo(() => {
    return {
      icon: {
        // https://tailwindcss.com/docs/transition-property
        transitionProperty: "transform",
        transitionTimingFunction: "cubic-bezier(0.4, 0, 0.2, 1)",
        transitionDuration: "150ms",
        transform: extended ? "rotate(-180deg)" : undefined,
      },
    };
  }, [extended]);

  return (
    <div
      className={cn(className, styles.root)}
      style={{
        boxShadow: DefaultEffects.elevation4,
        // When the widget is expanded,
        // it does not matter if max-height is unset or set to measured height.
        // Either one would lead to a visually expanded widget.
        // When the widget is collapsed,
        // we set max-height to a constant, so flash of unstyled content is impossible.
        maxHeight: extended
          ? measuredHeight == null
            ? undefined
            : `${measuredHeight}px`
          : `${COLLAPSED_HEIGHT}px`,
      }}
    >
      {/* The height of this div is stable. It will not change during expand/collapse */}
      <div
        ref={divRef}
        className={
          contentLayout === "flex-column"
            ? styles.contentFlexColumn
            : styles.contentGrid
        }
      >
        {children}
      </div>
      <IconButton
        styles={buttonStyles}
        className={cn(styles.button, showToggleButton ? "" : styles.hide)}
        onClick={onClick}
        disabled={toggleButtonDisabled}
        iconProps={ICON_PROPS}
      />
    </div>
  );
};

export default Widget;
