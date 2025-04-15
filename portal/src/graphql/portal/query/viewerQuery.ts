import { useQuery, QueryResult, WatchQueryFetchPolicy } from "@apollo/client";
import { Viewer } from "../globalTypes.generated";
import { ViewerQueryQuery, ViewerQueryDocument } from "./viewerQuery.generated";
import { usePortalClient } from "../../portal/apollo";

export interface UseViewerQueryReturnType
  extends Pick<QueryResult<ViewerQueryQuery>, "loading" | "error" | "refetch"> {
  viewer?: Viewer | null;
}

export function useViewerQuery(options?: {
  fetchPolicy?: WatchQueryFetchPolicy;
}): UseViewerQueryReturnType {
  const client = usePortalClient();
  const { data, loading, error, refetch } = useQuery<ViewerQueryQuery>(
    ViewerQueryDocument,
    {
      client,
      fetchPolicy: options?.fetchPolicy ?? "network-only",
    }
  );

  return { viewer: data?.viewer, loading, error, refetch };
}
