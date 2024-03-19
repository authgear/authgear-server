import { QueryResult, WatchQueryFetchPolicy, useQuery } from "@apollo/client";
import { useMemo } from "react";
import {
  GroupQueryQuery,
  GroupQueryQueryVariables,
  GroupQueryDocument,
  GroupQueryNodeFragment,
} from "./groupQuery.generated";
import { NodeType } from "../node";

interface GroupQueryResult
  extends Pick<
    QueryResult<GroupQueryQuery, GroupQueryQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  group: GroupQueryNodeFragment | null;
}

export function useGroupQuery(
  groupID: string,
  options?: { skip?: boolean; fetchPolicy?: WatchQueryFetchPolicy }
): GroupQueryResult {
  const { data, loading, error, refetch } = useQuery<
    GroupQueryQuery,
    GroupQueryQueryVariables
  >(GroupQueryDocument, {
    variables: {
      groupID,
    },
    skip: options?.skip,
    fetchPolicy: options?.fetchPolicy,
  });

  const group = useMemo(() => {
    return data?.node?.__typename === NodeType.Group ? data.node : null;
  }, [data]);

  return { group, loading, error, refetch };
}
