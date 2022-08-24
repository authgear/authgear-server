import React from "react";
import {
  IButtonProps,
  // eslint-disable-next-line no-restricted-imports
  CommandBarButton as FluentUICommandBarButton,
} from "@fluentui/react";

export interface CommandBarButtonProps
  extends Omit<IButtonProps, "children" | "text"> {
  text?: React.ReactNode;
}

const CommandBarButton: React.VFC<CommandBarButtonProps> =
  function CommandBarButton(props: CommandBarButtonProps) {
    // @ts-expect-error
    return <FluentUICommandBarButton {...props} />;
  };

export default CommandBarButton;
