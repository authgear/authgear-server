import React from "react";
import cn from "classnames";
import {
  TooltipHost,
  ITooltipProps,
  Icon,
  DirectionalHint,
} from "@fluentui/react";

import styles from "./Tooltip.module.scss";
import { FormattedMessage } from "@oursky/react-messageformat";

interface TooltipProps {
  className?: string;
  tooltipMessageId: string;
}

const Tooltip: React.FC<TooltipProps> = function Tooltip(props: TooltipProps) {
  const { className, tooltipMessageId } = props;
  const tooltipProps: ITooltipProps = React.useMemo(() => {
    return {
      onRenderContent: () => (
        <div className={styles.tooltip}>
          <span>
            <FormattedMessage id={tooltipMessageId} />
          </span>
        </div>
      ),
    };
  }, [tooltipMessageId]);

  return (
    <div className={cn(className, styles.root)}>
      <TooltipHost
        tooltipProps={tooltipProps}
        directionalHint={DirectionalHint.bottomCenter}
      >
        <Icon className={styles.infoIcon} iconName={"info"} />
      </TooltipHost>
    </div>
  );
};

export default Tooltip;
