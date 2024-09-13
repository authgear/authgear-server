import React, { useCallback, useContext, useMemo, useState } from "react";
import { useQuery } from "@apollo/client";
import { useParams, useLocation, useNavigate } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import {
  INavLink,
  INavLinkGroup,
  INavStyleProps,
  Nav,
  Text,
} from "@fluentui/react";
import authgear from "@authgear/web";
import { useSystemConfig } from "./context/SystemConfigContext";
import {
  ScreenNavQueryQuery,
  ScreenNavQueryDocument,
} from "./graphql/portal/query/screenNavQuery.generated";
import { usePortalClient } from "./graphql/portal/apollo";
import { useAppFeatureConfigQuery } from "./graphql/portal/query/appFeatureConfigQuery";
import { useViewerQuery } from "./graphql/portal/query/viewerQuery";
import styles from "./ScreenNav.module.css";
import ExternalLink from "./ExternalLink";

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
  const { authgearEndpoint } = useSystemConfig();
  const { viewer } = useViewerQuery();
  const client = usePortalClient();
  const queryResult = useQuery<ScreenNavQueryQuery>(ScreenNavQueryDocument, {
    client,
    variables: {
      id: appID,
    },
  });
  const { effectiveFeatureConfig } = useAppFeatureConfigQuery(appID);

  const app =
    queryResult.data?.node?.__typename === "App" ? queryResult.data.node : null;
  const showIntegrations =
    (app?.effectiveFeatureConfig.google_tag_manager?.disabled ?? false) ===
    false;
  const skippedTutorial = app?.tutorialStatus.data.skipped === true;

  const { auditLogEnabled, analyticEnabled, web3Enabled } = useSystemConfig();

  const app2appEnabled = useMemo(() => {
    if (effectiveFeatureConfig != null) {
      return effectiveFeatureConfig.oauth?.client?.app2app_enabled ?? false;
    }
    return false;
  }, [effectiveFeatureConfig]);

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
        type: "group" as const,
        textKey: "ScreenNav.user-management",
        urlPrefix: `/project/${appID}/user-management`,
        children: [
          {
            type: "link" as const,
            textKey: "ScreenNav.users",
            url: `/project/${appID}/user-management/users`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.roles",
            url: `/project/${appID}/user-management/roles`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.groups",
            url: `/project/${appID}/user-management/groups`,
          },
        ],
      },
      {
        type: "group" as const,
        textKey: "ScreenNav.authentication",
        urlPrefix: `/project/${appID}/configuration/authentication`,
        children: [
          {
            type: "link" as const,
            textKey: "ScreenNav.login-methods",
            url: `/project/${appID}/configuration/authentication/login-methods`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.external-oauth",
            url: `/project/${appID}/configuration/authentication/external-oauth`,
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
          ...(app2appEnabled
            ? [
                {
                  type: "link" as const,
                  textKey: "ScreenNav.app2app",
                  url: `/project/${appID}/configuration/authentication/app2app`,
                },
              ]
            : []),
        ],
      },
      {
        type: "link" as const,
        textKey: "ScreenNav.client-applications",
        url: `/project/${appID}/configuration/apps`,
      },
      {
        type: "group" as const,
        textKey: "ScreenNav.branding",
        urlPrefix: `/project/${appID}/branding`,
        children: [
          {
            type: "link" as const,
            textKey: "ScreenNav.design",
            url: `/project/${appID}/branding/design`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.localization",
            url: `/project/${appID}/branding/localization`,
          },
          {
            type: "link" as const,
            textKey: "CustomDomainListScreen.title",
            url: `/project/${appID}/branding/custom-domains`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.customText",
            url: `/project/${appID}/branding/custom-text`,
          },
        ],
      },
      {
        type: "link" as const,
        textKey: "ScreenNav.languages",
        url: `/project/${appID}/configuration/languages`,
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
      {
        type: "link" as const,
        textKey: "ScreenNav.bot-protection",
        url: `/project/${appID}/bot-protection`,
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
            textKey: "ScreenNav.hooks",
            url: `/project/${appID}/advanced/hooks`,
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
            textKey: "ScreenNav.account-anonymization",
            url: `/project/${appID}/advanced/account-anonymization`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.session",
            url: `/project/${appID}/advanced/session`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.smtp",
            url: `/project/${appID}/advanced/smtp`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.endpoint-direct-access",
            url: `/project/${appID}/advanced/endpoint-direct-access`,
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
    mobileView,
    skippedTutorial,
    appID,
    analyticEnabled,
    web3Enabled,
    app2appEnabled,
    showIntegrations,
    auditLogEnabled,
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

  const settingURL = authgearEndpoint + "/settings";
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

  if (queryResult.loading) {
    return null;
  }

  return (
    <>
      <Nav
        ariaLabel={label}
        groups={navGroups}
        onLinkClick={onLinkClick}
        onLinkExpandClick={onLinkExpandClick}
        selectedKey={selectedKey}
        styles={getStyles}
      />
      {mobileView ? (
        <div className={styles.userActions}>
          <Text variant="small" className={styles.userActionEmail}>
            {viewer?.email}
          </Text>
          <ExternalLink
            href={settingURL}
            target="_self"
            className={styles.userActionItem}
          >
            <Text variant="small">
              {renderToString("ScreenHeader.settings")}
            </Text>
          </ExternalLink>
          <button
            type="button"
            className={styles.userActionItem}
            onClick={onClickLogout}
          >
            <Text variant="small">
              {renderToString("ScreenHeader.sign-out")}
            </Text>
          </button>
        </div>
      ) : null}
    </>
  );
};

export default ScreenNav;
