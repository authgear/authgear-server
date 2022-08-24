import React from "react";
import {
  IButtonProps,
  // eslint-disable-next-line no-restricted-imports
  ActionButton as FluentUIActionButton,
} from "@fluentui/react";

export interface ActionButtonProps
  extends Omit<IButtonProps, "children" | "text"> {
  text?: React.ReactNode;
}

const ActionButton: React.VFC<ActionButtonProps> = function ActionButton(
  props: ActionButtonProps
) {
  // @ts-expect-error
  return <FluentUIActionButton {...props} />;
};

export default ActionButton;
