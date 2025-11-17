import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { Context } from "@oursky/react-messageformat";
import {
  Icon,
  Text,
  CommandButton,
  IconButton,
  Panel,
  PanelType,
  IRenderFunction,
  IPanelProps,
  IContextualMenuProps,
  TooltipHost,
} from "@fluentui/react";
import { useAppAndSecretConfigQuery } from "./graphql/portal/query/appAndSecretConfigQuery";
import { useViewerQuery } from "./graphql/portal/query/viewerQuery";
import ScreenNav from "./ScreenNav";
import Link from "./Link";

import styles from "./ScreenHeader.module.css";
import { useSystemConfig } from "./context/SystemConfigContext";
import { useBoolean } from "@fluentui/react-hooks";
import ExternalLink from "./ExternalLink";
import { useLogout } from "./graphql/portal/Authenticated";
import { useCapture } from "./gtm_v2";
import { useSettingsAnchor } from "./hook/authgear";
import { Logo } from "./components/common/Logo";
import { useCopyFeedback } from "./hook/useCopyFeedback";

interface MobileViewHeaderAppSectionProps {
  appID: string;
}

const MobileViewHeaderAppSection: React.VFC<MobileViewHeaderAppSectionProps> = (
  props
) => {
  const { appID } = props;
  const { effectiveAppConfig, isLoading: loading } =
    useAppAndSecretConfigQuery(appID);
  const { themes } = useSystemConfig();

  if (loading) {
    return null;
  }

  const rawAppID = effectiveAppConfig?.id;

  return (
    <Text className={styles.headerAppID} theme={themes.inverted}>
      {rawAppID != null ? rawAppID : appID}
    </Text>
  );
};

interface DesktopViewHeaderAppSectionProps {
  appID: string;
}

const DesktopViewHeaderAppSection: React.VFC<
  DesktopViewHeaderAppSectionProps
> = (props) => {
  const { appID } = props;
  const { effectiveAppConfig, isLoading: loading } =
    useAppAndSecretConfigQuery(appID);
  const { themes } = useSystemConfig();
  const { renderToString } = useContext(Context);

  const rawAppID = effectiveAppConfig?.id;

  const displayAppID = useMemo(
    () => (rawAppID != null ? rawAppID : appID),
    [rawAppID, appID]
  );

  const { copyButtonProps, Feedback } = useCopyFeedback({
    textToCopy: displayAppID,
  });

  if (loading) {
    return null;
  }

  return (
    <>
      <Icon className={styles.headerArrow} iconName="ChevronRight" />
      <TooltipHost content={renderToString("ScreenHeader.copy")}>
        <Text
          id={copyButtonProps.id}
          className={styles.headerAppID}
          theme={themes.inverted}
          onClick={copyButtonProps.onClick}
          onMouseLeave={copyButtonProps.onMouseLeave}
        >
          {displayAppID}
        </Text>
      </TooltipHost>
      <Feedback />
    </>
  );
};

const commandButtonStyles = {
  label: {
    fontSize: "12px",
  },
  menuIcon: {
    fontSize: "12px",
    color: "white",
  },
};

interface MobileViewHeaderIconSectionProps {
  onClick: () => void;
  showHamburger: boolean;
}

const MobileViewHeaderIconSection: React.VFC<
  MobileViewHeaderIconSectionProps
> = (props) => {
  const { onClick, showHamburger } = props;
  const { themes } = useSystemConfig();

  return (
    <>
      {showHamburger ? (
        <IconButton
          ariaLabel="hamburger"
          iconProps={{ iconName: "WaffleOffice365" }}
          className={styles.hamburger}
          theme={themes.inverted}
          onClick={onClick}
        />
      ) : (
        <Link to="/" className={styles.logoLink}>
          <Logo />
        </Link>
      )}
    </>
  );
};

const DesktopViewHeaderIconSection: React.VFC = () => {
  return (
    <Link to="/" className={styles.logoLink}>
      <Logo />
    </Link>
  );
};

const MobileViewNavbarHeader: IRenderFunction<IPanelProps> = (props) => {
  // eslint-disable-next-line @typescript-eslint/no-non-null-asserted-optional-chain
  const onClick: () => void = props?.onDismiss!;
  return (
    <div className={styles.headerMobile}>
      <IconButton
        ariaLabel="hamburger"
        iconProps={{ iconName: "WaffleOffice365" }}
        className={styles.hamburger}
        onClick={onClick}
      />
      <Logo inverted={true} />
    </div>
  );
};

