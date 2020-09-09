import React from "react";
import cn from "classnames";
import { Checkbox, ICheckboxProps } from "@fluentui/react";

import styles from "./CheckboxWithContent.module.scss";

interface CheckboxWithContentProps extends ICheckboxProps {
  className?: string;
  children: React.ReactNode;
}

const CheckboxWithContent: React.FC<CheckboxWithContentProps> = function CheckboxWithContent(
  props: CheckboxWithContentProps
) {
  const { className, children, ...rest } = props;
  return (
    <div className={cn(className, styles.root)}>
      <Checkbox {...rest} />
      <div className={styles.content}>{children}</div>
    </div>
  );
};

export default CheckboxWithContent;
