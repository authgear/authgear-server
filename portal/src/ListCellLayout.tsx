import React from "react";
import cn from "classnames";

import styles from "./ListCellLayout.module.scss";

interface ListCellLayoutProps {
  className?: string;
  children: React.ReactNode;
}

const ListCellLayout: React.FC<ListCellLayoutProps> = function ListCellLayout(
  props: ListCellLayoutProps
) {
  return (
    <div className={cn(styles.cellContainer, props.className)}>
      {props.children}
    </div>
  );
};

export default ListCellLayout;
