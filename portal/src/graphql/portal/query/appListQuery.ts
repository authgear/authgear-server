import { QueryResult, useQuery } from "@apollo/client";
import { usePortalClient } from "../../portal/apollo";
import {
  AppListQueryQuery,
  AppListQueryDocument,
} from "./appListQuery.generated";
import { AppListItem } from "../globalTypes.generated";

interface AppListQueryResult
  extends Pick<
    QueryResult<AppListQueryQuery>,
    "loading" | "error" | "refetch"
  > {
  apps: AppListItem[] | null;
}

export const useAppListQuery = (): AppListQueryResult => {
  const client = usePortalClient();
  const { data, loading, error, refetch } = useQuery<AppListQueryQuery>(
    AppListQueryDocument,
    { client }
  );

  return { apps: data?.appList ?? null, loading, error, refetch };
};
