import React, { useCallback, useContext, useMemo } from "react";
import { Link, useParams } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import authgear from "@authgear/web";
import {
  Icon,
  Text,
  Link as FluentUILink,
  CommandButton,
} from "@fluentui/react";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";
import { useViewerQuery } from "./graphql/portal/query/viewerQuery";

import styles from "./ScreenHeader.module.scss";
import { useSystemConfig } from "./context/SystemConfigContext";

interface ScreenHeaderAppSectionProps {
  appID: string;
}

const ScreenHeaderAppSection: React.FC<ScreenHeaderAppSectionProps> =
  function ScreenHeaderAppSection(props: ScreenHeaderAppSectionProps) {
    const { appID } = props;
    const { effectiveAppConfig, loading } = useAppAndSecretConfigQuery(appID);
    const { themes } = useSystemConfig();

    if (loading) {
      return null;
    }

    const rawAppID = effectiveAppConfig?.id;
    const endpoint = effectiveAppConfig?.http?.public_origin;

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

const commandButtonStyles = {
  label: {
    fontSize: "12px",
  },
  menuIcon: {
    fontSize: "12px",
    color: "white",
  },
};

const ScreenHeader: React.FC = function ScreenHeader() {
  const { renderToString } = useContext(Context);
  const { themes, authgearEndpoint } = useSystemConfig();
  const { appID } = useParams();
  const { viewer } = useViewerQuery();

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

  const menuProps = useMemo(() => {
    return {
      items: [
        {
          key: "settings",
          text: renderToString("ScreenHeader.settings"),
          iconProps: {
            iconName: "PlayerSettings",
          },
          href: authgearEndpoint + "/settings",
        },
        {
          key: "logout",
          text: renderToString("ScreenHeader.sign-out"),
          iconProps: {
            iconName: "SignOut",
          },
          onClick: onClickLogout,
        },
      ],
    };
  }, [onClickLogout, renderToString, authgearEndpoint]);

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
      {viewer != null && (
        <CommandButton
          menuProps={menuProps}
          theme={themes.inverted}
          styles={commandButtonStyles}
        >
          {viewer.email}
        </CommandButton>
      )}
    </header>
  );
};

export default ScreenHeader;
