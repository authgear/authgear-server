import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useLocation, useNavigate } from "react-router-dom";
import {
  Callout,
  DirectionalHint,
  Icon,
  IconButton,
  IButtonStyles,
  ITooltipHost,
  Text,
  Theme,
  TooltipHost,
} from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { useAppListQuery } from "../../graphql/portal/query/appListQuery";
import { useViewerQuery } from "../../graphql/portal/query/viewerQuery";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useCapture } from "../../gtm_v2";
import { toTypedID } from "../../util/graphql";
import { resolveProjectSwitchPath } from "../../util/projectPath";
import { isProjectQuotaReached } from "../../util/projectQuota";
import { copyToClipboard } from "../../util/clipboard";
import styles from "./ProjectSelector.module.css";

const COPY_ICON_PROPS = { iconName: "Copy" };

interface ProjectSelectorCopyButtonProps {
  projectID: string;
  buttonStyles: IButtonStyles;
}

const ProjectSelectorCopyButton: React.VFC<ProjectSelectorCopyButtonProps> =
  function ProjectSelectorCopyButton({ projectID, buttonStyles }) {
    const { renderToString } = useContext(Context);
    const [copied, setCopied] = useState(false);
    const tooltipHostRef = useRef<ITooltipHost | null>(null);
    const resetCopiedTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(
      null
    );

    const copyLabel = renderToString("ScreenHeader.copy");
    const copiedLabel = renderToString("copied-to-clipboard");

    const scheduleResetCopied = useCallback(() => {
      if (resetCopiedTimeoutRef.current != null) {
        clearTimeout(resetCopiedTimeoutRef.current);
      }
      resetCopiedTimeoutRef.current = setTimeout(() => {
        setCopied(false);
        tooltipHostRef.current?.dismiss();
        resetCopiedTimeoutRef.current = null;
      }, 2000);
    }, []);

    useEffect(() => {
      return () => {
        if (resetCopiedTimeoutRef.current != null) {
          clearTimeout(resetCopiedTimeoutRef.current);
        }
      };
    }, []);

    useEffect(() => {
      if (copied) {
        tooltipHostRef.current?.show();
      }
    }, [copied]);

    const onCopyClick = useCallback(
      (e: React.MouseEvent<HTMLButtonElement>) => {
        e.preventDefault();
        e.stopPropagation();
        copyToClipboard(projectID);
        setCopied(true);
        scheduleResetCopied();
      },
      [projectID, scheduleResetCopied]
    );

    const stopPropagation = useCallback((e: React.SyntheticEvent) => {
      e.stopPropagation();
    }, []);

    const onCopyMouseLeave = useCallback(() => {
      scheduleResetCopied();
    }, [scheduleResetCopied]);

    const copyButton = (
      <IconButton
        iconProps={COPY_ICON_PROPS}
        styles={buttonStyles}
        ariaLabel={copied ? copiedLabel : copyLabel}
        title={copied ? copiedLabel : copyLabel}
        onClick={onCopyClick}
      />
    );

    return (
      <span
        className={styles.copyButtonWrap}
        onMouseDown={stopPropagation}
        onClick={stopPropagation}
        onMouseLeave={onCopyMouseLeave}
      >
        {copied ? (
          <TooltipHost
            componentRef={tooltipHostRef}
            content={copiedLabel}
            delay={0}
            directionalHint={DirectionalHint.topCenter}
            calloutProps={{
              gapSpace: 4,
              role: "tooltip",
            }}
          >
            {copyButton}
          </TooltipHost>
        ) : (
          copyButton
        )}
      </span>
    );
  };

interface ProjectSelectorProps {
  appID: string;
  theme?: Theme;
}

