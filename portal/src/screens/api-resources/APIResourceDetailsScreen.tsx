import React, { useContext, useState } from "react";
import { useLocation, useParams } from "react-router-dom";
import { useResourceQueryQuery } from "../../graphql/adminapi/query/resourceQuery.generated";
import { useLoadableView } from "../../hook/useLoadableView";
import { FormattedMessage, Context as MessageContext } from "../../intl";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { PivotItem } from "@fluentui/react";
import { AGPivot } from "../../components/common/AGPivot";
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

function APIResourceDetailsContent({
  resource,
  effectiveAppConfig,
  secretConfig,
}: {
  resource: Resource;
  effectiveAppConfig: PortalAPIAppConfig;
  secretConfig: PortalAPISecretConfig | null;
}) {
  const { selectedKey, onLinkClick } = usePivotNavigation([
    "details",
    "scopes",
    "applications",
    "test",
  ]);
  const { renderToString } = useContext(MessageContext);
  return (
    <div className="pt-6 flex flex-col col-span-full">
      <AGPivot selectedKey={selectedKey} onLinkClick={onLinkClick}>
        <PivotItem
          headerText={renderToString("APIResourceDetailsScreen.tab.details")}
          itemKey="details"
        />
        <PivotItem
          headerText={renderToString("APIResourceDetailsScreen.tab.scopes")}
          itemKey="scopes"
        />
        <PivotItem
          headerText={renderToString(
            "APIResourceDetailsScreen.tab.applications"
          )}
          itemKey="applications"
        />
        <PivotItem
          headerText={renderToString("APIResourceDetailsScreen.tab.test")}
          itemKey="test"
        />
      </AGPivot>
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
