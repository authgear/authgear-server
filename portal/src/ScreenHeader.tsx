import React, { useCallback, useContext, useMemo } from "react";
import { Link, useParams } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import authgear from "@authgear/web";
import { Icon, IconButton, Text, Link as FluentUILink } from "@fluentui/react";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";

import styles from "./ScreenHeader.module.scss";
import { useSystemConfig } from "./context/SystemConfigContext";

interface ScreenHeaderAppSectionProps {
  appID: string;
}

const iconProps = {
  iconName: "SignOut",
};

const ScreenHeaderAppSection: React.FC<ScreenHeaderAppSectionProps> =
  function ScreenHeaderAppSection(props: ScreenHeaderAppSectionProps) {
    const { appID } = props;
    const { effectiveAppConfig, loading } = useAppAndSecretConfigQuery(appID);
    const { appHostSuffix, themes } = useSystemConfig();

    const rawAppID = effectiveAppConfig?.id;
    const endpoint =
      rawAppID != null ? "https://" + rawAppID + appHostSuffix : null;

    if (loading) {
      return null;
    }

    return (
      <>
        <Icon className={styles.headerArrow} iconName="ChevronRight" />
        {rawAppID != null && endpoint != null ? (
          <FluentUILink
            className={styles.headerAppID}
            target="_blank"
            rel="noopener"
            href={endpoint}
            theme={themes.inverted}
          >
            {`${rawAppID} - ${endpoint}`}
          </FluentUILink>
        ) : (
          <Text className={styles.headerAppID}>{appID}</Text>
        )}
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
