import React, { useCallback } from "react";
import { Link } from "react-router-dom";
import authgear from "@authgear/web";
import { IconButton } from "@fluentui/react";
import styles from "./ScreenHeader.module.scss";
import srcLogo from "./image/screen-header-logo@3x.png";
import { invertedTheme } from "./theme";

const iconProps = {
  iconName: "SignOut",
};

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
      <Link to="/">
        <img className={styles.logo} alt="Authgear" src={srcLogo} />
      </Link>
      <IconButton
        type="button"
        iconProps={iconProps}
        onClick={onClickLogout}
        title="Logout"
        ariaLabel="Logout"
        theme={invertedTheme}
      />
    </header>
  );
};

export default ScreenHeader;
