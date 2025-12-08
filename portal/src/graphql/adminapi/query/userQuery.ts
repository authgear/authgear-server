import { QueryResult, useQuery, WatchQueryFetchPolicy } from "@apollo/client";
import { useMemo } from "react";
import {
  UserQueryQuery,
  UserQueryQueryVariables,
  UserQueryDocument,
  UserQueryNodeFragment,
} from "./userQuery.generated";
import { NodeType } from "../node";

interface UserQueryResult
  extends Pick<
    QueryResult<UserQueryQuery, UserQueryQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  user: UserQueryNodeFragment | null;
}

export function useUserQuery(
  userID: string,
  options?: { skip?: boolean; fetchPolicy?: WatchQueryFetchPolicy }
): UserQueryResult {
  const { data, loading, error, refetch } = useQuery<
    UserQueryQuery,
    UserQueryQueryVariables
  >(UserQueryDocument, {
    variables: {
      userID,
    },
    // cache-first is the default
    fetchPolicy: options?.fetchPolicy ?? "cache-first",
    skip: options?.skip,
  });

  const user = useMemo(() => {
    return data?.node?.__typename === NodeType.User ? data.node : null;
  }, [data]);

  return { user, loading, error, refetch };
}
