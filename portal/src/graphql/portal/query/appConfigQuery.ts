import { gql, QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";
import { client } from "../../portal/apollo";
import {
  AppConfigQuery,
  AppConfigQueryVariables,
  AppConfigQuery_node_App,
} from "./__generated__/AppConfigQuery";

export const appConfigQuery = gql`
  query AppConfigQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        effectiveAppConfig
        rawAppConfig
      }
    }
  }
`;

interface AppConfigQueryResult
  extends Pick<
    QueryResult<AppConfigQuery, AppConfigQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  rawAppConfig: AppConfigQuery_node_App["rawAppConfig"] | null;
  effectiveAppConfig: AppConfigQuery_node_App["effectiveAppConfig"] | null;
}

export const useAppConfigQuery = (appID: string): AppConfigQueryResult => {
  const { data, loading, error, refetch } = useQuery<AppConfigQuery>(
    appConfigQuery,
    {
      client,
      variables: {
        id: appID,
      },
    }
  );

  const { rawAppConfig, effectiveAppConfig } = useMemo(() => {
    const appConfigNode = data?.node?.__typename === "App" ? data.node : null;
    return {
      rawAppConfig: appConfigNode?.rawAppConfig ?? null,
      effectiveAppConfig: appConfigNode?.effectiveAppConfig ?? null,
    };
  }, [data]);

  return { rawAppConfig, effectiveAppConfig, loading, error, refetch };
};
