import React, {
  useCallback,
  useMemo,
  useEffect,
  useId,
  useRef,
  useState,
} from "react";
import cn from "classnames";
import { Link, Navigate, useNavigate } from "react-router-dom";
import { Heading, Text } from "@radix-ui/themes";
import { FormattedMessage, Context } from "../../intl";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import {
  TextField,
  TextFieldIcon,
} from "../../components/v2/TextField/TextField";
import { Callout } from "../../components/v2/Callout/Callout";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenHeader from "../../ScreenHeader";
import ExternalLink from "../../ExternalLink";
import { useAppListQuery } from "./query/appListQuery";
import { useViewerQuery } from "./query/viewerQuery";
import { AppListItem, Viewer } from "./globalTypes.generated";
import { isProjectQuotaReached } from "../../util/projectQuota";
import { makeHelpDeskURL } from "../../util/helpDesk";
import styles from "./AppsScreen.module.css";
import { useCapture } from "../../gtm_v2";
import { toTypedID } from "../../util/graphql";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { shouldShowSurvey } from "../../util/survey";

interface AppRowData {
  appID: string;
  publicOrigin: string;
  url: string;
}

function useCaptureEnteredProject(): (appID: string) => void {
  const capture = useCapture();
  return useCallback(
    (appID: string) => {
      capture(
        "enteredProject",
        {
          projectID: appID,
        },
        {
          project_id: appID,
        }
      );
    },
    [capture]
  );
}

interface AppRowProps extends AppRowData {
  optionID: string;
  index: number;
  active: boolean;
  onActivate: (index: number) => void;
  rowRef: (index: number, el: HTMLAnchorElement | null) => void;
}

const AppRow: React.VFC<AppRowProps> = function AppRow(props: AppRowProps) {
  const {
    appID,
    publicOrigin,
    url,
    optionID,
    index,
    active,
    onActivate,
    rowRef,
  } = props;
  const captureEnteredProject = useCaptureEnteredProject();
  const onClick = useCallback(() => {
    captureEnteredProject(appID);
  }, [appID, captureEnteredProject]);
  // mousemove rather than mouseenter, so a row scrolled under a resting
  // pointer by the arrow keys does not take the highlight back.
  const onMouseMove = useCallback(() => {
    if (!active) {
      onActivate(index);
    }
  }, [active, index, onActivate]);
  const ref = useCallback(
    (el: HTMLAnchorElement | null) => {
      rowRef(index, el);
    },
    [index, rowRef]
  );

  return (
    // An option of the search combobox's listbox, so it is reached with the
    // arrow keys from the search input rather than with Tab.
    <Link
      ref={ref}
      id={optionID}
      role="option"
      aria-selected={active}
      tabIndex={-1}
      to={url}
      className={cn(styles.row, active ? styles.rowActive : null)}
      onClick={onClick}
      onMouseMove={onMouseMove}
    >
      <Text as="p" size="2" weight="medium" className={styles.rowAppID}>
        {appID}
      </Text>
      <Text as="p" size="2" truncate={true} className={styles.rowOrigin}>
        {publicOrigin}
      </Text>
    </Link>
  );
};

interface ProjectQuotaMessageBarProps {
  viewer: Viewer | null;
}

function ProjectQuotaMessageBar(
  props: ProjectQuotaMessageBarProps
): React.ReactElement | null {
  const { viewer } = props;
  const reached = isProjectQuotaReached(viewer);
  if (!reached) {
    return null;
  }
  const helpDeskURL = makeHelpDeskURL({
    purpose: "project-limit",
    email: viewer?.email,
    fullName: viewer?.formattedName,
  });
  return (
    <div className={styles.quota}>
      <Callout
        type="info"
        showCloseButton={false}
        text={
          <FormattedMessage
            id="AppsScreen.project-quota-reached"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              externalLink: (chunks: React.ReactNode) => (
                <ExternalLink href={helpDeskURL}>{chunks}</ExternalLink>
              ),
            }}
          />
        }
      />
    </div>
  );
}

interface AppListPanelProps {
  apps: AppListItem[];
  viewer: Viewer;
  isAuthgearOnce: boolean;
  onCreateClick: (e: React.MouseEvent<HTMLButtonElement>) => void;
}

