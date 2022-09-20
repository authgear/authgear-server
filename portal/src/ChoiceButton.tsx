import React, { ReactNode, ReactElement } from "react";
import { CompoundButton, IButtonStyles, useTheme } from "@fluentui/react";

export interface ChoiceButtonProps {
  className?: string;
  checked?: boolean;
  disabled?: boolean;
  text?: ReactNode;
  secondaryText?: ReactNode;
}

export default function ChoiceButton(props: ChoiceButtonProps): ReactElement {
  const originalTheme = useTheme();
  const styles: IButtonStyles = {
    root: {
      maxWidth: "auto",
    },
    rootChecked: {
      borderColor: originalTheme.palette.themePrimary,
      backgroundColor: originalTheme.semanticColors.buttonBackground,
    },
  };
  // @ts-expect-error
  return <CompoundButton {...props} toggle={true} styles={styles} />;
}
