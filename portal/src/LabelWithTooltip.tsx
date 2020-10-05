import React, { useMemo } from "react";
import cn from "classnames";
import {
  DirectionalHint,
  Icon,
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
  } = props;

  const tooltipProps: ITooltipProps = useMemo(() => {
    return {
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
      <Label className={labelClassName}>
        <FormattedMessage id={labelId} />
      </Label>
      <TooltipHost
        tooltipProps={tooltipProps}
        directionalHint={directionalHint ?? DirectionalHint.bottomCenter}
      >
        <Icon className={styles.infoIcon} iconName={"info"} />
      </TooltipHost>
    </div>
  );
};

export default LabelWithTooltip;
