import React from "react";
import cn from "classnames";

import styles from "./ListCellLayout.module.css";

interface ListCellLayoutProps {
  className?: string;
  children: React.ReactNode;
}

const ListCellLayout: React.VFC<ListCellLayoutProps> = function ListCellLayout(
  props: ListCellLayoutProps
) {
  return (
    <div className={cn(styles.cellContainer, props.className)}>
      {props.children}
    </div>
  );
};

export default ListCellLayout;
