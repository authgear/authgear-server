import React, { useMemo, useCallback } from "react";
import {
  TooltipHost,
  IconButton,
  Label,
  ITextFieldProps,
  IIconProps,
  IIconStyleProps,
  ITooltipProps,
} from "@fluentui/react";
import { useSystemConfig } from "./context/SystemConfigContext";
import { useTooltipTargetElement } from "./Tooltip";
import styles from "./useTextFieldTooltip.module.css";

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

  const { id, setRef, targetElement } = useTooltipTargetElement();

  const tooltipProps: ITooltipProps = useMemo(() => {
    return {
      content: tooltipLabel,
      targetElement,
    };
  }, [tooltipLabel, targetElement]);

  const onRenderLabel = useCallback(
    (props?: ITextFieldProps) => {
      if (props == null) {
        return null;
      }
      return (
        <TooltipHost tooltipProps={tooltipProps} content={tooltipLabel}>
          <div className={styles.labelContainer}>
            <Label>{props.label}</Label>
            <IconButton
              id={id}
              ref={setRef}
              iconProps={iconProps}
              title={tooltipLabel}
              ariaLabel={tooltipLabel}
              styles={iconButtonStyles}
            />
            {props.required === true ? (
              <span className={styles.required} style={{ color: errorText }} />
            ) : null}
          </div>
        </TooltipHost>
      );
    },
    [id, setRef, tooltipProps, tooltipLabel, errorText]
  );

  return {
    onRenderLabel,
  };
}
