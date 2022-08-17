import React, { useCallback, useContext, useMemo } from "react";
import { Link, useParams } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import authgear from "@authgear/web";
import {
  Icon,
  Text,
  Link as FluentUILink,
  CommandButton,
  IconButton,
} from "@fluentui/react";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";
import { useViewerQuery } from "./graphql/portal/query/viewerQuery";

import styles from "./ScreenHeader.module.css";
import { useSystemConfig } from "./context/SystemConfigContext";

interface ScreenHeaderAppSectionProps {
  appID: string;
  mobileView?: boolean;
}

const ScreenHeaderAppSection: React.FC<ScreenHeaderAppSectionProps> =
  function ScreenHeaderAppSection(props: ScreenHeaderAppSectionProps) {
    const { appID, mobileView = false } = props;
    const { effectiveAppConfig, loading } = useAppAndSecretConfigQuery(appID);
    const { themes } = useSystemConfig();

    if (loading) {
      return null;
    }

    const rawAppID = effectiveAppConfig?.id;
    const endpoint = effectiveAppConfig?.http?.public_origin;

    return (
      <>
        {mobileView ? null : (
          <Icon className={styles.headerArrow} iconName="ChevronRight" />
        )}
        {rawAppID != null && endpoint != null ? (
          <>
            {mobileView ? (
              <Text className={styles.headerAppID} theme={themes.inverted}>
                {rawAppID}
              </Text>
            ) : (
              <FluentUILink
                className={styles.headerAppID}
                target="_blank"
                rel="noopener"
                href={endpoint}
                theme={themes.inverted}
              >
                {`${rawAppID} - ${endpoint}`}
              </FluentUILink>
            )}
          </>
        ) : (
          <Text className={styles.headerAppID} theme={themes.inverted}>
            {appID}
          </Text>
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
  const { appID } = useParams() as { appID: string };
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
      <div className="hidden mobile:flex mobile:flex-row mobile:items-center mobile:text-white">
        <IconButton
          ariaLabel="hamburger"
          iconProps={{ iconName: "WaffleOffice365" }}
          className={styles.hamburger}
          theme={themes.inverted}
        />
        {appID && <ScreenHeaderAppSection appID={appID} mobileView={true} />}
      </div>
      <div className="flex flex-row items-center text-white mobile:hidden">
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
