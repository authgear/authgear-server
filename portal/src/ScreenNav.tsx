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
import { useSystemConfig } from "./context/SystemConfigContext";
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

const ScreenNav: React.FC = function ScreenNav() {
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);
  const location = useLocation();
  const path = getAppRouterPath(location);

  const { auditLogEnabled } = useSystemConfig();

  const label = renderToString("ScreenNav.label");
  const [expandState, setExpandState] = useState<Record<string, boolean>>({});

  const links: NavLinkProps[] = useMemo(() => {
    const links = [
      { textKey: "ScreenNav.users", url: "users" },
      ...(auditLogEnabled
        ? [{ textKey: "ScreenNav.audit-log", url: "audit-log" }]
        : []),
      {
        textKey: "ScreenNav.authentication",
        url: "configuration/authentication",
        children: [
          {
            textKey: "ScreenNav.login-id",
            url: "configuration/authentication/login-id",
          },
          {
            textKey: "ScreenNav.authenticators",
            url: "configuration/authentication/authenticators",
          },
          {
            textKey: "ScreenNav.verification",
            url: "configuration/authentication/verification",
          },
        ],
      },
      {
        textKey: "ScreenNav.anonymous-users",
        url: "configuration/anonymous-users",
      },
      {
        textKey: "ScreenNav.biometric",
        url: "configuration/biometric",
      },
      {
        textKey: "ScreenNav.single-sign-on",
        url: "configuration/single-sign-on",
      },
      {
        textKey: "ScreenNav.password-policy",
        url: "configuration/passwords-policy",
      },
      {
        textKey: "ScreenNav.client-applications",
        url: "configuration/apps",
      },
      {
        textKey: "CustomDomainListScreen.title",
        url: "configuration/custom-domains",
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
        textKey: "ScreenNav.billing",
        url: "configuration/billing",
      },
      {
        textKey: "ScreenNav.advanced",
        url: "configuration/advanced",
        children: [
          {
            textKey: "ScreenNav.sessions",
            url: "configuration/advanced/sessions",
          },
          {
            textKey: "ScreenNav.password-reset-code",
            url: "configuration/advanced/password-reset-code",
          },
          {
            textKey: "ScreenNav.webhooks",
            url: "configuration/advanced/web-hooks",
          },
          {
            textKey: "ScreenNav.admin-api",
            url: "configuration/advanced/admin-api",
          },
        ],
      },
      {
        textKey: "PortalAdminSettings.title",
        url: "configuration/portal-admins",
      },
    ];

    return links;
  }, [auditLogEnabled]);

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
  }, [path, links]);

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
    [navItem, links]
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
