import {
  AuthenticatedForInvitationQuery,
  AuthenticatedForInvitationQueryVariables,
} from "./__generated__/AuthenticatedForInvitationQuery";
import { gql, QueryResult, useQuery } from "@apollo/client";
import { client } from "../apollo";

export const authenticatedQuery = gql`
  query AuthenticatedForInvitationQuery($code: String!) {
    viewer {
      email
    }
    checkCollaboratorInvitation(code: $code) {
      isCodeValid
      isInvitee
      appID
    }
  }
`;

export interface AuthenticatedForInvitationQueryResult
  extends Pick<
    QueryResult<AuthenticatedForInvitationQuery>,
    "loading" | "error" | "refetch"
  > {
  isCodeValid?: boolean;
  isAuthenticated?: boolean;
  isInvitee?: boolean;
  appID: string;
}

export const useAuthenticatedForInvitationQuery = (
  code: string
): AuthenticatedForInvitationQueryResult => {
  const { data, loading, error, refetch } = useQuery<
    AuthenticatedForInvitationQuery,
    AuthenticatedForInvitationQueryVariables
  >(authenticatedQuery, { client, variables: { code } });

  if (
    error?.networkError &&
    "statusCode" in error.networkError &&
    error.networkError.statusCode === 401
  ) {
    return { loading, error, refetch, appID: "" };
  }

  return {
    loading,
    error,
    refetch,
    isCodeValid: !!data?.checkCollaboratorInvitation.isCodeValid,
    isAuthenticated: !!data?.viewer,
    isInvitee: !!data?.checkCollaboratorInvitation.isInvitee,
    appID: data?.checkCollaboratorInvitation.appID ?? "",
  };
};