// AppListPanel is the presentational part of the projects page: the bordered
// panel holding the title, the sticky search/create toolbar and the rows. It
// does no data fetching of its own.
const AppListPanel: React.VFC<AppListPanelProps> = function AppListPanel(
  props: AppListPanelProps
) {
  const { apps, viewer, isAuthgearOnce, onCreateClick } = props;
  const { renderToString } = React.useContext(Context);
  const projectQuotaReached = isProjectQuotaReached(viewer);
  const createButtonDisabled = projectQuotaReached || isAuthgearOnce;

  const [searchKeyword, setSearchKeyword] = useState("");

  const appRowsData: AppRowData[] = useMemo(() => {
    // The API returns the list sorted A-Z by app ID.
    return apps.map((app) => {
      const appID = app.appID;
      const typedID = toTypedID("App", appID);
      const relPath = "/project/" + encodeURIComponent(typedID);
      return {
        appID,
        publicOrigin: app.publicOrigin,
        url: relPath,
      };
    });
  }, [apps]);

  const filteredAppRowsData = useMemo(() => {
    const keyword = searchKeyword.trim().toLowerCase();
    if (keyword === "") {
      return appRowsData;
    }
    return appRowsData.filter(
      (a) =>
        a.appID.toLowerCase().includes(keyword) ||
        a.publicOrigin.toLowerCase().includes(keyword)
    );
  }, [appRowsData, searchKeyword]);

  // The row highlighted by the arrow keys or the pointer; -1 is none.
  const [activeIndex, setActiveIndex] = useState(-1);
  const rowEls = useRef<Map<number, HTMLAnchorElement>>(new Map());
  // Only the arrow keys scroll the highlight into view; a hovered row is
  // already under the pointer, and scrolling it would move the list away.
  const scrollToActive = useRef(false);
  const navigate = useNavigate();
  const captureEnteredProject = useCaptureEnteredProject();
  const titleID = useId();
  const listboxID = useId();
  const optionID = useCallback(
    (index: number) => `${listboxID}-${index}`,
    [listboxID]
  );
  const hasOptions = filteredAppRowsData.length > 0;

  const rowRef = useCallback((index: number, el: HTMLAnchorElement | null) => {
    if (el == null) {
      rowEls.current.delete(index);
    } else {
      rowEls.current.set(index, el);
    }
  }, []);

  useEffect(() => {
    if (scrollToActive.current) {
      scrollToActive.current = false;
      rowEls.current.get(activeIndex)?.scrollIntoView({ block: "nearest" });
    }
  }, [activeIndex]);

  const onChangeSearchKeyword = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchKeyword(e.currentTarget.value);
      setActiveIndex(-1);
    },
    []
  );

  const onSearchKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      let next: number;
      switch (e.key) {
        case "ArrowDown":
          e.preventDefault();
          next = Math.min(activeIndex + 1, filteredAppRowsData.length - 1);
          break;
        case "ArrowUp":
          e.preventDefault();
          next = Math.max(activeIndex - 1, -1);
          break;
        case "Escape":
          setActiveIndex(-1);
          return;
        case "Enter": {
          // Skip while an IME is composing; Enter commits the composition.
          if (e.nativeEvent.isComposing) {
            return;
          }
          const row = filteredAppRowsData[activeIndex] as
            | AppRowData
            | undefined;
          if (row != null) {
            e.preventDefault();
            captureEnteredProject(row.appID);
            navigate(row.url);
          }
          return;
        }
        default:
          return;
      }
      if (next !== activeIndex) {
        scrollToActive.current = true;
        setActiveIndex(next);
      }
    },
    [filteredAppRowsData, activeIndex, captureEnteredProject, navigate]
  );

  const onListMouseLeave = useCallback(() => {
    setActiveIndex(-1);
  }, []);

  return (
    <section className={styles.body}>
      <div className={styles.panel}>
        <Heading
          id={titleID}
          as="h1"
          size="6"
          weight="bold"
          className={styles.panelTitle}
        >
          <FormattedMessage id="AppsScreen.title" />
        </Heading>
        <div className={styles.toolbar}>
          <div className={styles.search}>
            <TextField
              size="3"
              iconStart={TextFieldIcon.MagnifyingGlass}
              placeholder={renderToString("AppsScreen.search.placeholder")}
              value={searchKeyword}
              onChange={onChangeSearchKeyword}
              onKeyDown={onSearchKeyDown}
              role="combobox"
              aria-expanded={hasOptions}
              aria-controls={listboxID}
              aria-autocomplete="list"
              aria-activedescendant={
                activeIndex >= 0 ? optionID(activeIndex) : undefined
              }
            />
          </div>
          {!isAuthgearOnce ? (
            <PrimaryButton
              size="3"
              onClick={onCreateClick}
              text={<FormattedMessage id="AppsScreen.create-app" />}
              disabled={createButtonDisabled}
            />
          ) : null}
        </div>
        {!isAuthgearOnce ? <ProjectQuotaMessageBar viewer={viewer} /> : null}
        <section
          id={listboxID}
          role={hasOptions ? "listbox" : undefined}
          aria-labelledby={titleID}
          className={styles.list}
          onMouseLeave={onListMouseLeave}
        >
          {hasOptions ? (
            filteredAppRowsData.map((appRowData, index) => {
              return (
                <AppRow
                  key={appRowData.appID}
                  {...appRowData}
                  optionID={optionID(index)}
                  index={index}
                  active={index === activeIndex}
                  onActivate={setActiveIndex}
                  rowRef={rowRef}
                />
              );
            })
          ) : (
            <Text as="p" size="2" color="gray" className={styles.emptyText}>
              <FormattedMessage
                id={
                  searchKeyword.trim() === ""
                    ? "AppsScreen.no-projects"
                    : "AppsScreen.no-search-results"
                }
              />
            </Text>
          )}
        </section>
      </div>
    </section>
  );
};

