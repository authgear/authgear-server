import React, {
  useState,
  useCallback,
  useLayoutEffect,
  useRef,
  useMemo,
} from "react";
import { IconButton, DefaultEffects } from "@fluentui/react";
import cn from "classnames";

import styles from "./Widget.module.css";

interface WidgetProps {
  className?: string;
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
const COLLAPSED_HEIGHT = 64;

const Widget: React.VFC<WidgetProps> = function Widget(props: WidgetProps) {
  const {
    className,
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

  useLayoutEffect(() => {
    if (divRef.current instanceof HTMLDivElement) {
      setMeasureHeight(divRef.current.clientHeight);
    }
  }, []);

  return (
    <div
      ref={divRef}
      className={cn(className, styles.root)}
      style={{
        boxShadow: DefaultEffects.elevation4,
        maxHeight:
          measuredHeight == null || !showToggleButton
            ? undefined
            : extended
            ? `${measuredHeight}px`
            : `${COLLAPSED_HEIGHT}px`,
      }}
    >
      {children}
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
