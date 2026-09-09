import React, { useCallback, useMemo, useEffect, useState } from "react";
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

const AppRow: React.VFC<AppRowData> = function AppRow(props: AppRowData) {
  const { appID, publicOrigin, url } = props;
  const capture = useCapture();
  const onClick = useCallback(() => {
    capture(
      "enteredProject",
      {
        projectID: appID,
      },
      {
        project_id: appID,
      }
    );
  }, [appID, capture]);

  return (
    <Link to={url} className={styles.row} onClick={onClick}>
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
                <ExternalLink href="https://go.authgear.com/portal-support">
                  {chunks}
                </ExternalLink>
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
    // The API returns projects in no particular order; sort them A-Z by ID,
    // which is the leading column of each row.
    return [...apps]
      .sort((a, b) => a.appID.localeCompare(b.appID))
      .map((app) => {
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

  const onChangeSearchKeyword = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearchKeyword(e.currentTarget.value);
    },
    []
  );

  return (
    <section className={styles.body}>
      <div className={styles.panel}>
        <Heading as="h1" size="6" weight="bold" className={styles.panelTitle}>
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
        <section className={styles.list}>
          {filteredAppRowsData.length > 0 ? (
            filteredAppRowsData.map((appRowData) => {
              return <AppRow key={appRowData.appID} {...appRowData} />;
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
