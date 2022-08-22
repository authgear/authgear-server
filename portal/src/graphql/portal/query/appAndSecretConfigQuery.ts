import { useMemo } from "react";
import { QueryResult, useQuery } from "@apollo/client";
import { client } from "../../portal/apollo";
import {
  AppAndSecretConfigQueryQuery,
  AppAndSecretConfigQueryQueryVariables,
  AppAndSecretConfigQueryDocument,
} from "./appAndSecretConfigQuery.generated";
import { PortalAPIAppConfig, PortalAPISecretConfig } from "../../../types";
import { CollaboratorRole } from "../globalTypes.generated";

export interface AppAndSecretConfigQueryResult
  extends Pick<
    QueryResult<
      AppAndSecretConfigQueryQuery,
      AppAndSecretConfigQueryQueryVariables
    >,
    "loading" | "error" | "refetch"
  > {
  rawAppConfig: PortalAPIAppConfig | null;
  effectiveAppConfig: PortalAPIAppConfig | null;
  secretConfig: PortalAPISecretConfig | null;
  collaboratorRole: CollaboratorRole | null;
}
export const useAppAndSecretConfigQuery = (
  appID: string
): AppAndSecretConfigQueryResult => {
  const { data, loading, error, refetch } =
    useQuery<AppAndSecretConfigQueryQuery>(AppAndSecretConfigQueryDocument, {
      client,
      variables: {
        id: appID,
      },
    });

  const queryData = useMemo(() => {
    const appConfigNode = data?.node?.__typename === "App" ? data.node : null;
    return {
      rawAppConfig: appConfigNode?.rawAppConfig ?? null,
      effectiveAppConfig: appConfigNode?.effectiveAppConfig ?? null,
      secretConfig: appConfigNode?.secretConfig ?? null,
      collaboratorRole: appConfigNode?.currentCollaboratorRole ?? null,
    };
  }, [data]);

  return { ...queryData, loading, error, refetch };
};
