import { gql, QueryResult, useQuery } from "@apollo/client";
import { client } from "../apollo";
import { AuthenticatedQuery } from "./__generated__/AuthenticatedQuery";

export const authenticatedQuery = gql`
  query AuthenticatedQuery {
    viewer {
      id
    }
  }
`;

export interface AuthenticatedQueryResult
  extends Pick<
    QueryResult<AuthenticatedQuery>,
    "loading" | "error" | "refetch"
  > {
  isAuthenticated: boolean;
}

export const useAuthenticatedQuery = (): AuthenticatedQueryResult => {
  const { data, loading, error, refetch } = useQuery<AuthenticatedQuery>(
    authenticatedQuery,
    { client }
  );

  if (
    error?.networkError &&
    "statusCode" in error.networkError &&
    error.networkError.statusCode === 401
  ) {
    return { isAuthenticated: false, loading, refetch };
  }

  const isAuthenticated = !!data?.viewer;
  return { isAuthenticated, loading, error, refetch };
};
