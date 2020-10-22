import React, { useCallback, useContext, useMemo } from "react";
import { Link, useParams } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import authgear from "@authgear/web";
import { Icon, IconButton, Text } from "@fluentui/react";
import { useAppConfigQuery } from "./graphql/portal/query/appConfigQuery";

import styles from "./ScreenHeader.module.scss";
import { useSystemConfig } from "./context/SystemConfigContext";

interface ScreenHeaderAppSectionProps {
  appID: string;
}

const iconProps = {
  iconName: "SignOut",
};

const ScreenHeaderAppSection: React.FC<ScreenHeaderAppSectionProps> = function ScreenHeaderAppSection(
  props: ScreenHeaderAppSectionProps
) {
  const { appID } = props;
  const { effectiveAppConfig, loading } = useAppConfigQuery(appID);

  if (loading) {
    return null;
  }

  return (
    <>
      <Icon className={styles.headerArrow} iconName="ChevronRight" />
      <Text className={styles.headerAppID}>
        {/* TODO: update app name */}
        {effectiveAppConfig?.http?.public_origin ?? appID}
      </Text>
    </>
  );
};

const ScreenHeader: React.FC = function ScreenHeader() {
  const { renderToString } = useContext(Context);
  const { themes } = useSystemConfig();
  const { appID } = useParams();

  const labelSignOut = renderToString("header.sign-out");

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

  const headerStyle = useMemo(
    () => ({
      backgroundColor: themes.main.palette.themePrimary,
    }),
    [themes.main]
  );

  return (
    <header className={styles.header} style={headerStyle}>
      <div className={styles.headerLeft}>
        <Link to="/" className={styles.logoLink}>
          <img
            className={styles.logo}
            alt={renderToString("system.name")}
            src={renderToString("system.logo-uri")}
          />
        </Link>
        {appID && <ScreenHeaderAppSection appID={appID} />}
      </div>
      <IconButton
        type="button"
        iconProps={iconProps}
        onClick={onClickLogout}
        title={labelSignOut}
        ariaLabel={labelSignOut}
        theme={themes.inverted}
      />
    </header>
  );
};

export default ScreenHeader;
