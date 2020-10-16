import React from "react";
import ReactDOM from "react-dom";
import cn from "classnames";

import { ModifiedIndicator, ModifiedIndicatorProps } from "./ModifiedIndicator";

import styles from "./ModifiedIndicator.module.scss";

interface ModifiedIndicatorWrapperProps {
  className?: string;
}

const MODIFIED_INDICATOR_CONTAINER_ID = "__modified-indicator-container";

export const ModifiedIndicatorContainer: React.FC = function ModifiedIndicatorContainer() {
  return <div id={MODIFIED_INDICATOR_CONTAINER_ID} />;
};

export const ModifiedIndicatorWrapper: React.FC<ModifiedIndicatorWrapperProps> = function ModifiedIndicatorWrapper(
  props
) {
  const { className } = props;

  return (
    <div className={cn(className, styles.wrapper)}>
      <ModifiedIndicatorContainer />
      {props.children}
    </div>
  );
};

export const ModifiedIndicatorPortal: React.FC<ModifiedIndicatorProps> = function ModifiedIndicatorPortal(
  props: ModifiedIndicatorProps
) {
  const container = document.getElementById(MODIFIED_INDICATOR_CONTAINER_ID);

  // NOTE: when portal is rendered for first time, container would be null
  return container != null
    ? ReactDOM.createPortal(<ModifiedIndicator {...props} />, container)
    : null;
};
