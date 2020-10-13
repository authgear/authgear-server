import React, { useMemo, useCallback, useContext } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import { Nav, INavLink, INavLinkGroup, INavProps } from "@fluentui/react";

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
            key: "clientApplications",
            name: renderToString("ScreenNav.client-applications"),
            url: "configuration/oauth-clients",
            icon: "AuthenticatorApp",
          },
          {
            key: "UserInterface",
            name: renderToString("ScreenNav.user-interface"),
            url: "configuration/user-interface",
            icon: "PreviewLink",
          },
        ],
      },
    ];
  }, [renderToString]);

  const onLinkClick: INavProps["onLinkClick"] = useCallback(
    (ev?: React.MouseEvent<HTMLElement>, item?: INavLink) => {
      ev?.stopPropagation();
      ev?.preventDefault();
      if (item != null) {
        navigate(item.url);
      }
    },
    [navigate]
  );

  const selectedKey = useMemo(() => {
    const linkFound = navGroups[0].links.find((link) => {
      // app router -> /app/:appID/*
      // discard first 3 segment (include leading slash)
      const appRouterPath = location.pathname.split("/").slice(3).join("/");
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
