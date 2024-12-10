import React, { useCallback, ReactElement, ComponentType } from "react";
import {
  CompoundButton,
  IButtonProps,
  useTheme,
  IRenderFunction,
} from "@fluentui/react";
import { useMergedStylesPlain } from "./util/mergeStyles";

export interface IconComponentProps {
  disabledColor?: string;
}

export interface ChoiceButtonProps {
  className?: IButtonProps["className"];
  styles?: IButtonProps["styles"];
  checked?: IButtonProps["checked"];
  disabled?: IButtonProps["disabled"];
  text?: IButtonProps["text"];
  secondaryText?: IButtonProps["secondaryText"];
  onClick?: IButtonProps["onClick"];
  IconComponent?: ComponentType<IconComponentProps>;
}

export default function ChoiceButton(props: ChoiceButtonProps): ReactElement {
  const { IconComponent, styles: stylesProp, ...rest } = props;
  const originalTheme = useTheme();
  const styles = useMergedStylesPlain(
    {
      root: {
        maxWidth: "auto",
        // Remove minHeight so that ChoiceButton looks nice if it does not have secondaryText,
        // otherwise, it is too tall.
        minHeight: "0",
        borderColor: originalTheme.palette.neutralLight,
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
    },
    stylesProp
  );

  const onRenderIcon: IRenderFunction<IButtonProps> = useCallback(
    (props?: IButtonProps) => {
      // @ts-expect-error
      const disabledColor = props?.styles?.iconDisabled?.color;
      if (typeof disabledColor !== "string") {
        return null;
      }
      if (IconComponent == null) {
        return null;
      }

      return <IconComponent disabledColor={disabledColor} />;
    },
    [IconComponent]
  );

  return (
    <CompoundButton
      {...rest}
      toggle={true}
      styles={styles}
      onRenderIcon={IconComponent == null ? undefined : onRenderIcon}
    />
  );
}
