import React, { useCallback, useContext, useMemo, useState } from "react";
import { useQuery } from "@apollo/client";
import { useParams, useLocation, useNavigate } from "react-router-dom";
import { Context } from "./intl";
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
import { useSettingsAnchor } from "./hook/authgear";
import { projectPath, ProjectSectionPath } from "./util/projectPath";

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
  url?: string;
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

  const { isAuthgearOnce, auditLogEnabled, analyticEnabled } =
    useSystemConfig();

  const app2appEnabled = useMemo(() => {
    if (effectiveFeatureConfig != null) {
      return effectiveFeatureConfig.oauth?.client?.app2app_enabled ?? false;
    }
    return false;
  }, [effectiveFeatureConfig]);

  const fraudProtectionModifiable = useMemo(() => {
    return effectiveFeatureConfig?.fraud_protection?.is_modifiable ?? false;
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
      {
        type: "link" as const,
        textKey: "ScreenNav.getting-started",
        url: projectPath(appID, ProjectSectionPath.gettingStarted),
      },
      ...(analyticEnabled
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.analytics",
              url: projectPath(appID, ProjectSectionPath.analytics),
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
            url: projectPath(appID, ProjectSectionPath.users),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.roles",
            url: projectPath(appID, ProjectSectionPath.roles),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.groups",
            url: projectPath(appID, ProjectSectionPath.groups),
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
            url: projectPath(appID, ProjectSectionPath.loginMethods),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.external-oauth",
            url: projectPath(appID, ProjectSectionPath.externalOAuth),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.biometric",
            url: projectPath(appID, ProjectSectionPath.biometric),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.mfa",
            url: projectPath(appID, ProjectSectionPath.mfa),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.anonymous-users",
            url: projectPath(appID, ProjectSectionPath.anonymousUsers),
          },
          ...(app2appEnabled
            ? [
                {
                  type: "link" as const,
                  textKey: "ScreenNav.app2app",
                  url: projectPath(appID, ProjectSectionPath.app2app),
                },
              ]
            : []),
        ],
      },
      {
        type: "link" as const,
        textKey: "ScreenNav.client-applications",
        url: projectPath(appID, ProjectSectionPath.clientApplications),
      },
      {
        type: "link" as const,
        textKey: "ScreenNav.api-resources",
        url: projectPath(appID, ProjectSectionPath.apiResources),
      },
      {
        type: "group" as const,
        textKey: "ScreenNav.branding",
        urlPrefix: `/project/${appID}/branding`,
        children: [
          {
            type: "link" as const,
            textKey: "ScreenNav.design",
            url: projectPath(appID, ProjectSectionPath.design),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.localization",
            url: projectPath(appID, ProjectSectionPath.localization),
          },
          {
            type: "link" as const,
            textKey: "CustomDomainListScreen.title",
            url: projectPath(appID, ProjectSectionPath.customDomains),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.customText",
            url: projectPath(appID, ProjectSectionPath.customText),
          },
        ],
      },
      {
        type: "link" as const,
        textKey: "ScreenNav.languages",
        url: projectPath(appID, ProjectSectionPath.languages),
      },
      {
        type: "group" as const,
        textKey: "ScreenNav.user-profile",
        urlPrefix: `/project/${appID}/configuration/user-profile`,
        children: [
          {
            type: "link" as const,
            textKey: "ScreenNav.standard-attributes",
            url: projectPath(appID, ProjectSectionPath.standardAttributes),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.custom-attributes",
            url: projectPath(appID, ProjectSectionPath.customAttributes),
          },
        ],
      },
      {
        type: "group" as const,
        textKey: "ScreenNav.attack-protection",
        urlPrefix: `/project/${appID}/attack-protection`,
        children: [
          {
            type: "link" as const,
            textKey: "ScreenNav.bot-protection",
            url: projectPath(appID, ProjectSectionPath.botProtection),
          },
          ...(fraudProtectionModifiable
            ? [
                {
                  type: "link" as const,
                  textKey: "ScreenNav.fraud-protection",
                  url: projectPath(appID, ProjectSectionPath.fraudProtection),
                },
              ]
            : []),
          {
            type: "link" as const,
            textKey: "ScreenNav.ip-blocklist",
            url: projectPath(appID, ProjectSectionPath.ipBlocklist),
          },
        ],
      },
      ...(showIntegrations
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.integrations",
              url: projectPath(appID, ProjectSectionPath.integrations),
            },
          ]
        : []),

      ...(isAuthgearOnce
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.license",
              url: projectPath(appID, ProjectSectionPath.license),
            },
          ]
        : [
            {
              type: "link" as const,
              textKey: "ScreenNav.billing",
              url: projectPath(appID, ProjectSectionPath.billing),
            },
          ]),

      {
        type: "group" as const,
        textKey: "ScreenNav.advanced",
        urlPrefix: `/project/${appID}/advanced`,
        children: [
          {
            type: "link" as const,
            textKey: "ScreenNav.hooks",
            url: projectPath(appID, ProjectSectionPath.hooks),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.admin-api",
            url: projectPath(appID, ProjectSectionPath.adminAPI),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.account-deletion",
            url: projectPath(appID, ProjectSectionPath.accountDeletion),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.account-anonymization",
            url: projectPath(appID, ProjectSectionPath.accountAnonymization),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.session",
            url: projectPath(appID, ProjectSectionPath.session),
          },
          {
            type: "link" as const,
            textKey: isAuthgearOnce
              ? "ScreenNav.smtp--authgearonce"
              : "ScreenNav.smtp",
            url: projectPath(appID, ProjectSectionPath.smtp),
          },
          {
            type: "link" as const,
            textKey: isAuthgearOnce
              ? "ScreenNav.sms-gateway--authgearonce"
              : "ScreenNav.sms-gateway",
            url: projectPath(appID, ProjectSectionPath.smsGateway),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.endpoint-direct-access",
            url: projectPath(appID, ProjectSectionPath.endpointDirectAccess),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.saml-certificate",
            url: projectPath(appID, ProjectSectionPath.samlCertificate),
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.edit-config",
            url: projectPath(appID, ProjectSectionPath.editConfig),
          },
        ],
      },
      ...(auditLogEnabled
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.audit-log",
              url: projectPath(appID, ProjectSectionPath.auditLog),
            },
          ]
        : []),
      {
        type: "link" as const,
        textKey: "PortalAdminSettings.title",
        url: projectPath(appID, ProjectSectionPath.portalAdmins),
      },
    ];

    return links;
  }, [
    isAuthgearOnce,
    mobileView,
    appID,
    analyticEnabled,
    app2appEnabled,
    fraudProtectionModifiable,
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
            url: item.url ?? "",
            links: item.children.map((child) => navItem(child)),
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
        links: links.map((item) => navItem(item)),
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

  const { href: settingURL, onClick: onClickSettings } = useSettingsAnchor();

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
            onClick={onClickSettings}
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
