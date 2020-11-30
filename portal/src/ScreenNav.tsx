import React, { useCallback, useContext, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import { IIconProps, INavLink, INavLinkGroup, Nav } from "@fluentui/react";
import { Location } from "history";
import styles from "./ScreenNav.module.scss";

function getAppRouterPath(location: Location) {
  // app router -> /app/:appID/*
  // discard first 3 segment (include leading slash)
  return location.pathname.split("/").slice(3).join("/");
}

function getPath(url: string) {
  // remove fragment
  const path = new URL("scheme:" + url).pathname;
  // remove leading trailing slash
  const pathWithoutLeadingTrailingSlash = path
    .replace(/^\//, "")
    .replace(/\/$/, "");
  return pathWithoutLeadingTrailingSlash;
}

function isPathSame(url1: string, url2: string) {
  const path1 = getPath(url1);
  const path2 = getPath(url2);
  return path1 === path2;
}

function icon(iconName: string): IIconProps {
  return {
    className: styles.icon,
    iconName,
  };
}

const ScreenNav: React.FC = function ScreenNav() {
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);
  const location = useLocation();

  const label = renderToString("ScreenNav.label");
  const [expandState, setExpandState] = useState<Record<string, boolean>>({});

  const navGroups: INavLinkGroup[] = useMemo(
    () => [
      {
        links: [
          {
            key: "users",
            name: renderToString("ScreenNav.users"),
            url: "users",
            iconProps: icon("People"),
          },
          {
            key: "authentication",
            name: renderToString("ScreenNav.authentication"),
            url: "configuration/authentication",
            iconProps: icon("Shield"),
          },
          {
            key: "anonymousUsers",
            name: renderToString("ScreenNav.anonymous-users"),
            url: "configuration/anonymous-users",
            iconProps: icon("Color"),
          },
          {
            key: "singleSignOn",
            name: renderToString("ScreenNav.single-sign-on"),
            url: "configuration/single-sign-on",
            iconProps: icon("PlugConnected"),
          },
          {
            key: "passwords",
            name: renderToString("ScreenNav.passwords"),
            url: "configuration/passwords",
            iconProps: icon("PasswordField"),
          },
          {
            key: "UserInterface",
            name: renderToString("ScreenNav.user-interface"),
            url: "configuration/user-interface",
            iconProps: icon("PreviewLink"),
          },
          {
            key: "clientApplications",
            name: renderToString("ScreenNav.client-applications"),
            url: "configuration/oauth-clients",
            iconProps: icon("Devices3"),
          },
          {
            key: "dns",
            name: renderToString("ScreenNav.dns"),
            url: "configuration/dns",
            iconProps: icon("ServerProcesses"),
          },
          {
            key: "templates",
            name: renderToString("ScreenNav.localization-appearance"),
            url: "configuration/localization-appearance",
            iconProps: icon("WebTemplate"),
          },
          {
            key: "settings",
            name: renderToString("ScreenNav.settings"),
            iconProps: icon("Settings"),
            url: "",
            isExpanded: expandState["settings"],
            links: [
              {
                key: "admins",
                name: renderToString("PortalAdminSettings.title"),
                url: "configuration/settings/portal-admins",
              },
              {
                key: "sessions",
                name: renderToString("SessionSettings.title"),
                url: "configuration/settings/sessions",
              },
              {
                key: "web-hooks",
                name: renderToString("HooksSettings.title"),
                url: "configuration/settings/web-hooks",
              },
            ],
          },
        ],
      },
    ],
    [renderToString, expandState]
  );

  const onLinkClick = useCallback(
    (e?: React.MouseEvent, item?: INavLink) => {
      e?.stopPropagation();
      e?.preventDefault();

      const path = getAppRouterPath(location);
      if (
        item != null &&
        (item.links?.length ?? 0) === 0 &&
        !isPathSame(item.url, path)
      ) {
        navigate(item.url);
      }
    },
    [navigate, location]
  );
  const onLinkExpandClick = useCallback(
    (e?: React.MouseEvent, item?: INavLink) => {
      e?.stopPropagation();
      e?.preventDefault();
      const key = item!.key!;
      setExpandState((s) => ({ ...s, [key]: !Boolean(s[key]) }));
    },
    []
  );

  const allLinks = useMemo(() => {
    const links: INavLink[] = [];
    const populateLinks = (link: INavLink) => {
      links.push(link);
      for (const child of link.links ?? []) {
        populateLinks(child);
      }
    };
    for (const group of navGroups) {
      for (const link of group.links) {
        populateLinks(link);
      }
    }
    return links;
  }, [navGroups]);

  const path = getAppRouterPath(location);
  const selectedKey = useMemo(() => {
    let matchLength = 0;
    let matchLink: INavLink | null = null;
    for (const link of allLinks) {
      if (path.startsWith(link.url) && link.url.length > matchLength) {
        matchLink = link;
        matchLength = link.url.length;
      }
    }
    return matchLink?.key;
  }, [path, allLinks]);

  return (
    <Nav
      ariaLabel={label}
      groups={navGroups}
      onLinkClick={onLinkClick}
      onLinkExpandClick={onLinkExpandClick}
      selectedKey={selectedKey}
    />
  );
};

export default ScreenNav;
