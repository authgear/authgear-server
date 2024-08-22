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
  useThemePrimaryForBorderColor?: boolean;
}

const DefaultButton: React.VFC<DefaultButtonProps> = function DefaultButton(
  props: DefaultButtonProps
) {
  const { styles, useThemePrimaryForBorderColor, ...rest } = props;

  const _styles: IButtonStyles = useMemo(
    () => ({
      root: {
        backgroundColor: "#ffffff",
        borderColor: useThemePrimaryForBorderColor
          ? rest.theme?.palette.themePrimary
          : undefined,
      },
      ...styles,
    }),
    [rest.theme?.palette.themePrimary, styles, useThemePrimaryForBorderColor]
  );
  // @ts-expect-error
  return <FluentUIDefaultButton {...rest} styles={_styles} />;
};

export default DefaultButton;
