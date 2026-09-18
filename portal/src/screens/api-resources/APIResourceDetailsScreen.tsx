import React, { useContext, useMemo, useState } from "react";
import cn from "classnames";
import { useLocation, useParams } from "react-router-dom";
import { CodeIcon } from "@radix-ui/react-icons";
import { useResourceQueryQuery } from "../../graphql/adminapi/query/resourceQuery.generated";
import { useLoadableView } from "../../hook/useLoadableView";
import { Context, FormattedMessage } from "../../intl";
import APIResourceScreenLayout from "../../components/api-resources/APIResourceScreenLayout";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import { Resource } from "../../graphql/adminapi/globalTypes.generated";
import { APIResourceDetailsScreenDetailsSection } from "./APIResourceDetailsScreenDetailsSection";
import { APIResourceDetailsScreenScopesSection } from "./APIResourceDetailsScreenScopesSection";
import { APIResourceDetailsScreenAccessPolicySection } from "./APIResourceDetailsScreenAccessPolicySection";
import { APIResourceDetailsScreenTestSection } from "./APIResourceDetailsScreenTestSection";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { usePivotNavigation } from "../../hook/usePivot";
import { OverflowTabs } from "../../components/v2/OverflowTabs/OverflowTabs";
import { useAppSecretVisitToken } from "../../graphql/portal/mutations/generateAppSecretVisitTokenMutation";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { AppSecretKey } from "../../graphql/portal/globalTypes.generated";
import { PortalAPIAppConfig, PortalAPISecretConfig } from "../../types";
import styles from "./APIResourceDetailsScreen.module.css";

export interface LocationState {
  isClientSecretRevealed: boolean;
}

const SECRETS = [AppSecretKey.OauthClientSecrets];

type TabKey = "settings" | "scopes";

// Module-level so its identity is stable: usePivotNavigation keeps it in an
// effect's dependency list.
const TAB_KEYS: TabKey[] = ["settings", "scopes"];

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
  selectedKey,
}: {
  resource: Resource;
  effectiveAppConfig: PortalAPIAppConfig;
  secretConfig: PortalAPISecretConfig | null;
  selectedKey: TabKey;
}) {
  // Scopes gets the full width: its table and the add-scope form use it.
  const wide = selectedKey === "scopes";
  return (
    <div className={cn(styles.content, wide && styles["content--wide"])}>
      {wide ? (
        <APIResourceDetailsScreenScopesSection resource={resource} />
      ) : (
        <>
          <APIResourceDetailsScreenDetailsSection resource={resource} />
          <APIResourceDetailsScreenAccessPolicySection
            resource={resource}
            effectiveAppConfig={effectiveAppConfig}
          />
          <APIResourceDetailsScreenTestSection
            resource={resource}
            effectiveAppConfig={effectiveAppConfig}
            secretConfig={secretConfig}
          />
        </>
      )}
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
    const { selectedKey, onChangeKey } = usePivotNavigation<TabKey>(TAB_KEYS);
    const { renderToString } = useContext(Context);
    const tabs = useMemo(
      () => [
        {
          value: "settings",
          label: renderToString("APIResourceDetailsScreen.tab.settings"),
        },
        {
          value: "scopes",
          label: renderToString("APIResourceDetailsScreen.tab.scopes"),
        },
      ],
      [renderToString]
    );
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
            layout={selectedKey === "scopes" ? "list" : "auto-rows"}
            headerTabs={
              <OverflowTabs
                value={selectedKey}
                onValueChange={onChangeKey as (value: string) => void}
                tabs={tabs}
              />
            }
            headerIcon={<CodeIcon width="2rem" height="2rem" />}
            headerDescription={
              <div className={styles.headerMeta}>
                <span className={styles.headerMetaLabel}>
                  <FormattedMessage id="ResourceForm.resourceURI.label" />
                </span>
                <code className={styles.headerIdentifier}>
                  {resource.resourceURI}
                </code>
                <CopyIconButton textToCopy={resource.resourceURI} />
              </div>
            }
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
              selectedKey={selectedKey}
            />
          </APIResourceScreenLayout>
        );
      },
    });
  };

export default APIResourceDetailsScreen;
