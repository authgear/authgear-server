import React, { ReactNode, useContext, useMemo, forwardRef } from "react";
import cn from "classnames";
import { Context } from "./intl";
import {
  // eslint-disable-next-line no-restricted-imports
  Toggle as FluentUIToggle,
  IToggleProps,
  Text,
  useTheme,
} from "@fluentui/react";
import { useMergedStyles } from "./util/mergeStyles";
import styles from "./Toggle.module.css";

export interface ToggleProps extends IToggleProps {
  description?: ReactNode;
  toggleClassName?: string;
}

export default forwardRef<HTMLDivElement, ToggleProps>(function Toggle(
  props,
  ref
) {
  const {
    description,
    className,
    toggleClassName,
    inlineLabel,
    disabled,
    styles: stylesProp,
    ...rest
  } = props;
  const { renderToString } = useContext(Context);
  const theme = useTheme();
  const ownProps =
    inlineLabel === false
      ? {
          onText: renderToString("Toggle.on"),
          offText: renderToString("Toggle.off"),
        }
      : undefined;

  const toggleStyles = useMergedStyles(
    {
      root: {
        marginBottom: "0",
      },
    },
    stylesProp
  );

  const textStyles = useMemo(() => {
    return {
      root: {
        lineHeight: "20px",
        color:
          disabled === true ? theme.semanticColors.disabledText : undefined,
      },
    };
  }, [disabled, theme.semanticColors.disabledText]);

  return (
    <div className={cn(className, styles.root)} ref={ref}>
      <FluentUIToggle
        {...ownProps}
        {...rest}
        styles={toggleStyles}
        inlineLabel={inlineLabel}
        className={toggleClassName}
        disabled={disabled}
      />
      {description ? (
        <Text variant="medium" block={true} styles={textStyles}>
          {description}
        </Text>
      ) : null}
    </div>
  );
});
