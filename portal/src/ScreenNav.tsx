import React, { useCallback, useContext, useMemo, useState } from "react";
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

type NavLinkItem = NavLink | NavLinkGroup;

interface NavLinkGroup {
  type: "group";
  textKey: string;
  urlPrefix: string;
  children: NavLink[];
}

interface NavLink {
  type: "link";
  textKey: string;
  url: string;
}

interface ScreenNavProps {
  mobileView?: boolean;
  onLinkClick?: () => void;
}

function makeInitialExpandState(
  items: NavLinkItem[],
  pathname: string
): Record<string, boolean> {
  const out: Record<string, boolean> = {};
  for (const item of items) {
    if (item.type === "group") {
      if (pathname.startsWith(item.urlPrefix)) {
        out[item.urlPrefix] = true;
      }
    }
  }
  return out;
}

function getSelectedKey(
  items: NavLinkItem[],
  pathname: string
): string | undefined {
  let out = "";
  for (const item of items) {
    switch (item.type) {
      case "group": {
        for (const link of item.children) {
          if (pathname.startsWith(link.url)) {
            if (link.url.length > out.length) {
              out = link.url;
            }
          }
        }
        break;
      }
      case "link": {
        if (pathname.startsWith(item.url)) {
          if (item.url.length > out.length) {
            out = item.url;
          }
        }
        break;
      }
      default:
        break;
    }
  }
  if (out === "") {
    return undefined;
  }
  return out;
}

