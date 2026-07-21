import React from "react";
import { Navigate, Outlet, useParams } from "react-router-dom";
import { useAppFeatureConfigQuery } from "./graphql/portal/query/appFeatureConfigQuery";
import { PortalAPIFeatureConfig } from "./types";
import ShowLoading from "./ShowLoading";
import ShowError from "./ShowError";

export interface RequireAppFeatureProps {
  // isAvailable returns whether the current project may access the gated
  // section, given its effective feature config.
  isAvailable: (featureConfig: PortalAPIFeatureConfig | null) => boolean;
}

// RequireAppFeature is a route-level guard for pages that only some projects can
// access based on their feature config (e.g. fraud protection, app2app,
// integrations). When the project lacks the feature it redirects to getting
// started; otherwise it renders the matched child route via <Outlet>.
//
// It deliberately renders the guarded screen through <Outlet> only once access
// is confirmed, so the screen (and any hook it uses that rewrites the URL, such
// as usePivotNavigation) never mounts on the redirect path and cannot override
// the redirect.
const RequireAppFeature: React.VFC<RequireAppFeatureProps> =
  function RequireAppFeature({ isAvailable }) {
    const { appID } = useParams() as { appID: string };
    const featureConfig = useAppFeatureConfigQuery(appID);

    if (featureConfig.isLoading) {
      return <ShowLoading />;
    }

    if (featureConfig.loadError) {
      return (
        <ShowError
          error={featureConfig.loadError}
          onRetry={featureConfig.reload}
        />
      );
    }

    if (!isAvailable(featureConfig.effectiveFeatureConfig)) {
      return (
        <Navigate to={`/project/${appID}/getting-started`} replace={true} />
      );
    }

    return <Outlet />;
  };

export default RequireAppFeature;
