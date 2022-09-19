import React, { ReactNode, useContext, useMemo, forwardRef } from "react";
import { Context } from "@oursky/react-messageformat";
import {
  // eslint-disable-next-line no-restricted-imports
  Toggle as FluentUIToggle,
  IToggleProps,
  Text,
  useTheme,
} from "@fluentui/react";

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
    <div className={className} ref={ref}>
      <FluentUIToggle
        {...ownProps}
        {...rest}
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
