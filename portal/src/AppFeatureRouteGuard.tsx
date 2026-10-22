import React from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "./intl";
import { AppFeature } from "./appFeatures";
import { useAppFeatures } from "./hook/useAppFeatures";
import ScreenContent from "./ScreenContent";
import { Callout } from "./components/v2/Callout/Callout";
import ShowLoading from "./ShowLoading";
import ShowError from "./ShowError";
import styles from "./AppFeatureRouteGuard.module.css";

interface AppFeatureRouteGuardProps {
  feature: AppFeature;
  children: React.ReactNode;
}

// AppFeatureRouteGuard renders a route's element only when its feature is
// available. Otherwise, e.g. the URL was entered directly, it shows a message
// instead. The screen is never mounted, so it cannot load or save config for
// an unavailable feature. Use it through `feature` on AppRoute.
const AppFeatureRouteGuard: React.VFC<AppFeatureRouteGuardProps> =
  function AppFeatureRouteGuard({ feature, children }) {
    const { appID } = useParams() as { appID: string };
    const features = useAppFeatures(appID);

    if (features.isLoading) {
      return <ShowLoading />;
    }

    if (features.loadError != null) {
      return <ShowError error={features.loadError} onRetry={features.reload} />;
    }

    if (!features.isEnabled(feature)) {
      return (
        <ScreenContent>
          <Callout
            className={styles.widget}
            type="warning"
            showCloseButton={false}
            text={<FormattedMessage id="AppFeatureRouteGuard.not-available" />}
          />
        </ScreenContent>
      );
    }

    return <>{children}</>;
  };

export default AppFeatureRouteGuard;
