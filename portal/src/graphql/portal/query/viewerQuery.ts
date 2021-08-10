import { gql, useQuery, QueryResult } from "@apollo/client";
import { client } from "../../portal/apollo";
import { ViewerQuery, ViewerQuery_viewer } from "./__generated__/ViewerQuery";

const viewerQuery = gql`
  query ViewerQuery {
    viewer {
      id
      email
    }
  }
`;

export type Viewer = ViewerQuery_viewer;

export interface UseViewerQueryReturnType
  extends Pick<QueryResult<ViewerQuery>, "loading" | "error" | "refetch"> {
  viewer?: Viewer | null;
}

export function useViewerQuery(): UseViewerQueryReturnType {
  const { data, loading, error, refetch } = useQuery<ViewerQuery>(viewerQuery, {
    client,
  });

  return {
    viewer: data?.viewer,
    loading,
    error,
    refetch,
  };
}
