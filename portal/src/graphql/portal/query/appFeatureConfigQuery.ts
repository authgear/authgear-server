import { useMemo } from "react";
import { gql, QueryResult, useQuery } from "@apollo/client";
import { client } from "../../portal/apollo";
import {
  AppFeatureConfigQuery,
  AppFeatureConfigQueryVariables,
  AppFeatureConfigQuery_node_App,
} from "./__generated__/AppFeatureConfigQuery";

export const appFeatureConfigQuery = gql`
  query AppFeatureConfigQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        effectiveFeatureConfig
      }
    }
  }
`;

interface AppFeatureConfigQueryResult
  extends Pick<
    QueryResult<AppFeatureConfigQuery, AppFeatureConfigQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  effectiveFeatureConfig:
    | AppFeatureConfigQuery_node_App["effectiveFeatureConfig"]
    | null;
}

export const useAppFeatureConfigQuery = (
  appID: string
): AppFeatureConfigQueryResult => {
  const { data, loading, error, refetch } = useQuery<AppFeatureConfigQuery>(
    appFeatureConfigQuery,
    {
      client,
      variables: {
        id: appID,
      },
    }
  );

  const queryData = useMemo(() => {
    const featureConfigNode =
      data?.node?.__typename === "App" ? data.node : null;
    return {
      effectiveFeatureConfig: featureConfigNode?.effectiveFeatureConfig ?? null,
    };
  }, [data]);

  return { ...queryData, loading, error, refetch };
};
