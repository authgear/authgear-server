import React from "react";
import { Nav, INavLinkGroup } from "@fluentui/react";
import ScreenHeader from "./ScreenHeader";
import styles from "./ScreenLayout.module.scss";

const navGroups: INavLinkGroup[] = [
  {
    links: [
      {
        name: "Home",
        url: "https://example.com",
      },
      {
        name: "Home",
        url: "https://example.com",
      },
      {
        name: "Home",
        url: "https://example.com",
      },
      {
        name: "Home",
        url: "https://example.com",
      },
      {
        name: "Home",
        url: "https://example.com",
      },
      {
        name: "Home",
        url: "https://example.com",
      },
    ],
  },
];

interface ScreenLayoutProps {
  children: React.ReactElement;
}

const ScreenLayout: React.FC<ScreenLayoutProps> = function ScreenLayout(
  props: ScreenLayoutProps
) {
  return (
    <div className={styles.root}>
      <ScreenHeader />
      <div className={styles.body}>
        <div className={styles.nav}>
          <Nav groups={navGroups} />
        </div>
        <div className={styles.content}>{props.children}</div>
      </div>
    </div>
  );
};

export default ScreenLayout;