const MobileViewNavbarBody: IRenderFunction<IPanelProps> = (props) => {
  // eslint-disable-next-line @typescript-eslint/no-non-null-asserted-optional-chain
  const onClick: () => void = props?.onDismiss!;
  return <ScreenNav mobileView={true} onLinkClick={onClick} />;
};

interface ScreenNavProps {
  showHamburger?: boolean;
}

const ScreenHeader: React.VFC<ScreenNavProps> = function ScreenHeader(props) {
  const { showHamburger = true } = props;
  const { renderToString } = useContext(Context);
  const capture = useCapture();
  const { themes } = useSystemConfig();
  const { appID } = useParams() as { appID: string };
  const { viewer } = useViewerQuery();
  const [isNavbarOpen, { setTrue: openNavbar, setFalse: dismissNavbar }] =
    useBoolean(false);

  const logout = useLogout();

  const onClickLogout = useCallback(() => {
    logout().catch((err: unknown) => {
      console.error(err);
    });
  }, [logout]);

  const onClickCookiePreference = useCallback(() => {
    if (window.Osano?.cm !== undefined) {
      window.Osano.cm.showDrawer("osano-cm-dom-info-dialog-open");
    } else {
      console.error("Osano is not loaded");
    }
  }, []);

  const onClickContactUs = useCallback(() => {
    capture("header.clicked-contact_us");
  }, [capture]);

  const onClickDocs = useCallback(() => {
    capture("header.clicked-docs");
  }, [capture]);

  const scheduleDemoLink = useMemo(() => {
    const url = new URL("https://www.authgear.com/schedule-demo");
    if (viewer?.email) {
      url.searchParams.append("email", viewer.email);
    }
    if (viewer?.formattedName) {
      url.searchParams.append("name", viewer.formattedName);
    }
    return url.toString();
  }, [viewer?.email, viewer?.formattedName]);

  const headerStyle = useMemo(
    () => ({
      backgroundColor: themes.main.palette.themePrimary,
    }),
    [themes.main]
  );

  const { href: settingURL, onClick: onClickSettings } = useSettingsAnchor();

  const menuProps = useMemo(() => {
    const items = [
      {
        key: "settings",
        text: renderToString("ScreenHeader.settings"),
        iconProps: {
          iconName: "PlayerSettings",
        },
        href: settingURL,
        onClick: onClickSettings,
      },
      {
        key: "logout",
        text: renderToString("ScreenHeader.sign-out"),
        iconProps: {
          iconName: "SignOut",
        },
        onClick: onClickLogout,
      },
    ] satisfies IContextualMenuProps["items"];

    if (window.Osano !== undefined) {
      items.splice(1, 0, {
        key: "cookie",
        text: renderToString("ScreenHeader.cookie-preference"),
        iconProps: {
          iconName: "Cookies",
        },
        onClick: onClickCookiePreference,
      });
    }

    return { items } satisfies IContextualMenuProps;
  }, [
    renderToString,
    settingURL,
    onClickSettings,
    onClickLogout,
    onClickCookiePreference,
  ]);

  return (
    <header className={styles.header} style={headerStyle}>
      <div className={styles.mobileView}>
        <MobileViewHeaderIconSection
          showHamburger={showHamburger}
          onClick={openNavbar}
        />
        {appID ? <MobileViewHeaderAppSection appID={appID} /> : null}
        <Panel
          isLightDismiss={true}
          hasCloseButton={false}
          isOpen={isNavbarOpen}
          onDismiss={dismissNavbar}
          type={PanelType.smallFixedNear}
          onRenderNavigation={MobileViewNavbarHeader}
          onRenderBody={MobileViewNavbarBody}
        />
      </div>
      <div className={styles.desktopView}>
        <DesktopViewHeaderIconSection />
        {appID ? <DesktopViewHeaderAppSection appID={appID} /> : null}
      </div>
      <div className={styles.links}>
        <ExternalLink
          href={scheduleDemoLink}
          className={styles.link}
          onClick={onClickContactUs}
        >
          <Text variant="small">
            {renderToString("ScreenHeader.links.schedule-demo")}
          </Text>
        </ExternalLink>
        <ExternalLink
          href="https://docs.authgear.com/"
          className={styles.link}
          onClick={onClickDocs}
        >
          <Text variant="small">
            {renderToString("ScreenHeader.links.documentation")}
          </Text>
        </ExternalLink>
      </div>
      {viewer != null ? (
        <CommandButton
          className={styles.desktopView}
          menuProps={menuProps}
          theme={themes.inverted}
          styles={commandButtonStyles}
        >
          {viewer.email}
        </CommandButton>
      ) : null}
    </header>
  );
};

export default ScreenHeader;
