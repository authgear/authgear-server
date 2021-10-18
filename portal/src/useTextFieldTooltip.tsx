import React, { useCallback, useRef } from "react";
import {
  TooltipHost,
  IconButton,
  Label,
  ITextFieldProps,
  IIconProps,
  IIconStyleProps,
} from "@fluentui/react";
import { useSystemConfig } from "./context/SystemConfigContext";
import styles from "./useTextFieldTooltip.module.scss";

const iconButtonStyles = {
  root: {
    margin: "0px",
    padding: "4px",
    width: "auto",
    height: "auto",
  },
};

const iconProps: IIconProps = {
  iconName: "Info",
  styles: (props: IIconStyleProps) => {
    return {
      root: {
        width: "12px",
        height: "12px",
        fontSize: "12px",
        lineHeight: "12px",
        margin: "0px",
        color: props.theme?.semanticColors.bodyText,
      },
    };
  },
};

export interface TextFieldTooltipInputProps {
  tooltipLabel?: string;
}

export interface TextFieldTooltipOutputProps {
  onRenderLabel: ITextFieldProps["onRenderLabel"];
}

export function useTextFieldTooltip(
  props: TextFieldTooltipInputProps
): TextFieldTooltipOutputProps {
  const { tooltipLabel } = props;
  const {
    themes: {
      main: {
        semanticColors: { errorText },
      },
    },
  } = useSystemConfig();

  const { current: id } = useRef(String(Math.random()));

  const onRenderLabel = useCallback(
    (props?: ITextFieldProps) => {
      if (props == null) {
        return null;
      }
      return (
        <>
          <div className={styles.labelContainer}>
            <Label>{props.label}</Label>
            <TooltipHost content={tooltipLabel} id={id}>
              <IconButton
                iconProps={iconProps}
                title={tooltipLabel}
                ariaLabel={tooltipLabel}
                styles={iconButtonStyles}
              />
            </TooltipHost>
            {props.required === true && (
              <span className={styles.required} style={{ color: errorText }} />
            )}
          </div>
        </>
      );
    },
    [tooltipLabel, id, errorText]
  );

  return {
    onRenderLabel,
  };
}
