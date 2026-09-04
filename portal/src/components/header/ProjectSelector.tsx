import React, { useCallback, useContext, useMemo, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Popover, Text } from "@radix-ui/themes";
import { CaretSortIcon, PlusIcon } from "@radix-ui/react-icons";
import { Context, FormattedMessage } from "../../intl";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { useAppListQuery } from "../../graphql/portal/query/appListQuery";
import { useViewerQuery } from "../../graphql/portal/query/viewerQuery";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useCapture } from "../../gtm_v2";
import { toTypedID } from "../../util/graphql";
import { isProjectQuotaReached } from "../../util/projectQuota";
import { CopyIconButton } from "../v2/CopyIconButton/CopyIconButton";
import styles from "./ProjectSelector.module.css";

interface ProjectSelectorProps {
  appID: string;
}

const ProjectSelector: React.VFC<ProjectSelectorProps> =
  function ProjectSelector(props) {
    const { appID } = props;
    const { renderToString } = useContext(Context);
    const navigate = useNavigate();
    const location = useLocation();
    const capture = useCapture();
    const [isOpen, setIsOpen] = useState(false);

    const { effectiveAppConfig, isLoading: loadingAppConfig } =
      useAppAndSecretConfigQuery(appID);
    const { apps, loading: loadingAppList } = useAppListQuery();
    const { viewer } = useViewerQuery();
    const { authgearAppID, isAuthgearOnce } = useSystemConfig();

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

    const onSelectProject = useCallback(
      (selectedAppID: string) => {
        if (selectedAppID === displayAppID) {
          setIsOpen(false);
          return;
        }
        capture(
          "enteredProject",
          { projectID: selectedAppID },
          { project_id: selectedAppID }
        );
        const typedID = toTypedID("App", selectedAppID);
        const newProjectBasePath = `/project/${encodeURIComponent(typedID)}`;
        // Keep the user on the same page under the new project by swapping the
        // /project/:id prefix. Screens that load project-scoped entities are
        // responsible for redirecting to their list page when the entity does
        // not exist in the new project. Search and hash are intentionally
        // dropped as they typically reference data from the previous project.
        const nextPathname = location.pathname.replace(
          /^\/project\/[^/]+/,
          newProjectBasePath
        );

        navigate(nextPathname);
        setIsOpen(false);
      },
      [capture, displayAppID, location.pathname, navigate]
    );

    const onCreateProject = useCallback(() => {
      if (createButtonDisabled) {
        return;
      }
      navigate("/projects/create");
      setIsOpen(false);
    }, [createButtonDisabled, navigate]);

    if (loadingAppConfig || loadingAppList) {
      return null;
    }

    const openMenuLabel = renderToString(
      "ScreenHeader.projectSelector.open-menu"
    );

    return (
      <Popover.Root open={isOpen} onOpenChange={setIsOpen}>
        <Popover.Trigger>
          <button
            type="button"
            className={styles.trigger}
            aria-label={openMenuLabel}
          >
            <Text className={styles.triggerLabel}>{displayAppID}</Text>
            <CaretSortIcon className={styles.triggerChevron} />
          </button>
        </Popover.Trigger>
        <Popover.Content
          align="start"
          sideOffset={4}
          // Don't auto-focus the first item (the copy button) on open, which
          // would otherwise trigger its focus tooltip ("Copy") without the
          // user hovering. Focus stays on the trigger; Tab still enters.
          onOpenAutoFocus={(e) => e.preventDefault()}
          // The copy button copies via a temporary <textarea> appended to
          // <body>, whose .select() steals focus out of the popover. Without
          // this, that focus-out would dismiss the popover on every copy.
          // Click-outside and Escape still close it.
          onFocusOutside={(e) => e.preventDefault()}
          // Match the metrics of a Radix size-2 menu (e.g. the avatar
          // DropdownMenu) so the two popovers read consistently.
          style={{
            minWidth: 200,
            maxWidth: 280,
            padding: "var(--space-2)",
            borderRadius: "var(--radius-4)",
          }}
        >
          <section className={styles.section}>
            <div className={styles.sectionHeader}>
              <FormattedMessage id="ScreenHeader.projectSelector.current-project" />
            </div>
            <div className={styles.currentProjectRow}>
              <span className={styles.currentProjectID}>{displayAppID}</span>
              <CopyIconButton textToCopy={displayAppID} />
            </div>
          </section>
          {otherApps.length > 0 ? (
            <>
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
            </>
          ) : null}
          {!isAuthgearOnce ? (
            <>
              <hr className={styles.divider} />
              <section className={styles.section}>
                <button
                  type="button"
                  className={styles.createProjectButton}
                  onClick={onCreateProject}
                  disabled={createButtonDisabled}
                >
                  <PlusIcon className={styles.createProjectIcon} />
                  <FormattedMessage id="ScreenHeader.projectSelector.create-project" />
                </button>
              </section>
            </>
          ) : null}
        </Popover.Content>
      </Popover.Root>
    );
  };

export default ProjectSelector;
