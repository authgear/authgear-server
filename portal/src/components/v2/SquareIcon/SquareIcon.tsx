import React, { useMemo } from "react";
import cn from "classnames";
import styles from "./SquareIcon.module.css";

// This inteface is from @radix-ui/react-icons, but it is not exported
// So I copy it once here
interface IconProps extends React.SVGAttributes<SVGElement> {
  children?: never;
  color?: string;
}

type SqaureIconSize = "7";
type SqaureIconRadius = "3" | "4";

export interface SquareIconProps {
  className?: string;
  Icon: React.ExoticComponent<IconProps>;
  size?: SqaureIconSize;
  radius?: SqaureIconRadius;
  iconSize?: string;
  backgroundColor?: string;
}

function toSizeClass(size: SqaureIconSize) {
  switch (size) {
    // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
    case "7":
      return styles["squareIcon--size7"];
  }
}

function toRadiusClass(radius: SqaureIconRadius) {
  switch (radius) {
    case "3":
      return styles["squareIcon--radius3"];
    case "4":
      return styles["squareIcon--radius4"];
  }
}

export function SquareIcon({
  className,
  Icon,
  size,
  radius,
  iconSize,
  backgroundColor,
}: SquareIconProps): React.ReactElement {
  const elStyle = useMemo<React.CSSProperties>(() => {
    const style: Record<string, unknown> = {};
    if (backgroundColor != null) {
      style["--square-icon__background-color"] = backgroundColor;
    }
    if (iconSize != null) {
      style["--square-icon__icon-size"] = iconSize;
    }
    return style;
  }, [backgroundColor, iconSize]);

  return (
    <div
      className={cn(
        styles.squareIcon,
        size != null ? toSizeClass(size) : null,
        radius != null ? toRadiusClass(radius) : null,
        className
      )}
      style={elStyle}
    >
      <Icon className={styles.squareIcon__icon} />
    </div>
  );
}
