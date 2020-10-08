import { useMemo } from "react";
import { gql, QueryResult, useQuery } from "@apollo/client";
import { client } from "../../portal/apollo";
import {
  AppAndSecretConfigQuery,
  AppAndSecretConfigQueryVariables,
  AppAndSecretConfigQuery_node_App,
} from "./__generated__/AppAndSecretConfigQuery";

export const appAndSecretConfigQuery = gql`
  query AppAndSecretConfigQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        effectiveAppConfig
        rawAppConfig
        rawSecretConfig
      }
    }
  }
`;

interface AppAndSecretConfigQueryResult
  extends Pick<
    QueryResult<AppAndSecretConfigQuery, AppAndSecretConfigQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  rawAppConfig: AppAndSecretConfigQuery_node_App["rawAppConfig"] | null;
  effectiveAppConfig:
    | AppAndSecretConfigQuery_node_App["effectiveAppConfig"]
    | null;
  secretConfig: AppAndSecretConfigQuery_node_App["rawSecretConfig"] | null;
}
export const useAppAndSecretConfigQuery = (
  appID: string
): AppAndSecretConfigQueryResult => {
  const { data, loading, error, refetch } = useQuery<AppAndSecretConfigQuery>(
    appAndSecretConfigQuery,
    {
      client,
      variables: {
        id: appID,
      },
    }
  );

  const queryData = useMemo(() => {
    const appConfigNode = data?.node?.__typename === "App" ? data.node : null;
    return {
      rawAppConfig: appConfigNode?.rawAppConfig ?? null,
      effectiveAppConfig: appConfigNode?.effectiveAppConfig ?? null,
      secretConfig: appConfigNode?.rawSecretConfig ?? null,
    };
  }, [data]);

  return { ...queryData, loading, error, refetch };
};
