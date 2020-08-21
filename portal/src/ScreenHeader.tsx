import React, { useCallback } from "react";
import authgear from "@authgear/web";

import styles from "./ScreenHeader.module.scss";

const ScreenHeader: React.FC = function ScreenHeader() {
  const redirectURI = window.location.origin + "/";

  const onClickLogout = useCallback(() => {
    authgear
      .logout({
        redirectURI,
      })
      .catch((err) => {
        console.error(err);
      });
  }, [redirectURI]);

  return (
    <header className={styles.header}>
      <h1 className={styles.title}>Authgear Portal</h1>
      <button
        className={styles.logoutButton}
        type="button"
        onClick={onClickLogout}
      >
        Click here to logout
      </button>
    </header>
  );
};

export default ScreenHeader;
