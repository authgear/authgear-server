import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
import { Context } from "./intl";
import { Avatar, DropdownMenu } from "@radix-ui/themes";
import {
  CalendarIcon,
  Cross1Icon,
  FileTextIcon,
  HamburgerMenuIcon,
} from "@radix-ui/react-icons";
import { useViewerQuery } from "./graphql/portal/query/viewerQuery";
import ScreenNav from "./ScreenNav";
import Link from "./Link";

import styles from "./ScreenHeader.module.css";
import { useLogout } from "./graphql/portal/Authenticated";
import { useCapture } from "./gtm_v2";
import { useSettingsAnchor } from "./hook/authgear";
import { Logo } from "./components/common/Logo";
import logoStyles from "./components/common/Logo.module.css";
import ProjectSelector from "./components/header/ProjectSelector";

interface HeaderAppSectionProps {
  appID: string;
}

const HeaderAppSection: React.VFC<HeaderAppSectionProps> = (props) => {
  const { appID } = props;

  return (
    <>
      <span
        className={styles.headerDivider}
        role="separator"
        aria-hidden={true}
      />
      <ProjectSelector appID={appID} />
    </>
  );
};

interface MobileViewHeaderIconSectionProps {
  onClick: () => void;
  showHamburger: boolean;
  menuLabel: string;
}

const MobileViewHeaderIconSection: React.VFC<
  MobileViewHeaderIconSectionProps
> = (props) => {
  const { onClick, showHamburger, menuLabel } = props;

  return showHamburger ? (
    <button
      type="button"
      className={styles.hamburger}
      aria-label={menuLabel}
      onClick={onClick}
    >
      <HamburgerMenuIcon className={styles.hamburgerIcon} />
    </button>
  ) : (
    <Link to="/" className={styles.logoLink}>
      {/* inverted renders the colored logo (logo-inverted.png), which is
          the dark/colored variant meant for a light background. */}
      <Logo
        inverted={true}
        containerClassName={logoStyles.logo__containerHeader}
      />
    </Link>
  );
};

const DesktopViewHeaderIconSection: React.VFC = () => {
  return (
    <Link to="/" className={styles.logoLink}>
      {/* inverted renders the colored logo (logo-inverted.png), which is
          the dark/colored variant meant for a light background. */}
      <Logo
        inverted={true}
        containerClassName={logoStyles.logo__containerHeader}
      />
    </Link>
  );
};

interface ScreenNavProps {
  showHamburger?: boolean;
}

