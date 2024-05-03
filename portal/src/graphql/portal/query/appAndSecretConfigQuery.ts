import { useMemo } from "react";
import { QueryResult, useQuery } from "@apollo/client";
import { usePortalClient } from "../../portal/apollo";
import {
  AppAndSecretConfigQueryQuery,
  AppAndSecretConfigQueryQueryVariables,
  AppAndSecretConfigQueryDocument,
} from "./appAndSecretConfigQuery.generated";
import { PortalAPIAppConfig, PortalAPISecretConfig } from "../../../types";
import { Collaborator } from "../globalTypes.generated";

export interface AppAndSecretConfigQueryResult
  extends Pick<
    QueryResult<
      AppAndSecretConfigQueryQuery,
      AppAndSecretConfigQueryQueryVariables
    >,
    "loading" | "error" | "refetch"
  > {
  rawAppConfig: PortalAPIAppConfig | null;
  rawAppConfigChecksum?: string;
  effectiveAppConfig: PortalAPIAppConfig | null;
  secretConfig: PortalAPISecretConfig | null;
  secretConfigChecksum?: string;
  viewer: Collaborator | null;
}
export const useAppAndSecretConfigQuery = (
  appID: string,
  token: string | null = null
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
    };
  }, [data]);

  return { ...queryData, loading, error, refetch };
};
