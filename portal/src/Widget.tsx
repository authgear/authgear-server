import React, { useState, useCallback, useRef, useMemo } from "react";
import { IconButton, DefaultEffects, IIconProps } from "@fluentui/react";
import cn from "classnames";

import styles from "./Widget.module.css";

interface WidgetProps {
  className?: string;
  contentLayout?: "flex-column" | "grid";
  children?: React.ReactNode;
  extended?: boolean;
  showElevation?: boolean;
  showToggleButton?: boolean;
  toggleButtonDisabled?: boolean;
  onToggleButtonClick?: () => void;
  collapsedLayout?: "title-only" | "title-description";
}

const ICON_PROPS: IIconProps = {
  iconName: "ChevronDown",
  className: styles.icon,
};

const COLLAPSED_HEIGHT: Record<"title-only" | "title-description", number> = {
  "title-only": 64,
  "title-description": 100,
};

const Widget: React.VFC<WidgetProps> = function Widget(props: WidgetProps) {
  const {
    className,
    contentLayout = "flex-column",
    children,
    extended = true,
    showElevation = false,
    showToggleButton = false,
    toggleButtonDisabled = false,
    onToggleButtonClick,
    collapsedLayout = "title-only",
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
        boxShadow: showElevation ? DefaultEffects.elevation4 : undefined,
        // When the widget is expanded,
        // it does not matter if max-height is unset or set to measured height.
        // Either one would lead to a visually expanded widget.
        // When the widget is collapsed,
        // we set max-height to a constant, so flash of unstyled content is impossible.
        maxHeight: extended
          ? measuredHeight == null
            ? undefined
            : `${measuredHeight}px`
          : `${COLLAPSED_HEIGHT[collapsedLayout]}px`,
      }}
    >
      {/* The height of this div is stable. It will not change during expand/collapse */}
      <div
        ref={divRef}
        className={cn(
          contentLayout === "flex-column"
            ? styles.contentFlexColumn
            : styles.contentGrid,
          showElevation && styles.contentElevated
        )}
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
