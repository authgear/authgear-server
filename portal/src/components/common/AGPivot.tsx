import React, { useMemo } from "react";
import {
  Pivot,
  IPivotProps,
  IPivotStyleProps,
  IPivotStyles,
} from "@fluentui/react";

const pivotStyles: IPivotProps["styles"] = {
  linkIsSelected: {
    "&::before": {
      left: 0,
      right: 0,
      bottom: -1,
      transition: "none",
    },
  },
  root: {
    lineHeight: 0,
    borderBottom: "1px solid #EDEBE9",
  },
};

export const AGPivot: React.FC<IPivotProps> = function AGPivot(props) {
  const { styles, ...rest } = props;

  const mergedStyles = useMemo(() => {
    return (styleProps: IPivotStyleProps): Partial<IPivotStyles> => {
      const overrideStyles =
        typeof styles === "function" ? styles(styleProps) : styles;

      return {
        ...pivotStyles,
        ...overrideStyles,
      };
    };
  }, [styles]);

  return <Pivot {...rest} styles={mergedStyles} />;
};
