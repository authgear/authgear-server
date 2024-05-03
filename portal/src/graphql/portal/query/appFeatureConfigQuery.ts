import { useMemo } from "react";
import { QueryResult, useQuery } from "@apollo/client";
import { usePortalClient } from "../../portal/apollo";
import {
  AppFeatureConfigQueryQuery,
  AppFeatureConfigQueryQueryVariables,
  AppFeatureConfigFragment,
  AppFeatureConfigQueryDocument,
} from "./appFeatureConfigQuery.generated";

interface AppFeatureConfigQueryResult
  extends Pick<
    QueryResult<
      AppFeatureConfigQueryQuery,
      AppFeatureConfigQueryQueryVariables
    >,
    "loading" | "error" | "refetch"
  > {
  effectiveFeatureConfig: AppFeatureConfigFragment["effectiveFeatureConfig"];
  planName: AppFeatureConfigFragment["planName"] | null;
}

export const useAppFeatureConfigQuery = (
  appID: string
): AppFeatureConfigQueryResult => {
  const client = usePortalClient();
  const { data, loading, error, refetch } =
    useQuery<AppFeatureConfigQueryQuery>(AppFeatureConfigQueryDocument, {
      client,
      variables: {
        id: appID,
      },
    });

  const queryData = useMemo(() => {
    const featureConfigNode =
      data?.node?.__typename === "App" ? data.node : null;
    return {
      effectiveFeatureConfig: featureConfigNode?.effectiveFeatureConfig ?? null,
      planName: featureConfigNode?.planName ?? null,
    };
  }, [data]);

  return { ...queryData, loading, error, refetch };
};
