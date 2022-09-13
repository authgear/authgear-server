import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { useQuery } from "@apollo/client";
import { useParams, useLocation, useNavigate } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import { INavLink, INavLinkGroup, INavStyleProps, Nav } from "@fluentui/react";
import { useSystemConfig } from "./context/SystemConfigContext";
import {
  ScreenNavQueryQuery,
  ScreenNavQueryDocument,
} from "./graphql/portal/query/screenNavQuery.generated";
import { client } from "./graphql/portal/apollo";
import { Location } from "history";

function getAppRouterPath(location: Location) {
  // app router -> /app/:appID/*
  // discard first 3 segment (include leading slash)
  return "/" + location.pathname.split("/").slice(3).join("/");
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

function getStyles(props: INavStyleProps) {
  return {
    chevronButton: {
      backgroundColor: "transparent",
    },
    chevronIcon: {
      transform: props.isExpanded ? "rotate(0deg)" : "rotate(-90deg)",
    },
  };
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

interface ScreenNavProps {
  mobileView?: boolean;
  onLinkClick?: () => void;
}

const ScreenNav: React.VFC<ScreenNavProps> = function ScreenNav(props) {
  const { mobileView = false } = props;
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);
  const location = useLocation();
  const path = getAppRouterPath(location);
  const queryResult = useQuery<ScreenNavQueryQuery>(ScreenNavQueryDocument, {
    client,
    variables: {
      id: appID,
    },
  });
  const app =
    queryResult.data?.node?.__typename === "App" ? queryResult.data.node : null;
  const showIntegrations =
    (app?.effectiveFeatureConfig.google_tag_manager?.disabled ?? false) ===
    false;
  const skippedTutorial = app?.tutorialStatus.data.skipped === true;

  const { auditLogEnabled, analyticEnabled } = useSystemConfig();

  const label = renderToString("ScreenNav.label");
  const [expandState, setExpandState] = useState<Record<string, boolean>>({});

  const links: NavLinkProps[] = useMemo(() => {
    const links = [
      ...(mobileView ? [{ textKey: "ScreenNav.all-projects", url: "/" }] : []),
      ...(skippedTutorial
        ? []
        : [
            {
              textKey: "ScreenNav.getting-started",
              url: `/project/${appID}/getting-started`,
            },
          ]),
      ...(analyticEnabled
        ? [
            {
              textKey: "ScreenNav.analytics",
              url: `/project/${appID}/analytics`,
            },
          ]
        : []),
      { textKey: "ScreenNav.users", url: `/project/${appID}/users` },
      {
        textKey: "ScreenNav.authentication",
        url: `/project/${appID}/configuration/authentication`,
        children: [
          {
            textKey: "ScreenNav.login-id",
            url: `/project/${appID}/configuration/authentication/login-id`,
          },
          {
            textKey: "ScreenNav.authenticators",
            url: `/project/${appID}/configuration/authentication/authenticators`,
          },
          {
            textKey: "ScreenNav.verification",
            url: `/project/${appID}/configuration/authentication/verification`,
          },
        ],
      },
      {
        textKey: "ScreenNav.anonymous-users",
        url: `/project/${appID}/configuration/anonymous-users`,
      },
      {
        textKey: "ScreenNav.biometric",
        url: `/project/${appID}/configuration/biometric`,
      },
      {
        textKey: "ScreenNav.single-sign-on",
        url: `/project/${appID}/configuration/single-sign-on`,
      },
      {
        textKey: "ScreenNav.password-policy",
        url: `/project/${appID}/configuration/password-policy`,
      },
      {
        textKey: "ScreenNav.client-applications",
        url: `/project/${appID}/configuration/apps`,
      },
      {
        textKey: "CustomDomainListScreen.title",
        url: `/project/${appID}/custom-domains`,
      },
      {
        textKey: "ScreenNav.smtp",
        url: `/project/${appID}/configuration/smtp`,
      },
      {
        textKey: "ScreenNav.ui-settings",
        url: `/project/${appID}/configuration/ui-settings`,
      },
      {
        textKey: "ScreenNav.localization",
        url: `/project/${appID}/configuration/localization`,
      },
      {
        textKey: "ScreenNav.user-profile",
        url: `/project/${appID}/configuration/user-profile`,
        children: [
          {
            textKey: "ScreenNav.standard-attributes",
            url: `/project/${appID}/configuration/user-profile/standard-attributes`,
          },
          {
            textKey: "ScreenNav.custom-attributes",
            url: `/project/${appID}/configuration/user-profile/custom-attributes`,
          },
        ],
      },
      ...(showIntegrations
        ? [
            {
              textKey: "ScreenNav.integrations",
              url: `/project/${appID}/integrations`,
            },
          ]
        : []),
      {
        textKey: "ScreenNav.billing",
        url: `/project/${appID}/billing`,
      },
      {
        textKey: "ScreenNav.advanced",
        url: `/project/${appID}/advanced`,
        children: [
          {
            textKey: "ScreenNav.password-reset-code",
            url: `/project/${appID}/advanced/password-reset-code`,
          },
          {
            textKey: "ScreenNav.webhooks",
            url: `/project/${appID}/advanced/webhooks`,
          },
          {
            textKey: "ScreenNav.admin-api",
            url: `/project/${appID}/advanced/admin-api`,
          },
          {
            textKey: "ScreenNav.account-deletion",
            url: `/project/${appID}/advanced/account-deletion`,
          },
          {
            textKey: "ScreenNav.session",
            url: `/project/${appID}/advanced/session`,
          },
        ],
      },
      ...(auditLogEnabled
        ? [
            {
              textKey: "ScreenNav.audit-log",
              url: `/project/${appID}/audit-log`,
            },
          ]
        : []),
      {
        textKey: "PortalAdminSettings.title",
        url: `/project/${appID}/portal-admins`,
      },
    ];

    return links;
  }, [
    analyticEnabled,
    appID,
    auditLogEnabled,
    mobileView,
    showIntegrations,
    skippedTutorial,
  ]);

  const [selectedKeys, selectedKey] = useMemo(() => {
    const matchedKeys: string[] = [];
    let matchLength = 0;
    let matchedKey: string | undefined;
    for (const link of links) {
      const urls = [link.url, ...(link.children ?? []).map((l) => l.url)];
      for (const url of urls) {
        if (!url.includes(path)) {
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
        props.onLinkClick?.();
      }
    },
    [location, navigate, props]
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

  if (queryResult.loading) {
    return null;
  }

  return (
    <Nav
      ariaLabel={label}
      groups={navGroups}
      onLinkClick={onLinkClick}
      onLinkExpandClick={onLinkExpandClick}
      selectedKey={selectedKey}
      styles={getStyles}
    />
  );
};

export default ScreenNav;
