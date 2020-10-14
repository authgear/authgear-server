import React, { useMemo, useCallback, useContext } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import { Nav, INavLink, INavLinkGroup, INavProps } from "@fluentui/react";
import { Location } from "history";

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

const ScreenNav: React.FC = function ScreenNav() {
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);
  const location = useLocation();

  const label = renderToString("ScreenNav.label");

  const navGroups: INavLinkGroup[] = useMemo(() => {
    return [
      {
        links: [
          {
            key: "users",
            name: renderToString("ScreenNav.users"),
            url: "users",
            icon: "People",
          },
          {
            key: "authentication",
            name: renderToString("ScreenNav.authentication"),
            url: "configuration/authentication",
            icon: "Shield",
          },
          {
            key: "anonymousUsers",
            name: renderToString("ScreenNav.anonymous-users"),
            url: "configuration/anonymous-users",
            icon: "Color",
          },
          {
            key: "singleSignOn",
            name: renderToString("ScreenNav.single-sign-on"),
            url: "configuration/single-sign-on",
            icon: "PlugConnected",
          },
          {
            key: "passwords",
            name: renderToString("ScreenNav.passwords"),
            url: "configuration/passwords",
            icon: "PasswordField",
          },
          {
            key: "passwordlessAuthenticator",
            name: renderToString("ScreenNav.passwordless-authenticator"),
            url: "configuration/passwordless-authenticator",
            icon: "PassiveAuthentication",
          },
          {
            key: "UserInterface",
            name: renderToString("ScreenNav.user-interface"),
            url: "configuration/user-interface",
            icon: "PreviewLink",
          },
          {
            key: "clientApplications",
            name: renderToString("ScreenNav.client-applications"),
            url: "configuration/oauth-clients",
            icon: "Devices3",
          },
          {
            key: "dns",
            name: renderToString("ScreenNav.dns"),
            url: "configuration/dns",
            icon: "ServerProcesses",
          },
        ],
      },
    ];
  }, [renderToString]);

  const onLinkClick: INavProps["onLinkClick"] = useCallback(
    (ev?: React.MouseEvent<HTMLElement>, item?: INavLink) => {
      ev?.stopPropagation();
      ev?.preventDefault();

      const appRouterPath = getAppRouterPath(location);
      if (item != null && !isPathSame(item.url, appRouterPath)) {
        navigate(item.url);
      }
    },
    [navigate, location]
  );

  const selectedKey = useMemo(() => {
    const linkFound = navGroups[0].links.find((link) => {
      const appRouterPath = getAppRouterPath(location);
      return appRouterPath.startsWith(link.url);
    });
    return linkFound?.key;
  }, [location, navGroups]);

  return (
    <Nav
      ariaLabel={label}
      groups={navGroups}
      onLinkClick={onLinkClick}
      selectedKey={selectedKey}
    />
  );
};

export default ScreenNav;
