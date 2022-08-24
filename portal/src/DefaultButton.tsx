import React from "react";
import {
  IButtonProps,
  // eslint-disable-next-line no-restricted-imports
  DefaultButton as FluentUIDefaultButton,
} from "@fluentui/react";

export interface DefaultButtonProps
  extends Omit<IButtonProps, "children" | "text"> {
  text?: React.ReactNode;
}

const DefaultButton: React.VFC<DefaultButtonProps> = function DefaultButton(
  props: DefaultButtonProps
) {
  // @ts-expect-error
  return <FluentUIDefaultButton {...props} />;
};

export default DefaultButton;
