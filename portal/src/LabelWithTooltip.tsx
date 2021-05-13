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
import { FormattedMessage } from "@oursky/react-messageformat";

import styles from "./LabelWithTooltip.module.scss";

interface LabelWithTooltipProps {
  className?: string;
  labelClassName?: string;
  tooltipHeaderClassName?: string;
  labelId: string;
  tooltipHeaderId: string;
  tooltipMessageId: string;
  directionalHint?: ITooltipHostProps["directionalHint"];
  required?: boolean;
  labelIIconProps?: IIconProps;
}

const LabelWithTooltip: React.FC<LabelWithTooltipProps> = function LabelWithTooltip(
  props: LabelWithTooltipProps
) {
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
  } = props;

  const tooltipProps: ITooltipProps = useMemo(() => {
    return {
      // eslint-disable-next-line react/no-unstable-nested-components
      onRenderContent: () => (
        <div className={styles.tooltip}>
          {tooltipHeaderId && (
            <Text className={cn(styles.tooltipHeader, tooltipHeaderClassName)}>
              <FormattedMessage id={tooltipHeaderId} />
            </Text>
          )}
          <Text className={styles.tooltipMessage}>
            <FormattedMessage id={tooltipMessageId} />
          </Text>
        </div>
      ),
    };
  }, [tooltipHeaderClassName, tooltipHeaderId, tooltipMessageId]);

  return (
    <div className={cn(styles.root, className)}>
      <Label className={labelClassName} required={required}>
        {labelIIconProps && (
          <Icon {...labelIIconProps} className={styles.labelIcon} />
        )}
        <FormattedMessage id={labelId} />
      </Label>
      <TooltipHost
        tooltipProps={tooltipProps}
        directionalHint={directionalHint ?? DirectionalHint.bottomCenter}
      >
        <Icon
          className={cn(styles.infoIcon, {
            [styles.infoIconRequired]: required,
          })}
          iconName={"info"}
        />
      </TooltipHost>
    </div>
  );
};

export default LabelWithTooltip;
