import React from "react";
import cn from "classnames";

import styles from "./CheckboxWithContentLayout.module.scss";

interface CheckboxWithContentLayoutProps {
  className?: string;
  children: React.ReactNode;
}

const CheckboxWithContentLayout: React.FC<CheckboxWithContentLayoutProps> = function CheckboxWithContentLayout(
  props: CheckboxWithContentLayoutProps
) {
  const { className, children } = props;
  return (
    <div className={cn(className, styles.root)}>
      {React.Children.map(children, (child, index) => {
        if (index === 0) {
          return <div className={styles.checkbox}>{child}</div>;
        }
        return <div className={styles.content}>{child}</div>;
      })}
    </div>
  );
};

export default CheckboxWithContentLayout;
