import React, { ReactNode, useContext, forwardRef } from "react";
import { Context } from "@oursky/react-messageformat";
// eslint-disable-next-line no-restricted-imports
import { Toggle as FluentUIToggle, IToggleProps, Text } from "@fluentui/react";

export interface ToggleProps extends IToggleProps {
  description?: ReactNode;
  toggleClassName?: string;
}

export default forwardRef<HTMLDivElement, ToggleProps>(function Toggle(
  props,
  ref
) {
  const { description, className, toggleClassName, inlineLabel, ...rest } =
    props;
  const { renderToString } = useContext(Context);
  const ownProps =
    inlineLabel === false
      ? {
          onText: renderToString("Toggle.on"),
          offText: renderToString("Toggle.off"),
        }
      : undefined;

  return (
    <div className={className} ref={ref}>
      <FluentUIToggle
        {...ownProps}
        {...rest}
        inlineLabel={inlineLabel}
        className={toggleClassName}
      />
      {description ? (
        <Text variant="medium" block={true} style={{ lineHeight: "20px" }}>
          {description}
        </Text>
      ) : null}
    </div>
  );
});
