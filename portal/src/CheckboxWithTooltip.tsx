import React from "react";
import cn from "classnames";
import { Checkbox, ICheckboxProps } from "@fluentui/react";

import styles from "./CheckboxWithTooltip.module.css";
import Tooltip from "./Tooltip";

interface CheckboxWithTooltipProps extends ICheckboxProps {
  tooltipMessageId: string;
  tooltipMessageValues?: Record<string, any>;
}

const CheckboxWithTooltip: React.VFC<CheckboxWithTooltipProps> =
  function CheckboxWithTooltip(props: CheckboxWithTooltipProps) {
    const { tooltipMessageId, tooltipMessageValues, className, ...rest } =
      props;

    return (
      <div className={cn(styles.root, className)}>
        <Checkbox {...rest} />
        <Tooltip
          tooltipMessageId={tooltipMessageId}
          tooltipMessageValues={tooltipMessageValues}
          className={styles.tooltip}
        />
      </div>
    );
  };

export default CheckboxWithTooltip;
