import { gql, QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";
import { nonNullable } from "../../../util/types";
import { client } from "../../portal/apollo";
import {
  AppListQuery,
  AppListQuery_apps_edges_node,
} from "./__generated__/AppListQuery";

export const appListQuery = gql`
  query AppListQuery {
    apps {
      edges {
        node {
          id
          effectiveAppConfig
        }
      }
    }
  }
`;

export type App = AppListQuery_apps_edges_node;
interface AppListQueryResult
  extends Pick<QueryResult<AppListQuery>, "loading" | "error" | "refetch"> {
  apps: App[] | null;
}

export const useAppListQuery = (): AppListQueryResult => {
  const { data, loading, error, refetch } = useQuery<AppListQuery>(
    appListQuery,
    { client }
  );

  const apps = useMemo(() => {
    return (
      data?.apps?.edges?.map((edge) => edge?.node)?.filter(nonNullable) ?? null
    );
  }, [data]);

  return { apps, loading, error, refetch };
};
