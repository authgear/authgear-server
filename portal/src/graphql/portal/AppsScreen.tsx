import React, { useCallback, useMemo, useEffect } from "react";
import { Link, useNavigate } from "react-router-dom";
import { FormattedMessage } from "@oursky/react-messageformat";
import { DefaultEffects, Text } from "@fluentui/react";
import PrimaryButton from "../../PrimaryButton";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import ScreenHeader from "../../ScreenHeader";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import BlueMessageBar from "../../BlueMessageBar";
import { useAppListQuery } from "./query/appListQuery";
import { useViewerQuery } from "./query/viewerQuery";
import { AppListItem, Viewer } from "./globalTypes.generated";
import styles from "./AppsScreen.module.css";
import { useCapture } from "../../gtm_v2";
import { toTypedID } from "../../util/graphql";

interface AppCardData {
  appName: string;
  appID: string;
  url: string;
}

const AppCard: React.VFC<AppCardData> = function AppCard(props: AppCardData) {
  const { appName, appID, url } = props;
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
    <Link
      to={url}
      style={{ boxShadow: DefaultEffects.elevation4 }}
      className={styles.card}
      onClick={onClick}
    >
      <Text className={styles.cardAppID}>{appID}</Text>
      <Text className={styles.cardAppName}>{appName}</Text>
    </Link>
  );
};

function isProjectQuotaReached(viewer: Viewer | null): boolean {
  if (viewer == null) {
    return false;
  }
  const { projectQuota, projectOwnerCount } = viewer;
  // The viewer does not have quota.
  if (projectQuota == null) {
    return false;
  }

  const reached = projectOwnerCount >= projectQuota;
  return reached;
}

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
    <BlueMessageBar>
      <FormattedMessage id="AppsScreen.project-quota-reached" />
    </BlueMessageBar>
  );
}

interface AppListProps {
  apps: AppListItem[] | null;
  viewer: Viewer | null;
}

const AppList: React.VFC<AppListProps> = function AppList(props: AppListProps) {
  const { apps, viewer } = props;
  const projectQuotaReached = isProjectQuotaReached(viewer);
  const navigate = useNavigate();

  useEffect(() => {
    if (
      (apps === null || apps.length === 0) &&
      !viewer?.isOnboardingSurveyCompleted
    ) {
      navigate("/onboarding-survey");
    }
  }, [apps, viewer, navigate]);

  const onCreateClick = useCallback(
    (e) => {
      e?.preventDefault();
      e?.stopPropagation();
      navigate("/projects/create");
    },
    [navigate]
  );

  const appCardsData: AppCardData[] = useMemo(() => {
    return (apps ?? []).map((app) => {
      const appID = app.appID;
      const appOrigin = app.publicOrigin;
      const typedID = toTypedID("App", appID);
      const relPath = "/project/" + encodeURIComponent(typedID);
      return {
        appID,
        appName: appOrigin,
        url: relPath,
      };
    });
  }, [apps]);

  return (
    <main className={styles.root}>
      <ScreenHeader showHamburger={false} />
      <ScreenLayoutScrollView>
        <section className={styles.body}>
          <Text as="h1" variant="xLarge" block={true}>
            <FormattedMessage id="AppsScreen.title" />
          </Text>
          <section className={styles.cardsContainer}>
            {appCardsData.map((appCardData) => {
              return <AppCard key={appCardData.appID} {...appCardData} />;
            })}
          </section>
          <div className="space-y-4">
            <PrimaryButton
              className={styles.createButton}
              onClick={onCreateClick}
              text={<FormattedMessage id="AppsScreen.create-app" />}
              disabled={projectQuotaReached}
            />
            <ProjectQuotaMessageBar viewer={viewer} />
          </div>
        </section>
      </ScreenLayoutScrollView>
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

  if (loadingViewer || loadingAppList) {
    return <ShowLoading />;
  }

  if (errorViewer != null) {
    return <ShowError error={errorViewer} onRetry={refetchViewer} />;
  }

  if (errorAppList != null) {
    return <ShowError error={errorAppList} onRetry={refetchAppList} />;
  }

  return <AppList apps={apps ?? null} viewer={viewer ?? null} />;
};

export default AppsScreen;
