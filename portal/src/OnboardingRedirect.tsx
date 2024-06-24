import React, { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import ShowLoading from "./ShowLoading";
import ShowError from "./ShowError";
import { useAppListQuery } from "./graphql/portal/query/appListQuery";
import { useViewerQuery } from "./graphql/portal/query/viewerQuery";

const OnboardingRedirect: React.VFC = function OnboardingRedirect() {
  const { loading, error, apps, refetch } = useAppListQuery();

  const navigate = useNavigate();
  const { viewer } = useViewerQuery();

  useEffect(() => {
    if (loading) {
      return;
    }
    if (error != null) {
      return;
    }
    if (
      (apps === null || apps.length === 0) &&
      !viewer?.isOnboardingSurveyCompleted
    ) {
      navigate("/onboarding-survey");
    } else {
      navigate("/");
    }
  }, [navigate, viewer, error, apps, loading]);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  return null;
};

export default OnboardingRedirect;
