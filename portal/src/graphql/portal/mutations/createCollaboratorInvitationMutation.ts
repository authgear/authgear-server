import React from "react";
import { gql, useMutation } from "@apollo/client";

import { client } from "../apollo";
import {
  CreateCollaboratorInvitationMutation,
  CreateCollaboratorInvitationMutationVariables,
} from "./__generated__/CreateCollaboratorInvitationMutation";
import { collaboratorsAndInvitationsQuery } from "../query/collaboratorsAndInvitationsQuery";

const createCollaboratorInvitationMutation = gql`
  mutation CreateCollaboratorInvitationMutation($appID: ID!, $email: String!) {
    createCollaboratorInvitation(
      input: { appID: $appID, inviteeEmail: $email }
    ) {
      collaboratorInvitation {
        id
        createdAt
        expireAt
        invitedBy
        inviteeEmail
      }
    }
  }
`;

export function useCreateCollaboratorInvitationMutation(
  appID: string
): {
  createCollaboratorInvitation: (email: string) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    CreateCollaboratorInvitationMutation,
    CreateCollaboratorInvitationMutationVariables
  >(createCollaboratorInvitationMutation, {
    client,
    refetchQueries: [
      {
        query: collaboratorsAndInvitationsQuery,
        variables: {
          appID,
        },
      },
    ],
  });
  const createCollaboratorInvitation = React.useCallback(
    async (email: string) => {
      const result = await mutationFunction({
        variables: {
          appID,
          email,
        },
      });
      return (
        result.data?.createCollaboratorInvitation.collaboratorInvitation.id ??
        null
      );
    },
    [appID, mutationFunction]
  );
  return { createCollaboratorInvitation, error, loading };
}
