import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useNavigate } from "react-router-dom";
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
        data-project-selector-copy={true}
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

const ProjectSelector: React.VFC<ProjectSelectorProps> = function ProjectSelector(
  props
) {
  const { appID, theme } = props;
  const { renderToString } = useContext(Context);
  const navigate = useNavigate();
  const capture = useCapture();
  const triggerRef = useRef<HTMLButtonElement>(null);
  const calloutPanelRef = useRef<HTMLDivElement>(null);
  const [isCalloutOpen, setIsCalloutOpen] = useState(false);

  const { effectiveAppConfig, isLoading: loadingAppConfig } =
    useAppAndSecretConfigQuery(appID);
  const { apps, loading: loadingAppList } = useAppListQuery();
  const { viewer } = useViewerQuery();
  const { themes, authgearAppID, isAuthgearOnce } = useSystemConfig();

  const accentColor = themes.main.palette.themePrimary;
  const accentHoverColor =
    themes.main.palette.themeDark ?? themes.main.palette.themePrimary;

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
      }) as React.CSSProperties,
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

  const onToggleCallout = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setIsCalloutOpen((open) => !open);
    },
    []
  );

  const isInsideProjectSelector = useCallback((target: EventTarget | null) => {
    if (!(target instanceof Node)) {
      return false;
    }
    if (triggerRef.current?.contains(target)) {
      return true;
    }
    if (calloutPanelRef.current?.contains(target)) {
      return true;
    }
    if (
      target instanceof HTMLElement &&
      target.closest("[data-project-selector-panel]") != null
    ) {
      return true;
    }
    if (target instanceof HTMLElement) {
      const copyButton = document.querySelector<HTMLButtonElement>(
        "[data-project-selector-copy] button"
      );
      const describedBy = copyButton?.getAttribute("aria-describedby");
      if (describedBy != null) {
        const tooltipEl = document.getElementById(describedBy);
        if (
          tooltipEl != null &&
          (tooltipEl.contains(target) || target.closest(`#${describedBy}`) != null)
        ) {
          return true;
        }
      }
    }
    return false;
  }, []);

  useEffect(() => {
    if (!isCalloutOpen) {
      return undefined;
    }

    const onPointerDown = (event: PointerEvent) => {
      if (isInsideProjectSelector(event.target)) {
        return;
      }
      setIsCalloutOpen(false);
    };

    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setIsCalloutOpen(false);
      }
    };

    document.addEventListener("pointerdown", onPointerDown, true);
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.removeEventListener("pointerdown", onPointerDown, true);
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [isCalloutOpen, isInsideProjectSelector]);

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
      navigate(`/project/${encodeURIComponent(typedID)}/getting-started`);
      setIsCalloutOpen(false);
    },
    [capture, displayAppID, navigate]
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
        >
          <div
            ref={calloutPanelRef}
            data-project-selector-panel={true}
            onMouseDown={(e) => e.stopPropagation()}
            onClick={(e) => e.stopPropagation()}
          >
          <section className={styles.section}>
            <div className={styles.sectionHeader}>
              <FormattedMessage id="ScreenHeader.projectSelector.current-project" />
            </div>
            <div className={styles.currentProjectRow}>
              <span className={styles.currentProjectID}>{displayAppID}</span>
              <ProjectSelectorCopyButton
                projectID={displayAppID}
                buttonStyles={copyIconButtonStyles}
              />
            </div>
          </section>
          <hr className={styles.divider} />
          <section className={styles.section}>
            <div className={styles.sectionHeader}>
              <FormattedMessage id="ScreenHeader.projectSelector.your-projects" />
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
