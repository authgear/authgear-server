import { QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";
import {
  UserQueryQuery,
  UserQueryQueryVariables,
  UserQueryDocument,
  UserQueryNodeFragment,
} from "./userQuery.generated";

interface UserQueryResult
  extends Pick<
    QueryResult<UserQueryQuery, UserQueryQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  user: UserQueryNodeFragment | null;
}

export function useUserQuery(userID: string): UserQueryResult {
  const { data, loading, error, refetch } = useQuery<
    UserQueryQuery,
    UserQueryQueryVariables
  >(UserQueryDocument, {
    variables: {
      userID,
    },
  });

  const user = useMemo(() => {
    return data?.node?.__typename === "User" ? data.node : null;
  }, [data]);

  return { user, loading, error, refetch };
}
