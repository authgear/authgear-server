import React from "react";
import {
  Checkbox,
  TooltipHost,
  ICheckboxProps,
  DirectionalHint,
  ITooltipProps,
  Icon,
} from "@fluentui/react";

import styles from "./CheckboxWithTooltip.module.scss";

interface CheckboxWithTooltipProps extends ICheckboxProps {
  helpText: string;
}

const CheckboxWithTooltip: React.FC<CheckboxWithTooltipProps> = function CheckboxWithTooltip(
  props: CheckboxWithTooltipProps
) {
  const { helpText, ...rest } = props;
  const tooltipProps: ITooltipProps = React.useMemo(() => {
    return {
      onRenderContent: () => (
        <div className={styles.tooltip}>
          <span>{helpText}</span>
        </div>
      ),
    };
  }, [helpText]);

  return (
    <div className={styles.root}>
      <Checkbox {...rest} />
      <TooltipHost
        tooltipProps={tooltipProps}
        directionalHint={DirectionalHint.bottomCenter}
      >
        <Icon className={styles.infoIcon} iconName={"info"} />
      </TooltipHost>
    </div>
  );
};

export default CheckboxWithTooltip;
