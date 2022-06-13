import { QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";
import { nonNullable } from "../../../util/types";
import { client } from "../../portal/apollo";
import {
  AppListAppFragment,
  AppListQueryQuery,
  AppListQueryDocument,
} from "./appListQuery.generated";

export type App = AppListAppFragment;

interface AppListQueryResult
  extends Pick<
    QueryResult<AppListQueryQuery>,
    "loading" | "error" | "refetch"
  > {
  apps: App[] | null;
}

export const useAppListQuery = (): AppListQueryResult => {
  const { data, loading, error, refetch } = useQuery<AppListQueryQuery>(
    AppListQueryDocument,
    { client }
  );

  const apps = useMemo(() => {
    return (
      data?.apps?.edges?.map((edge) => edge?.node)?.filter(nonNullable) ?? null
    );
  }, [data]);

  return { apps, loading, error, refetch };
};
