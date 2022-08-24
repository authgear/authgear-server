import React from "react";
import ScreenHeader from "./ScreenHeader";
import ScreenNav from "./ScreenNav";
import styles from "./ScreenLayout.module.css";

interface ScreenLayoutProps {
  children: React.ReactElement;
}

const ScreenLayout: React.VFC<ScreenLayoutProps> = function ScreenLayout(
  props: ScreenLayoutProps
) {
  return (
    <div className={styles.root}>
      <ScreenHeader />
      <div className={styles.body}>
        <div className={styles.nav}>
          <ScreenNav />
        </div>
        <div className={styles.content}>{props.children}</div>
      </div>
    </div>
  );
};

export default ScreenLayout;
