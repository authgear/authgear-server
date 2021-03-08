import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import { INavLink, INavLinkGroup, Nav } from "@fluentui/react";
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

interface NavLinkProps {
  textKey: string;
  url: string;
  children?: Array<{
    textKey: string;
    url: string;
  }>;
}

const links: NavLinkProps[] = [
  { textKey: "ScreenNav.users", url: "users" },
  {
    textKey: "ScreenNav.authentication",
    url: "configuration/authentication",
    children: [
      {
        textKey: "AuthenticationLoginIDSettingsScreen.title.nav",
        url: "configuration/authentication/login-id",
      },
      {
        textKey: "AuthenticationAuthenticatorSettingsScreen.title.nav",
        url: "configuration/authentication/authenticators",
      },
      {
        textKey: "VerificationConfigurationScreen.title.nav",
        url: "configuration/authentication/verification",
      },
    ],
  },
  {
    textKey: "ScreenNav.anonymous-users",
    url: "configuration/anonymous-users",
  },
  {
    textKey: "ScreenNav.single-sign-on",
    url: "configuration/single-sign-on",
  },
  {
    textKey: "ScreenNav.passwords",
    url: "configuration/passwords",
    children: [
      {
        textKey: "PasswordPolicySettingsScreen.title",
        url: "configuration/passwords/policy",
      },
      {
        textKey: "ScreenNav.forgot-password",
        url: "configuration/passwords/forgot-password",
      },
    ],
  },
  {
    textKey: "ScreenNav.client-applications",
    url: "configuration/apps",
    children: [
      {
        textKey: "CORSConfigurationScreen.title",
        url: "configuration/apps/cors",
      },
      {
        textKey: "OAuthClientConfigurationScreen.title",
        url: "configuration/apps/oauth",
      },
    ],
  },
  {
    textKey: "CustomDomainListScreen.title",
    url: "configuration/dns/custom-domains",
  },
  {
    textKey: "ScreenNav.ui-settings",
    url: "configuration/ui-settings",
  },
  {
    textKey: "ScreenNav.localization",
    url: "configuration/localization",
  },
  {
    textKey: "ScreenNav.settings",
    url: "configuration/settings",
    children: [
      {
        textKey: "PortalAdminSettings.title",
        url: "configuration/settings/portal-admins",
      },
      {
        textKey: "ScreenNav.sessions",
        url: "configuration/settings/sessions",
      },
      {
        textKey: "ScreenNav.webhooks",
        url: "configuration/settings/web-hooks",
      },
    ],
  },
];

const ScreenNav: React.FC = function ScreenNav() {
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);
  const location = useLocation();
  const path = getAppRouterPath(location);

  const label = renderToString("ScreenNav.label");
  const [expandState, setExpandState] = useState<Record<string, boolean>>({});

  const [selectedKeys, selectedKey] = useMemo(() => {
    const matchedKeys: string[] = [];
    let matchLength = 0;
    let matchedKey: string | undefined;
    for (const link of links) {
      const urls = [link.url, ...(link.children ?? []).map((l) => l.url)];
      for (const url of urls) {
        if (!path.startsWith(url)) {
          continue;
        }
        matchedKeys.push(url);
        if (url.length > matchLength) {
          matchedKey = url;
          matchLength = url.length;
        }
      }
    }
    return [matchedKeys, matchedKey];
  }, [path]);

  useEffect(() => {
    for (const key of selectedKeys) {
      if (!expandState[key]) {
        setExpandState((s) => ({ ...s, [key]: true }));
      }
    }
  }, [selectedKeys, expandState]);

  const navItem = useCallback(
    (props: NavLinkProps): INavLink => {
      const children = props.children ?? [];
      let isExpanded = false;
      if (children.length > 0) {
        isExpanded = expandState[props.url] || selectedKeys.includes(props.url);
      }
      return {
        isExpanded,
        key: props.url,
        name: renderToString(props.textKey),
        url: children.length > 0 ? "" : props.url,
        links: children.map((p) => navItem(p)),
      };
    },
    [expandState, selectedKeys, renderToString]
  );

  const navGroups: INavLinkGroup[] = useMemo(
    () => [
      {
        links: links.map(navItem),
      },
    ],
    [navItem]
  );

  const onLinkClick = useCallback(
    (e?: React.MouseEvent, item?: INavLink) => {
      e?.stopPropagation();
      e?.preventDefault();

      const path = getAppRouterPath(location);
      if (item?.url && !isPathSame(item.url, path)) {
        navigate(item.url);
      }
    },
    [navigate, location]
  );
  const onLinkExpandClick = useCallback(
    (e?: React.MouseEvent, item?: INavLink) => {
      e?.stopPropagation();
      e?.preventDefault();
      const key = item?.key ?? "";
      if (selectedKeys.includes(key)) {
        return;
      }
      setExpandState((s) => ({ ...s, [key]: !Boolean(s[key]) }));
    },
    [selectedKeys]
  );

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
