import React from "react";
import cn from "classnames";
import { Toggle, IToggleProps } from "@fluentui/react";

import styles from "./ToggleWithContent.module.scss";

interface ToggleWithContentProps extends IToggleProps {
  className?: string;
  children: React.ReactNode;
}

const ToggleWithContent: React.FC<ToggleWithContentProps> =
  function ToggleWithContent(props: ToggleWithContentProps) {
    const { className, children, ...rest } = props;
    return (
      <div className={cn(className, styles.root)}>
        <Toggle className={styles.toggle} {...rest} />
        <div className={styles.content}>{children}</div>
      </div>
    );
  };

export default ToggleWithContent;