interface AppListProps {
  apps: AppListItem[] | null;
  viewer: Viewer;
}

const AppList: React.VFC<AppListProps> = function AppList(props: AppListProps) {
  const { apps: unfilteredApps, viewer } = props;
  const navigate = useNavigate();
  const systemConfig = useSystemConfig();
  const { authgearAppID, isAuthgearOnce } = systemConfig;

  const onCreateClick = useCallback(
    (e: React.MouseEvent<HTMLButtonElement>) => {
      e.preventDefault();
      e.stopPropagation();
      navigate("/projects/create");
    },
    [navigate]
  );

  const apps = useMemo(() => {
    return (unfilteredApps ?? []).filter((a) => {
      if (isAuthgearOnce && a.appID === authgearAppID) {
        return false;
      }
      return true;
    });
  }, [unfilteredApps, isAuthgearOnce, authgearAppID]);

  useEffect(() => {
    if (shouldShowSurvey(systemConfig, apps, viewer)) {
      navigate("/onboarding-survey");
    }
  }, [
    apps.length,
    viewer.isOnboardingSurveyCompleted,
    navigate,
    systemConfig,
    apps,
    viewer,
  ]);

  if (isAuthgearOnce && apps.length === 1) {
    return (
      <Navigate
        to={`/project/${encodeURIComponent(toTypedID("App", apps[0].appID))}`}
        replace={true}
      />
    );
  }

  return (
    <main className={styles.root}>
      <ScreenHeader showHamburger={false} />
      <div className={styles.content}>
        <AppListPanel
          apps={apps}
          viewer={viewer}
          isAuthgearOnce={isAuthgearOnce}
          onCreateClick={onCreateClick}
        />
      </div>
    </main>
  );
};

const AppsScreen: React.VFC = function AppsScreen() {
  const {
    viewer,
    loading: loadingViewer,
    error: errorViewer,
    refetch: refetchViewer,
  } = useViewerQuery();

  const {
    apps,
    loading: loadingAppList,
    error: errorAppList,
    refetch: refetchAppList,
  } = useAppListQuery();

  // If viewer is null, <Authenticated> will redirect to login screen.
  if (loadingViewer || loadingAppList || viewer == null) {
    return <ShowLoading />;
  }

  if (errorViewer != null) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
    return <ShowError error={errorViewer} onRetry={refetchViewer} />;
  }

  if (errorAppList != null) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
    return <ShowError error={errorAppList} onRetry={refetchAppList} />;
  }

  return <AppList apps={apps ?? null} viewer={viewer} />;
};

export default AppsScreen;