// We simplified the group expand/collapse logic.
//
// 1. We no longer tangle with splitting the path. That is too fragile.
// 2. We clearly define NavLink and NavLinkGroup as separate types.
// 3. The expandState is initialized with pathname on mount. So switching route WILL NOT collapse a group.
// 4. isExpanded is true if urlPrefix is a prefix of pathname.
// 5. selectedKey is the longest prefix match.
const ScreenNav: React.VFC<ScreenNavProps> = function ScreenNav(props) {
  const { mobileView = false } = props;
  const { appID } = useParams() as { appID: string };
  const navigate = useNavigate();
  const { renderToString } = useContext(Context);
  const { pathname } = useLocation();
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

  const { auditLogEnabled, analyticEnabled, web3Enabled } = useSystemConfig();

  const label = renderToString("ScreenNav.label");

  const links: NavLinkItem[] = useMemo(() => {
    const links = [
      ...(mobileView
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.all-projects",
              url: "/",
            },
          ]
        : []),
      ...(skippedTutorial
        ? []
        : [
            {
              type: "link" as const,
              textKey: "ScreenNav.getting-started",
              url: `/project/${appID}/getting-started`,
            },
          ]),
      ...(analyticEnabled
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.analytics",
              url: `/project/${appID}/analytics`,
            },
          ]
        : []),
      {
        type: "link" as const,
        textKey: "ScreenNav.users",
        url: `/project/${appID}/users`,
      },
      {
        type: "group" as const,
        textKey: "ScreenNav.authentication",
        urlPrefix: `/project/${appID}/configuration/authentication`,
        children: [
          {
            type: "link" as const,
            textKey: "ScreenNav.login-id",
            url: `/project/${appID}/configuration/authentication/login-id`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.authenticators",
            url: `/project/${appID}/configuration/authentication/authenticators`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.verification",
            url: `/project/${appID}/configuration/authentication/verification`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.external-oauth",
            url: `/project/${appID}/configuration/authentication/external-oauth`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.passkey",
            url: `/project/${appID}/configuration/authentication/passkey`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.biometric",
            url: `/project/${appID}/configuration/authentication/biometric`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.mfa",
            url: `/project/${appID}/configuration/authentication/2fa`,
          },
          ...(web3Enabled
            ? [
                {
                  type: "link" as const,
                  textKey: "ScreenNav.web3",
                  url: `/project/${appID}/configuration/authentication/web3`,
                },
              ]
            : []),
          {
            type: "link" as const,
            textKey: "ScreenNav.anonymous-users",
            url: `/project/${appID}/configuration/authentication/anonymous-users`,
          },
        ],
      },
      {
        type: "link" as const,
        textKey: "ScreenNav.password-policy",
        url: `/project/${appID}/configuration/password-policy`,
      },
      {
        type: "link" as const,
        textKey: "ScreenNav.client-applications",
        url: `/project/${appID}/configuration/apps`,
      },
      {
        type: "link" as const,
        textKey: "CustomDomainListScreen.title",
        url: `/project/${appID}/custom-domains`,
      },
      {
        type: "link" as const,
        textKey: "ScreenNav.smtp",
        url: `/project/${appID}/configuration/smtp`,
      },
      {
        type: "link" as const,
        textKey: "ScreenNav.ui-settings",
        url: `/project/${appID}/configuration/ui-settings`,
      },
      {
        type: "link" as const,
        textKey: "ScreenNav.localization",
        url: `/project/${appID}/configuration/localization`,
      },
      {
        type: "group" as const,
        textKey: "ScreenNav.user-profile",
        urlPrefix: `/project/${appID}/configuration/user-profile`,
        children: [
          {
            type: "link" as const,
            textKey: "ScreenNav.standard-attributes",
            url: `/project/${appID}/configuration/user-profile/standard-attributes`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.custom-attributes",
            url: `/project/${appID}/configuration/user-profile/custom-attributes`,
          },
        ],
      },
      ...(showIntegrations
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.integrations",
              url: `/project/${appID}/integrations`,
            },
          ]
        : []),
      {
        type: "link" as const,
        textKey: "ScreenNav.billing",
        url: `/project/${appID}/billing`,
      },
      {
        type: "group" as const,
        textKey: "ScreenNav.advanced",
        urlPrefix: `/project/${appID}/advanced`,
        children: [
          {
            type: "link" as const,
            textKey: "ScreenNav.password-reset-code",
            url: `/project/${appID}/advanced/password-reset-code`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.webhooks",
            url: `/project/${appID}/advanced/webhooks`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.admin-api",
            url: `/project/${appID}/advanced/admin-api`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.account-deletion",
            url: `/project/${appID}/advanced/account-deletion`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.session",
            url: `/project/${appID}/advanced/session`,
          },
        ],
      },
      ...(auditLogEnabled
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.audit-log",
              url: `/project/${appID}/audit-log`,
            },
          ]
        : []),
      {
        type: "link" as const,
        textKey: "PortalAdminSettings.title",
        url: `/project/${appID}/portal-admins`,
      },
    ];

    return links;
  }, [
    analyticEnabled,
    appID,
    auditLogEnabled,
    web3Enabled,
    mobileView,
    showIntegrations,
    skippedTutorial,
  ]);

  const [expandState, setExpandState] = useState<Record<string, boolean>>(
    () => {
      return makeInitialExpandState(links, pathname);
    }
  );

  const selectedKey = useMemo(
    () => getSelectedKey(links, pathname),
    [links, pathname]
  );

  const navItem = useCallback(
    (item: NavLinkItem): INavLink => {
      switch (item.type) {
        case "group": {
          return {
            isExpanded:
              Boolean(expandState[item.urlPrefix]) ||
              pathname.startsWith(item.urlPrefix),
            key: item.urlPrefix,
            name: renderToString(item.textKey),
            url: "",
            links: item.children.map(navItem),
          };
        }
        case "link": {
          return {
            key: item.url,
            name: renderToString(item.textKey),
            url: item.url,
          };
        }
        default:
          throw new Error("unreachable");
      }
    },
    [expandState, pathname, renderToString]
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

      const url = item?.url;
      if (url != null && url !== "") {
        navigate(url);
        props.onLinkClick?.();
      }
    },
    [navigate, props]
  );
  const onLinkExpandClick = useCallback(
    (e?: React.MouseEvent, item?: INavLink) => {
      e?.stopPropagation();
      e?.preventDefault();
      const key = item?.key;
      if (key != null) {
        setExpandState((s) => ({ ...s, [key]: !Boolean(s[key]) }));
      }
    },
    []
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
