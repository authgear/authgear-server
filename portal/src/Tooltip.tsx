import React, { useRef, useState, useCallback, useMemo } from "react";
import cn from "classnames";
import {
  TooltipHost,
  ITooltipProps,
  Icon,
  DirectionalHint,
} from "@fluentui/react";

import styles from "./Tooltip.module.css";
import { FormattedMessage } from "./intl";

interface TooltipProps {
  className?: string;
  tooltipMessageId: string;
  tooltipMessageValues?: Record<string, any>;
  isHidden?: boolean;
  children?: React.ReactNode;
}

export interface UseTooltipTargetElementResult {
  id: string;
  setRef: React.RefCallback<unknown>;
  targetElement: HTMLElement | undefined;
}

export function useTooltipTargetElement(): UseTooltipTargetElementResult {
  const { current: id } = useRef(String(Math.random()));
  const [targetElement, setTargetElement] = useState<HTMLElement | null>(null);
  const setRef = useCallback(
    (ref) => {
      if (ref == null) {
        setTargetElement(null);
      } else {
        setTargetElement(document.getElementById(id));
      }
    },
    [id, setTargetElement]
  );
  return {
    id,
    setRef,
    targetElement: targetElement ?? undefined,
  };
}

export interface TooltipIconProps {
  id?: string;
  className?: string;
  setRef?: React.RefCallback<unknown>;
}

export function TooltipIcon(props: TooltipIconProps): React.ReactElement {
  const { id, setRef, className } = props;
  return (
    <Icon
      id={id}
      className={cn(className, styles.infoIcon)}
      /* @ts-expect-error */
      ref={setRef}
      iconName="info"
    />
  );
}

const Tooltip: React.VFC<TooltipProps> = function Tooltip(props: TooltipProps) {
  const {
    className,
    tooltipMessageId,
    tooltipMessageValues,
    isHidden,
    children,
  } = props;
  const tooltipProps: ITooltipProps = React.useMemo(() => {
    return {
      // eslint-disable-next-line react/no-unstable-nested-components
      onRenderContent: () => (
        <div className={styles.tooltip}>
          <span>
            {tooltipMessageId ? (
              <FormattedMessage
                id={tooltipMessageId}
                values={tooltipMessageValues}
              />
            ) : null}
          </span>
        </div>
      ),
    };
  }, [tooltipMessageId, tooltipMessageValues]);

  return (
    <div className={cn(className, styles.root)}>
      <TooltipHost
        hostClassName={styles.host}
        tooltipProps={tooltipProps}
        calloutProps={useMemo(() => ({ hidden: isHidden }), [isHidden])}
        directionalHint={DirectionalHint.bottomCenter}
      >
        {children ? children : <TooltipIcon />}
      </TooltipHost>
    </div>
  );
};

export default Tooltip;
