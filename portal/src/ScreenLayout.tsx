import React from "react";
import ScreenHeader from "./ScreenHeader";
import ScreenNav from "./ScreenNav";
import styles from "./ScreenLayout.module.css";
import cn from "classnames";

interface ScreenLayoutProps {
  children: React.ReactElement;
}

const ScreenLayout: React.FC<ScreenLayoutProps> = function ScreenLayout(
  props: ScreenLayoutProps
) {
  return (
    <div className={cn(styles.root, "mobile:h-full")}>
      <ScreenHeader />
      <div className={styles.body}>
        <div className={cn(styles.nav, "mobile:hidden")}>
          <ScreenNav />
        </div>
        <div className={styles.content}>{props.children}</div>
      </div>
    </div>
  );
};

export default ScreenLayout;
