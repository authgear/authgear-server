import React from "react";
import cn from "classnames";
import { Checkbox, ICheckboxProps } from "@fluentui/react";

import styles from "./CheckboxWithTooltip.module.scss";
import Tooltip from "./Tooltip";

interface CheckboxWithTooltipProps extends ICheckboxProps {
  helpText: string;
}

const CheckboxWithTooltip: React.FC<CheckboxWithTooltipProps> = function CheckboxWithTooltip(
  props: CheckboxWithTooltipProps
) {
  const { helpText, className, ...rest } = props;

  return (
    <div className={cn(styles.root, className)}>
      <Checkbox {...rest} />
      <Tooltip helpText={helpText} className={styles.tooltip} />
    </div>
  );
};

export default CheckboxWithTooltip;
