import React from "react";
import {
  IButtonProps,
  // eslint-disable-next-line no-restricted-imports
  DefaultButton as FluentUIDefaultButton,
  ITheme,
} from "@fluentui/react";
import { useMergedStylesPlain } from "../../util/mergeStyles";

export interface OutlinedActionButtonProps
  extends Omit<IButtonProps, "children" | "text"> {
  text?: React.ReactNode;
  theme: ITheme;
}
const OutlinedActionButton: React.VFC<OutlinedActionButtonProps> =
  function OutlinedActionButton(props: OutlinedActionButtonProps) {
    const { theme: themeProp, styles: stylesProp, ...rest } = props;

    const borderColor = themeProp.palette.themePrimary;

    const styles = useMergedStylesPlain(
      {
        root: {
          backgroundColor: "#ffffff",
          whiteSpace: "nowrap",
          borderColor,
        },
      },
      stylesProp
    );

    return (
      // @ts-expect-error
      <FluentUIDefaultButton {...rest} styles={styles} theme={themeProp} />
    );
  };

export default OutlinedActionButton;
