import { QueryResult, WatchQueryFetchPolicy, useQuery } from "@apollo/client";
import { useMemo } from "react";
import {
  RoleQueryQuery,
  RoleQueryQueryVariables,
  RoleQueryDocument,
  RoleQueryNodeFragment,
} from "./roleQuery.generated";
import { NodeType } from "../node";

interface RoleQueryResult
  extends Pick<
    QueryResult<RoleQueryQuery, RoleQueryQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  role: RoleQueryNodeFragment | null;
}

export function useRoleQuery(
  roleID: string,
  options?: { skip?: boolean; fetchPolicy?: WatchQueryFetchPolicy }
): RoleQueryResult {
  const { data, loading, error, refetch } = useQuery<
    RoleQueryQuery,
    RoleQueryQueryVariables
  >(RoleQueryDocument, {
    variables: {
      roleID,
    },
    skip: options?.skip,
    fetchPolicy: options?.fetchPolicy,
  });

  const role = useMemo(() => {
    return data?.node?.__typename === NodeType.Role ? data.node : null;
  }, [data]);

  return { role, loading, error, refetch };
}
