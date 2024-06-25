import React, { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import ShowLoading from "./ShowLoading";
import ShowError from "./ShowError";
import { useAppListQuery } from "./graphql/portal/query/appListQuery";
import { useViewerQuery } from "./graphql/portal/query/viewerQuery";

const OnboardingRedirect: React.VFC = function OnboardingRedirect() {
  const {
    apps,
    loading: loadingAppList,
    error: errorAppList,
    refetch: refetchAppList,
  } = useAppListQuery();
  const {
    viewer,
    loading: loadingViewer,
    error: errorViewer,
    refetch: refetchViewer,
  } = useViewerQuery();
  const navigate = useNavigate();

  useEffect(() => {
    if (loadingAppList || loadingViewer) {
      return;
    }
    if (errorAppList != null || errorViewer != null) {
      return;
    }
    if (viewer === undefined || viewer === null) {
      return;
    }
    if (
      (apps === null || apps.length === 0) &&
      !viewer.isOnboardingSurveyCompleted
    ) {
      navigate("/onboarding-survey");
    } else {
      navigate("/");
    }
  }, [
    navigate,
    viewer,
    apps,
    loadingAppList,
    loadingViewer,
    errorAppList,
    errorViewer,
  ]);

  if (loadingAppList || loadingViewer) {
    return <ShowLoading />;
  }

  if (errorAppList != null) {
    return <ShowError error={errorAppList} onRetry={refetchAppList} />;
  }
  if (errorViewer != null) {
    return <ShowError error={errorViewer} onRetry={refetchViewer} />;
  }

  return null;
};

export default OnboardingRedirect;
