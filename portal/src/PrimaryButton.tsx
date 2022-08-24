import React, { useMemo } from "react";
import {
  IButtonProps,
  // eslint-disable-next-line no-restricted-imports
  PrimaryButton as FluentUIPrimaryButton,
  useTheme,
} from "@fluentui/react";

export interface PrimaryButtonProps extends IButtonProps {}

const PrimaryButton: React.VFC<PrimaryButtonProps> = function PrimaryButton(
  props: PrimaryButtonProps
) {
  const theme = useTheme();
  const styles = useMemo(() => {
    return {
      rootDisabled: {
        color: theme.palette.neutralTertiary,
      },
      ...props.styles,
    };
  }, [props.styles, theme.palette.neutralTertiary]);

  return <FluentUIPrimaryButton {...props} styles={styles} />;
};

export default PrimaryButton;