const ProjectSelector: React.VFC<ProjectSelectorProps> =
  function ProjectSelector(props) {
    const { appID, theme } = props;
    const { renderToString } = useContext(Context);
    const navigate = useNavigate();
    const location = useLocation();
    const capture = useCapture();
    const triggerRef = useRef<HTMLButtonElement>(null);
    const [isCalloutOpen, setIsCalloutOpen] = useState(false);

    const { effectiveAppConfig, isLoading: loadingAppConfig } =
      useAppAndSecretConfigQuery(appID);
    const { apps, loading: loadingAppList } = useAppListQuery();
    const { viewer } = useViewerQuery();
    const { themes, authgearAppID, isAuthgearOnce } = useSystemConfig();

    const accentColor = themes.main.palette.themePrimary;
    const accentHoverColor = themes.main.palette.themeDark;

    const copyIconButtonStyles = useMemo(
      () => ({
        root: { color: accentColor },
        rootHovered: { color: accentHoverColor },
      }),
      [accentColor, accentHoverColor]
    );

    const calloutStyle = useMemo(
      () =>
        ({
          "--project-selector-accent": accentColor,
        } as React.CSSProperties),
      [accentColor]
    );

    const displayAppID = useMemo(() => {
      const rawAppID = effectiveAppConfig?.id;
      return rawAppID != null ? rawAppID : appID;
    }, [effectiveAppConfig?.id, appID]);

    const filteredApps = useMemo(() => {
      return (apps ?? []).filter((a) => {
        if (isAuthgearOnce && a.appID === authgearAppID) {
          return false;
        }
        return true;
      });
    }, [apps, isAuthgearOnce, authgearAppID]);

    const sortedApps = useMemo(() => {
      return [...filteredApps].sort((a, b) => a.appID.localeCompare(b.appID));
    }, [filteredApps]);

    const otherApps = useMemo(() => {
      return sortedApps.filter((app) => app.appID !== displayAppID);
    }, [sortedApps, displayAppID]);

    const createButtonDisabled =
      isProjectQuotaReached(viewer ?? null) || isAuthgearOnce;

    const onToggleCallout = useCallback((e: React.MouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setIsCalloutOpen((open) => !open);
    }, []);

    // Callout's onDismiss handles dismissing on outside clicks and on the
    // Escape key (via its inner Popup), so no manual listeners are needed.
    const onDismissCallout = useCallback(() => {
      setIsCalloutOpen(false);
    }, []);

    const onSelectProject = useCallback(
      (selectedAppID: string) => {
        if (selectedAppID === displayAppID) {
          setIsCalloutOpen(false);
          return;
        }
        capture(
          "enteredProject",
          { projectID: selectedAppID },
          { project_id: selectedAppID }
        );
        const typedID = toTypedID("App", selectedAppID);
        const newProjectBasePath = `/project/${encodeURIComponent(typedID)}`;
        // Keep the user on the same section of the new project, but drop any
        // project-specific suffix (entity IDs, create/edit forms) that would
        // not exist in the target project. Search and hash are intentionally
        // dropped as they typically reference data from the previous project.
        const nextPathname = resolveProjectSwitchPath(
          location.pathname,
          newProjectBasePath
        );

        navigate(nextPathname);
        setIsCalloutOpen(false);
      },
      [capture, displayAppID, location.pathname, navigate]
    );

    const onCreateProject = useCallback(
      (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        if (createButtonDisabled) {
          return;
        }
        navigate("/projects/create");
        setIsCalloutOpen(false);
      },
      [createButtonDisabled, navigate]
    );

    if (loadingAppConfig || loadingAppList) {
      return null;
    }

    const openMenuLabel = renderToString(
      "ScreenHeader.projectSelector.open-menu"
    );

    return (
      <>
        <button
          ref={triggerRef}
          type="button"
          className={styles.trigger}
          onClick={onToggleCallout}
          aria-expanded={isCalloutOpen}
          aria-haspopup="true"
          aria-label={openMenuLabel}
        >
          <Text className={styles.triggerLabel} theme={theme}>
            {displayAppID}
          </Text>
          <Icon
            className={styles.triggerChevron}
            iconName="ChevronDown"
            theme={theme}
          />
        </button>
        {isCalloutOpen ? (
          <Callout
            target={triggerRef.current}
            gapSpace={4}
            isBeakVisible={false}
            className={styles.callout}
            style={calloutStyle}
            onDismiss={onDismissCallout}
          >
            <div>
              <section className={styles.section}>
                <div className={styles.sectionHeader}>
                  <FormattedMessage id="ScreenHeader.projectSelector.current-project" />
                </div>
                <div className={styles.currentProjectRow}>
                  <span className={styles.currentProjectID}>
                    {displayAppID}
                  </span>
                  <ProjectSelectorCopyButton
                    projectID={displayAppID}
                    buttonStyles={copyIconButtonStyles}
                  />
                </div>
              </section>
              <hr className={styles.divider} />
              <section className={styles.section}>
                <div className={styles.sectionHeader}>
                  <FormattedMessage id="ScreenHeader.projectSelector.switch-project" />
                </div>
                <div className={styles.projectList}>
                  {otherApps.map((app) => (
                    <button
                      key={app.appID}
                      type="button"
                      className={styles.projectListItem}
                      onClick={() => onSelectProject(app.appID)}
                    >
                      <span className={styles.projectListItemLabel}>
                        {app.appID}
                      </span>
                    </button>
                  ))}
                </div>
              </section>
              {!isAuthgearOnce ? (
                <>
                  <hr className={styles.divider} />
                  <section
                    className={`${styles.section} ${styles.sectionCreateProject}`}
                  >
                    <button
                      type="button"
                      className={styles.createProjectButton}
                      onClick={onCreateProject}
                      disabled={createButtonDisabled}
                    >
                      <Icon
                        className={styles.createProjectIcon}
                        iconName="Add"
                        styles={{
                          root: { color: accentColor, fontSize: 16 },
                        }}
                      />
                      <FormattedMessage id="ScreenHeader.projectSelector.create-project" />
                    </button>
                  </section>
                </>
              ) : null}
            </div>
          </Callout>
        ) : null}
      </>
    );
  };

export default ProjectSelector;
