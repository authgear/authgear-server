import React from "react";
import { useParams } from "react-router-dom";
import { AppFeature } from "./appFeatures";
import { useAppFeatures } from "./hook/useAppFeatures";

interface AppFeatureGateProps {
  feature: AppFeature;
  children: React.ReactNode;
  // Rendered while loading, on error, or when the feature is unavailable.
  fallback?: React.ReactNode;
}

// AppFeatureGate renders children only when the feature is enabled for the
// current project. For pages, add the path to APP_FEATURES instead.
const AppFeatureGate: React.VFC<AppFeatureGateProps> = function AppFeatureGate({
  feature,
  children,
  fallback = null,
}) {
  const { appID } = useParams() as { appID: string };
  const { isEnabled } = useAppFeatures(appID);
  return <>{isEnabled(feature) ? children : fallback}</>;
};

export default AppFeatureGate;