const ScreenHeader: React.VFC<ScreenNavProps> = function ScreenHeader(props) {
  const { showHamburger = true } = props;
  const { renderToString } = useContext(Context);
  const capture = useCapture();
  const { appID } = useParams() as { appID: string };
  const { viewer } = useViewerQuery();
  const [isNavbarOpen, setIsNavbarOpen] = useState(false);
  // Kept mounted through the close animation, then unmounted.
  const [isNavbarMounted, setIsNavbarMounted] = useState(false);
  const openNavbar = useCallback(() => {
    // Mount immediately so the slide-in animation plays from the start.
    setIsNavbarMounted(true);
    setIsNavbarOpen(true);
  }, []);
  const dismissNavbar = useCallback(() => setIsNavbarOpen(false), []);

  // Delay unmount so the slide-out animation (200ms, matching the CSS) can
  // play before the drawer leaves the DOM.
  useEffect(() => {
    if (isNavbarOpen || !isNavbarMounted) {
      return undefined;
    }
    const timer = setTimeout(() => setIsNavbarMounted(false), 200);
    return () => clearTimeout(timer);
  }, [isNavbarOpen, isNavbarMounted]);

  // Close the mobile drawer on Escape and lock body scroll while it is open.
  useEffect(() => {
    if (!isNavbarOpen) {
      return undefined;
    }
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setIsNavbarOpen(false);
      }
    };
    document.addEventListener("keydown", onKeyDown);
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      document.removeEventListener("keydown", onKeyDown);
      document.body.style.overflow = prevOverflow;
    };
  }, [isNavbarOpen]);

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

  // eslint-disable-next-line react-hooks/preserve-manual-memoization
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

  const { href: settingURL, onClick: onClickSettings } = useSettingsAnchor();

  // Name is optional; fall back to showing the email as the primary line.
  const menuName =
    viewer?.formattedName != null && viewer.formattedName.trim() !== ""
      ? viewer.formattedName.trim()
      : null;
  const menuEmail =
    viewer?.email != null && viewer.email !== "" ? viewer.email : null;

  // Fallback shown when there is no profile picture: the first letter of the
  // name, or of the email when no name is set.
  const avatarFallback = useMemo(() => {
    const base = menuName ?? menuEmail ?? "";
    return base.trim().charAt(0).toUpperCase() || "?";
  }, [menuName, menuEmail]);

  const hasOsano = window.Osano !== undefined;

  return (
    <header className={styles.header}>
      <div className={styles.mobileView}>
        <MobileViewHeaderIconSection
          showHamburger={showHamburger}
          onClick={openNavbar}
          menuLabel={renderToString("ScreenNav.open-menu")}
        />
        {appID ? <HeaderAppSection appID={appID} /> : null}
        {isNavbarMounted ? (
          <div
            className={cn(styles.drawer, !isNavbarOpen && styles.drawerClosing)}
            role="dialog"
            aria-modal={true}
          >
            <button
              type="button"
              className={styles.drawerOverlay}
              aria-label={renderToString("close")}
              onClick={dismissNavbar}
            />
            <div className={styles.drawerPanel}>
              <div className={styles.drawerHeader}>
                <button
                  type="button"
                  className={styles.hamburger}
                  aria-label={renderToString("close")}
                  onClick={dismissNavbar}
                >
                  <Cross1Icon className={styles.hamburgerIcon} />
                </button>
                <Link
                  to="/"
                  className={styles.logoLink}
                  onClick={dismissNavbar}
                >
                  <Logo
                    inverted={true}
                    containerClassName={logoStyles.logo__containerHeader}
                  />
                </Link>
              </div>
              <div className={styles.drawerBody}>
                <ScreenNav mobileView={true} onLinkClick={dismissNavbar} />
              </div>
            </div>
          </div>
        ) : null}
      </div>
      <div className={styles.desktopView}>
        <DesktopViewHeaderIconSection />
        {appID ? <HeaderAppSection appID={appID} /> : null}
      </div>
      <div className={styles.actions}>
        <div className={styles.actionLinks}>
          <a
            className={styles.actionLink}
            href="https://docs.authgear.com/"
            target="_blank"
            rel="noreferrer"
            onClick={onClickDocs}
          >
            <FileTextIcon className={styles.actionIcon} />
            {renderToString("ScreenHeader.links.documentation")}
          </a>
          <a
            className={styles.actionLink}
            href={scheduleDemoLink}
            target="_blank"
            rel="noreferrer"
            onClick={onClickContactUs}
          >
            <CalendarIcon className={styles.actionIcon} />
            {renderToString("ScreenHeader.links.schedule-demo")}
          </a>
        </div>
        {viewer != null ? (
          <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <button
                type="button"
                className={styles.avatarButton}
                aria-label={renderToString("ScreenHeader.user-menu")}
              >
                <Avatar
                  size="2"
                  radius="full"
                  variant="soft"
                  color="blue"
                  src={viewer.picture ?? undefined}
                  fallback={avatarFallback}
                />
              </button>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content align="end">
              {menuName != null || menuEmail != null ? (
                <>
                  <div className={styles.userMenuIdentity}>
                    {menuName != null ? (
                      <span className={styles.userMenuName}>{menuName}</span>
                    ) : null}
                    {menuEmail != null ? (
                      <span
                        className={
                          menuName != null
                            ? styles.userMenuEmail
                            : styles.userMenuName
                        }
                      >
                        {menuEmail}
                      </span>
                    ) : null}
                  </div>
                  <DropdownMenu.Separator />
                </>
              ) : null}
              {hasOsano ? (
                <DropdownMenu.Item onSelect={onClickCookiePreference}>
                  {renderToString("ScreenHeader.cookie-preference")}
                </DropdownMenu.Item>
              ) : null}
              <DropdownMenu.Item asChild={true}>
                <a href={settingURL} onClick={onClickSettings}>
                  {renderToString("ScreenHeader.settings")}
                </a>
              </DropdownMenu.Item>
              <DropdownMenu.Separator />
              <DropdownMenu.Item color="red" onSelect={onClickLogout}>
                {renderToString("ScreenHeader.sign-out")}
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        ) : null}
      </div>
    </header>
  );
};

export default ScreenHeader;
