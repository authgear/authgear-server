import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { useQuery } from "@apollo/client";
import { useParams, useLocation, useNavigate } from "react-router-dom";
import {
  AvatarIcon,
  BarChartIcon,
  ChevronRightIcon,
  CodeIcon,
  DashboardIcon,
  ExclamationTriangleIcon,
  GearIcon,
  GlobeIcon,
  HomeIcon,
  IdCardIcon,
  LockClosedIcon,
  MixerHorizontalIcon,
  Pencil2Icon,
  PersonIcon,
  ReaderIcon,
  RocketIcon,
} from "@radix-ui/react-icons";
import { Context } from "./intl";
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
import { useSettingsAnchor } from "./hook/authgear";

type NavIconComponent = typeof RocketIcon;

type NavLinkItem = NavLink | NavLinkGroup;

interface NavLinkGroup {
  type: "group";
  textKey: string;
  urlPrefix: string;
  url?: string;
  icon: NavIconComponent;
  children: NavLink[];
}

interface NavLink {
  type: "link";
  textKey: string;
  url: string;
  icon?: NavIconComponent;
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
  const { mobileView = false, onLinkClick: onLinkClickProp } = props;
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
    const links: NavLinkItem[] = [
      ...(mobileView
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.all-projects",
              url: "/",
              icon: HomeIcon,
            },
          ]
        : []),
      {
        type: "link" as const,
        textKey: "ScreenNav.getting-started",
        url: `/project/${appID}/getting-started`,
        icon: RocketIcon,
      },
      ...(analyticEnabled
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.analytics",
              url: `/project/${appID}/analytics`,
              icon: BarChartIcon,
            },
          ]
        : []),
      {
        type: "group" as const,
        textKey: "ScreenNav.user-management",
        urlPrefix: `/project/${appID}/user-management`,
        icon: PersonIcon,
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
        icon: LockClosedIcon,
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
        icon: DashboardIcon,
      },
      {
        type: "link" as const,
        textKey: "ScreenNav.api-resources",
        url: `/project/${appID}/api-resources`,
        icon: CodeIcon,
      },
      {
        type: "group" as const,
        textKey: "ScreenNav.branding",
        urlPrefix: `/project/${appID}/branding`,
        icon: Pencil2Icon,
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
        icon: GlobeIcon,
      },
      {
        type: "group" as const,
        textKey: "ScreenNav.user-profile",
        urlPrefix: `/project/${appID}/configuration/user-profile`,
        icon: AvatarIcon,
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
        type: "group" as const,
        textKey: "ScreenNav.attack-protection",
        urlPrefix: `/project/${appID}/attack-protection`,
        icon: ExclamationTriangleIcon,
        children: [
          {
            type: "link" as const,
            textKey: "ScreenNav.account-lockout",
            url: `/project/${appID}/attack-protection/account-lockout`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.bot-protection",
            url: `/project/${appID}/attack-protection/bot-protection`,
          },
          ...(fraudProtectionModifiable
            ? [
                {
                  type: "link" as const,
                  textKey: "ScreenNav.fraud-protection",
                  url: `/project/${appID}/attack-protection/fraud-protection`,
                },
              ]
            : []),
          {
            type: "link" as const,
            textKey: "ScreenNav.ip-blocklist",
            url: `/project/${appID}/attack-protection/ip-blocklist`,
          },
        ],
      },
      ...(showIntegrations
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.integrations",
              url: `/project/${appID}/integrations`,
              icon: GearIcon,
            },
          ]
        : []),

      ...(isAuthgearOnce
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.license",
              url: `/project/${appID}/license`,
              icon: IdCardIcon,
            },
          ]
        : [
            {
              type: "link" as const,
              textKey: "ScreenNav.billing",
              url: `/project/${appID}/billing`,
              icon: IdCardIcon,
            },
          ]),

      {
        type: "group" as const,
        textKey: "ScreenNav.advanced",
        urlPrefix: `/project/${appID}/advanced`,
        icon: MixerHorizontalIcon,
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
            textKey: isAuthgearOnce
              ? "ScreenNav.smtp--authgearonce"
              : "ScreenNav.smtp",
            url: `/project/${appID}/advanced/smtp`,
          },
          {
            type: "link" as const,
            textKey: isAuthgearOnce
              ? "ScreenNav.sms-gateway--authgearonce"
              : "ScreenNav.sms-gateway",
            url: `/project/${appID}/advanced/sms-gateway`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.endpoint-direct-access",
            url: `/project/${appID}/advanced/endpoint-direct-access`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.saml-certificate",
            url: `/project/${appID}/advanced/saml-certificate`,
          },
          {
            type: "link" as const,
            textKey: "ScreenNav.edit-config",
            url: `/project/${appID}/edit-config`,
          },
        ],
      },
      ...(auditLogEnabled
        ? [
            {
              type: "link" as const,
              textKey: "ScreenNav.audit-log",
              url: `/project/${appID}/audit-log`,
              icon: ReaderIcon,
            },
          ]
        : []),
      {
        type: "link" as const,
        textKey: "PortalAdminSettings.title",
        url: `/project/${appID}/portal-admins`,
        icon: PersonIcon,
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

  const goTo = useCallback(
    (url: string) => {
      navigate(url);
      onLinkClickProp?.();
    },
    [navigate, onLinkClickProp]
  );

  const toggleGroup = useCallback((urlPrefix: string) => {
    setExpandState((s) => ({ ...s, [urlPrefix]: !Boolean(s[urlPrefix]) }));
  }, []);

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

  const renderLink = (item: NavLink, isChild: boolean) => {
    const Icon = item.icon;
    const selected = selectedKey === item.url;
    return (
      <a
        key={item.url}
        href={item.url}
        className={cn(
          styles.item,
          isChild && styles.childItem,
          selected && styles.itemSelected
        )}
        aria-current={selected ? "page" : undefined}
        onClick={(e) => {
          e.preventDefault();
          goTo(item.url);
        }}
      >
        {!isChild && Icon != null ? <Icon className={styles.itemIcon} /> : null}
        <span className={styles.itemLabel}>{renderToString(item.textKey)}</span>
      </a>
    );
  };

  const renderGroup = (item: NavLinkGroup) => {
    const Icon = item.icon;
    const expanded =
      Boolean(expandState[item.urlPrefix]) ||
      pathname.startsWith(item.urlPrefix);
    return (
      <div key={item.urlPrefix} className={styles.group}>
        <button
          type="button"
          className={styles.item}
          aria-expanded={expanded}
          onClick={() => toggleGroup(item.urlPrefix)}
        >
          <Icon className={styles.itemIcon} />
          <span className={styles.itemLabel}>
            {renderToString(item.textKey)}
          </span>
          <ChevronRightIcon
            className={cn(
              styles.groupChevron,
              expanded && styles.groupChevronExpanded
            )}
          />
        </button>
        {expanded ? (
          <div className={styles.groupChildren}>
            {item.children.map((child) => renderLink(child, true))}
          </div>
        ) : null}
      </div>
    );
  };

  if (queryResult.loading) {
    return null;
  }

  return (
    <>
      <nav className={styles.navList} aria-label={label}>
        {links.map((item) =>
          item.type === "group" ? renderGroup(item) : renderLink(item, false)
        )}
      </nav>
      {mobileView ? (
        <div className={styles.userActions}>
          <span className={styles.userActionEmail}>{viewer?.email}</span>
          <a
            href={settingURL}
            target="_self"
            className={styles.userActionItem}
            onClick={onClickSettings}
          >
            {renderToString("ScreenHeader.settings")}
          </a>
          <button
            type="button"
            className={styles.userActionItem}
            onClick={onClickLogout}
          >
            {renderToString("ScreenHeader.sign-out")}
          </button>
        </div>
      ) : null}
    </>
  );
};

export default ScreenNav;
