import React, { useMemo } from "react";
import cn from "classnames";
import {
  DirectionalHint,
  Icon,
  IIconProps,
  ITooltipHostProps,
  ITooltipProps,
  Label,
  Text,
  TooltipHost,
} from "@fluentui/react";
import { useTooltipTargetElement } from "./Tooltip";
import { FormattedMessage } from "./intl";

import styles from "./LabelWithTooltip.module.css";

interface LabelWithTooltipProps {
  className?: string;
  labelClassName?: string;
  tooltipHeaderClassName?: string;
  labelId: string;
  tooltipMessageId: string;
  tooltipHeaderId?: string;
  directionalHint?: ITooltipHostProps["directionalHint"];
  required?: boolean;
  labelIIconProps?: IIconProps;
  labelValues?: Record<string, any>;
  tooltipValues?: Record<string, any>;
}

const LabelWithTooltip: React.VFC<LabelWithTooltipProps> =
  function LabelWithTooltip(props: LabelWithTooltipProps) {
    const {
      className,
      labelClassName,
      tooltipHeaderClassName,
      labelId,
      tooltipHeaderId,
      tooltipMessageId,
      directionalHint,
      required,
      labelIIconProps,
      labelValues,
      tooltipValues,
    } = props;

    const { id, setRef, targetElement } = useTooltipTargetElement();

    const tooltipProps: ITooltipProps = useMemo(() => {
      return {
        // eslint-disable-next-line react/no-unstable-nested-components
        onRenderContent: () => (
          <div className={styles.tooltip}>
            {tooltipHeaderId ? (
              <Text
                className={cn(styles.tooltipHeader, tooltipHeaderClassName)}
              >
                <FormattedMessage id={tooltipHeaderId} />
              </Text>
            ) : null}
            <Text className={styles.tooltipMessage}>
              <FormattedMessage id={tooltipMessageId} values={tooltipValues} />
            </Text>
          </div>
        ),
        targetElement,
      };
    }, [
      tooltipHeaderClassName,
      tooltipHeaderId,
      tooltipMessageId,
      targetElement,
      tooltipValues,
    ]);

    return (
      <div className={className}>
        <TooltipHost
          tooltipProps={tooltipProps}
          directionalHint={directionalHint ?? DirectionalHint.bottomCenter}
        >
          <div className={styles.root}>
            <Label className={labelClassName} required={required}>
              {labelIIconProps ? (
                <Icon {...labelIIconProps} className={styles.labelIcon} />
              ) : null}
              <FormattedMessage id={labelId} values={labelValues} />
            </Label>
            <Icon
              id={id}
              /* @ts-expect-error */
              ref={setRef}
              className={cn(styles.infoIcon, {
                [styles.infoIconRequired]: required,
              })}
              iconName={"info"}
            />
          </div>
        </TooltipHost>
      </div>
    );
  };

export default LabelWithTooltip;
