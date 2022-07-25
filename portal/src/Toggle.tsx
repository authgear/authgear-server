import React from "react";
import { Toggle as FluentUIToggle, IToggleProps, Text } from "@fluentui/react";

export interface ToggleProps extends IToggleProps {
  description?: string;
  toggleClassName?: string;
}

const Toggle: React.FC<ToggleProps> = function Toggle(props: ToggleProps) {
  const { description, className, toggleClassName, ...rest } = props;

  return (
    <div className={className}>
      <FluentUIToggle {...rest} className={toggleClassName} />
      {description && (
        <Text variant="medium" block={true} style={{ lineHeight: "20px" }}>
          {description}
        </Text>
      )}
    </div>
  );
};

export default Toggle;
