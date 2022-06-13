import { useQuery, QueryResult } from "@apollo/client";
import { User } from "../globalTypes.generated";
import { ViewerQueryQuery, ViewerQueryDocument } from "./viewerQuery.generated";
import { client } from "../../portal/apollo";

export type Viewer = User;

export interface UseViewerQueryReturnType
  extends Pick<QueryResult<ViewerQueryQuery>, "loading" | "error" | "refetch"> {
  viewer?: Viewer | null;
}

export function useViewerQuery(): UseViewerQueryReturnType {
  const { data, loading, error, refetch } = useQuery<ViewerQueryQuery>(
    ViewerQueryDocument,
    {
      client,
    }
  );

  return { viewer: data?.viewer, loading, error, refetch };
}
