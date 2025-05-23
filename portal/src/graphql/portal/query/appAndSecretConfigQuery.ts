import { useCallback, useMemo } from "react";
import { QueryResult, useQuery } from "@apollo/client";
import { usePortalClient } from "../../portal/apollo";
import {
  AppAndSecretConfigQueryQuery,
  AppAndSecretConfigQueryQueryVariables,
  AppAndSecretConfigQueryDocument,
} from "./appAndSecretConfigQuery.generated";
import { PortalAPIAppConfig, PortalAPISecretConfig } from "../../../types";
import { Collaborator, EffectiveSecretConfig } from "../globalTypes.generated";
import { Loadable } from "../../../hook/useLoadableView";

export interface AppAndSecretConfigQueryResult
  extends Pick<
      QueryResult<
        AppAndSecretConfigQueryQuery,
        AppAndSecretConfigQueryQueryVariables
      >,
      "loading" | "error" | "refetch"
    >,
    Loadable {
  rawAppConfig: PortalAPIAppConfig | null;
  rawAppConfigChecksum?: string;
  effectiveAppConfig: PortalAPIAppConfig | null;
  secretConfig: PortalAPISecretConfig | null;
  secretConfigChecksum?: string;
  viewer: Collaborator | null;
  samlIdpEntityID?: string;
  effectiveSecretConfig?: EffectiveSecretConfig;
}
export const useAppAndSecretConfigQuery = (
  appID: string,
  token: string | null = null,
  skip: boolean = false
): AppAndSecretConfigQueryResult => {
  const client = usePortalClient();
  const { data, loading, error, refetch } = useQuery<
    AppAndSecretConfigQueryQuery,
    AppAndSecretConfigQueryQueryVariables
  >(AppAndSecretConfigQueryDocument, {
    client,
    variables: {
      id: appID,
      token: token,
    },
    skip: skip,
  });

  const queryData = useMemo(() => {
    const appConfigNode = data?.node?.__typename === "App" ? data.node : null;
    return {
      rawAppConfig: appConfigNode?.rawAppConfig ?? null,
      rawAppConfigChecksum: appConfigNode?.rawAppConfigChecksum ?? undefined,
      effectiveAppConfig: appConfigNode?.effectiveAppConfig ?? null,
      secretConfig: appConfigNode?.secretConfig ?? null,
      secretConfigChecksum: appConfigNode?.secretConfigChecksum ?? undefined,
      viewer: appConfigNode?.viewer ?? null,
      samlIdpEntityID: appConfigNode?.samlIdpEntityID ?? undefined,
      effectiveSecretConfig: appConfigNode?.effectiveSecretConfig ?? undefined,
    };
  }, [data]);

  return {
    ...queryData,
    loading,
    error,
    refetch,
    isLoading: loading,
    loadError: error,
    reload: useCallback(() => {
      refetch();
    }, [refetch]),
  };
};
