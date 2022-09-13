import React, {
  useState,
  useCallback,
  useLayoutEffect,
  useRef,
  useMemo,
} from "react";
import { IconButton, DefaultEffects } from "@fluentui/react";
import cn from "classnames";

import styles from "./ExtendableWidget.module.css";

interface ExtendableWidgetProps {
  className?: string;
  extendable?: boolean;
  children?: React.ReactNode;
}

const ICON_PROPS = {
  iconName: "ChevronDown",
};

// 16px top padding + 32px icon button height + 16px bottom padding
const COLLAPSED_HEIGHT = 64;

const ExtendableWidget: React.VFC<ExtendableWidgetProps> =
  function ExtendableWidget(props: ExtendableWidgetProps) {
    const { className, extendable, children } = props;
    const [extended, setExtended] = useState(extendable ?? true);
    const [measuredHeight, setMeasureHeight] = useState<number | null>(null);
    const divRef = useRef<HTMLDivElement | null>(null);

    const onClick = useCallback(() => {
      setExtended((prev) => {
        return !prev;
      });
    }, []);

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
            measuredHeight == null
              ? undefined
              : extended
              ? `${measuredHeight}px`
              : `${COLLAPSED_HEIGHT}px`,
        }}
      >
        {children}
        <IconButton
          styles={buttonStyles}
          className={styles.button}
          onClick={onClick}
          disabled={extendable ?? false}
          iconProps={ICON_PROPS}
        />
      </div>
    );
  };

export default ExtendableWidget;
