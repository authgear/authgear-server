import { QueryResult, useQuery } from "@apollo/client";
import { usePortalClient } from "../apollo";
import {
  AuthenticatedForInvitationQueryQuery,
  AuthenticatedForInvitationQueryQueryVariables,
  AuthenticatedForInvitationQueryDocument,
} from "./authenticatedForInvitationQuery.generated";

export interface AuthenticatedForInvitationQueryResult
  extends Pick<
    QueryResult<AuthenticatedForInvitationQueryQuery>,
    "loading" | "error" | "refetch"
  > {
  isCodeValid?: boolean;
  isAuthenticated?: boolean;
  isInvitee?: boolean;
  appID?: string;
}

export const useAuthenticatedForInvitationQuery = (
  code: string
): AuthenticatedForInvitationQueryResult => {
  const client = usePortalClient();
  const { data, loading, error, refetch } = useQuery<
    AuthenticatedForInvitationQueryQuery,
    AuthenticatedForInvitationQueryQueryVariables
  >(AuthenticatedForInvitationQueryDocument, { client, variables: { code } });

  if (error?.networkError && "statusCode" in error.networkError) {
    return { loading, error, refetch };
  }

  return {
    loading,
    error,
    refetch,
    isCodeValid: !!data?.checkCollaboratorInvitation,
    isAuthenticated: !!data?.viewer,
    isInvitee: data?.checkCollaboratorInvitation?.isInvitee,
    appID: data?.checkCollaboratorInvitation?.appID,
  };
};
