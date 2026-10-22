import { useCallback } from "react";
import { useAppFeatureConfigQuery } from "../graphql/portal/query/appFeatureConfigQuery";
import { useSystemConfig } from "../context/SystemConfigContext";
import { APP_FEATURES, AppFeature } from "../appFeatures";
import { Loadable } from "./useLoadableView";

interface AppFeatures extends Loadable {
  // isEnabled is false while loading or on error, so a gated feature is never
  // shown by accident.
  isEnabled: (feature: AppFeature) => boolean;
}

export function useAppFeatures(appID: string): AppFeatures {
  const { effectiveFeatureConfig, isLoading, loadError, reload } =
    useAppFeatureConfigQuery(appID);
  const systemConfig = useSystemConfig();

  const isEnabled = useCallback(
    (feature: AppFeature) => {
      if (isLoading || loadError != null) {
        return false;
      }
      return APP_FEATURES[feature].isAvailable({
        featureConfig: effectiveFeatureConfig,
        systemConfig,
      });
    },
    [effectiveFeatureConfig, isLoading, loadError, systemConfig]
  );

  return { isLoading, loadError, reload, isEnabled };
}
