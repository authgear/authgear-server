import React, { ReactNode, ReactElement } from "react";
import {
  CompoundButton,
  IButtonStyles,
  IButtonProps,
  useTheme,
} from "@fluentui/react";

export interface ChoiceButtonProps {
  className?: string;
  checked?: IButtonProps["checked"];
  disabled?: IButtonProps["disabled"];
  text?: ReactNode;
  secondaryText?: ReactNode;
  onClick?: IButtonProps["onClick"];
}

export default function ChoiceButton(props: ChoiceButtonProps): ReactElement {
  const originalTheme = useTheme();
  const styles: IButtonStyles = {
    root: {
      maxWidth: "auto",
      // Remove minHeight so that ChoiceButton looks nice if it does not have secondaryText,
      // otherwise, it is too tall.
      minHeight: "0",
    },
    rootChecked: {
      // Double the border width VISUALLY to make checked ChoiceButton more prominent.
      // Note that we cannot simply double border-width because border-width is part of
      // the border-box so it affects layout.
      outlineColor: originalTheme.palette.themePrimary,
      outlineStyle: "solid",
      outlineWidth: "1px",

      borderColor: originalTheme.palette.themePrimary,
      backgroundColor: originalTheme.semanticColors.buttonBackground,
    },
    description: {
      color: "inherit",
    },
    label: {
      // Make the label center aligned when there is no secondaryText.
      margin: props.secondaryText == null ? "0" : undefined,
    },
    // When ChoiceButton is taller than its intrinsic height,
    // make sure the content is still center aligned vertically.
    flexContainer: {
      alignItems: "center",
    },
  };
  // @ts-expect-error
  return <CompoundButton {...props} toggle={true} styles={styles} />;
}
