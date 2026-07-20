import React, { useContext, useMemo, useState } from "react";
import { useLocation, useParams } from "react-router-dom";
import { useResourceQueryQuery } from "../../graphql/adminapi/query/resourceQuery.generated";
import { useLoadableView } from "../../hook/useLoadableView";
import { FormattedMessage, Context as MessageContext } from "../../intl";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { OverflowTabs } from "../../components/v2/OverflowTabs/OverflowTabs";
import { usePivotNavigation } from "../../hook/usePivot";
import { APIResourceDetailsScreenDetailsTab } from "./APIResourceDetailsScreenDetailsTab";
import { APIResourceDetailsScreenScopesTab } from "./APIResourceDetailsScreenScopesTab";
import { APIResourceDetailsScreenApplicationsTab } from "./APIResourceDetailsScreenApplicationsTab";
import { APIResourceDetailsScreenTestTab } from "./APIResourceDetailsScreenTestTab";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "../../graphql/portal/mutations/generateAppSecretVisitTokenMutation";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { AppSecretKey } from "../../graphql/portal/globalTypes.generated";
import { PortalAPIAppConfig, PortalAPISecretConfig } from "../../types";
import styles from "./APIResourceDetailsScreen.module.css";

export interface LocationState {
  isClientSecretRevealed: boolean;
}

const SECRETS = [AppSecretKey.OauthClientSecrets];
function isLocationState(raw: unknown): raw is LocationState {
  return (
    raw != null &&
    typeof raw === "object" &&
    (raw as Partial<LocationState>).isClientSecretRevealed != null
  );
}

const TAB_KEYS = ["details", "scopes", "applications", "test"] as const;

function APIResourceDetailsContent({
  resource,
  effectiveAppConfig,
  secretConfig,
}: {
  resource: Resource;
  effectiveAppConfig: PortalAPIAppConfig;
  secretConfig: PortalAPISecretConfig | null;
}) {
  const { selectedKey, onChangeKey } = usePivotNavigation([...TAB_KEYS]);
  const { renderToString } = useContext(MessageContext);

  const tabs = useMemo(
    () => [
      {
        value: "details",
        label: renderToString("APIResourceDetailsScreen.tab.details"),
      },
      {
        value: "scopes",
        label: renderToString("APIResourceDetailsScreen.tab.scopes"),
      },
      {
        value: "applications",
        label: renderToString("APIResourceDetailsScreen.tab.applications"),
      },
      {
        value: "test",
        label: renderToString("APIResourceDetailsScreen.tab.test"),
      },
    ],
    [renderToString]
  );

  return (
    <div className={styles.content}>
      <OverflowTabs
        value={selectedKey}
        onValueChange={(value) => {
          onChangeKey(value as (typeof TAB_KEYS)[number]);
        }}
        listClassName={styles.tabsList}
        tabs={tabs}
      />
      {selectedKey === "details" ? (
        <APIResourceDetailsScreenDetailsTab resource={resource} />
      ) : null}
      {selectedKey === "scopes" ? (
        <APIResourceDetailsScreenScopesTab resource={resource} />
      ) : null}
      {selectedKey === "applications" ? (
        <APIResourceDetailsScreenApplicationsTab
          resource={resource}
          effectiveAppConfig={effectiveAppConfig}
        />
      ) : null}
      {selectedKey === "test" ? (
        <APIResourceDetailsScreenTestTab
          resource={resource}
          effectiveAppConfig={effectiveAppConfig}
          secretConfig={secretConfig}
        />
      ) : null}
    </div>
  );
}

const APIResourceDetailsScreen: React.VFC =
  function APIResourceDetailsScreen() {
    const { appID, resourceID } = useParams<{
      resourceID: string;
      appID: string;
    }>();
    const { data, loading, error, refetch } = useResourceQueryQuery({
      variables: { id: resourceID! },
    });
    const location = useLocation();
    const [shouldRefreshToken] = useState<boolean>(() => {
      const { state } = location;
      if (isLocationState(state) && state.isClientSecretRevealed) {
        return true;
      }
      return false;
    });
    useLocationEffect<LocationState>(() => {
      // Pop the location state if exist
    });
    const appSecretTokenQuery = useAppSecretVisitToken(
      appID!,
      SECRETS,
      shouldRefreshToken
    );
    const appConfigQuery = useAppAndSecretConfigQuery(
      appID!,
      appSecretTokenQuery.token
    );
    const appSecretTokenLoadable = {
      isLoading: appSecretTokenQuery.loading,
      reload: appSecretTokenQuery.retry,
      loadError: appSecretTokenQuery.error,
    };

    return useLoadableView({
      loadables: [
        {
          isLoading: loading,
          loadError: error,
          reload: refetch,
          data: data,
        },
        appConfigQuery,
        appSecretTokenLoadable,
      ] as const,
      render: ([resourceQuery, configQuery]) => {
        const { data } = resourceQuery;
        const resource =
          data?.node?.__typename === "Resource" ? data.node : null;
        if (!resource) {
          return null;
        }
        return (
          <APIResourceScreenLayout
            layout="auto-rows"
            breadcrumbItems={[
              {
                to: "~/api-resources",
                label: <FormattedMessage id="ScreenNav.api-resources" />,
              },
              {
                to: "",
                label: resource.name ?? resource.resourceURI,
              },
            ]}
          >
            <APIResourceDetailsContent
              resource={resource}
              effectiveAppConfig={configQuery.effectiveAppConfig!}
              secretConfig={configQuery.secretConfig}
            />
          </APIResourceScreenLayout>
        );
      },
    });
  };

export default APIResourceDetailsScreen;
