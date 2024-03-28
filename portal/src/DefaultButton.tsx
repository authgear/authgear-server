import React, { useMemo } from "react";
import {
  IButtonProps,
  // eslint-disable-next-line no-restricted-imports
  DefaultButton as FluentUIDefaultButton,
  IButtonStyles,
} from "@fluentui/react";

export interface DefaultButtonProps
  extends Omit<IButtonProps, "children" | "text"> {
  text?: React.ReactNode;
}

const DefaultButton: React.VFC<DefaultButtonProps> = function DefaultButton(
  props: DefaultButtonProps
) {
  const { styles, ...rest } = props;

  const _styles: IButtonStyles = useMemo(
    () => ({
      root: {
        backgroundColor: "#ffffff",
      },
      ...styles,
    }),
    [styles]
  );
  // @ts-expect-error
  return <FluentUIDefaultButton {...rest} styles={_styles} />;
};

export default DefaultButton;
