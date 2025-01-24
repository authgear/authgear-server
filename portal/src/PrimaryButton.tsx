import React from "react";
import {
  IButtonProps,
  // eslint-disable-next-line no-restricted-imports
  PrimaryButton as FluentUIPrimaryButton,
  useTheme,
} from "@fluentui/react";
import { useMergedStylesPlain } from "./util/mergeStyles";

export interface PrimaryButtonProps
  extends Omit<IButtonProps, "children" | "text"> {
  text?: React.ReactNode;
}

const PrimaryButton: React.VFC<PrimaryButtonProps> = function PrimaryButton(
  props: PrimaryButtonProps
) {
  const { styles: stylesProp, ...rest } = props;
  const theme = useTheme();
  const styles = useMergedStylesPlain(
    {
      rootDisabled: {
        color: theme.palette.neutralTertiary,
      },
      label: {
        whiteSpace: "nowrap",
      },
    },
    stylesProp
  );

  // @ts-expect-error
  return <FluentUIPrimaryButton {...rest} styles={styles} />;
};

export default PrimaryButton;
