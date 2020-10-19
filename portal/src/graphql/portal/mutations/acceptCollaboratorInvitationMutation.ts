import React from "react";
import { gql, useMutation } from "@apollo/client";

import { client } from "../../portal/apollo";
import {
  AcceptCollaboratorInvitationMutation,
  AcceptCollaboratorInvitationMutationVariables,
} from "./__generated__/AcceptCollaboratorInvitationMutation";

const acceptCollaboratorInvitationMutation = gql`
  mutation AcceptCollaboratorInvitationMutation($code: String!) {
    acceptCollaboratorInvitation(input: { code: $code }) {
      app {
        id
        collaborators {
          id
          createdAt
          user {
            id
            email
          }
        }
        collaboratorInvitations {
          id
          createdAt
          expireAt
          invitedBy {
            id
            email
          }
          inviteeEmail
        }
      }
    }
  }
`;

export function useAcceptCollaboratorInvitationMutation(): {
  acceptCollaboratorInvitation: (code: string) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    AcceptCollaboratorInvitationMutation,
    AcceptCollaboratorInvitationMutationVariables
  >(acceptCollaboratorInvitationMutation, {
    client,
  });
  const acceptCollaboratorInvitation = React.useCallback(
    async (code: string) => {
      const result = await mutationFunction({
        variables: {
          code,
        },
      });
      return result.data?.acceptCollaboratorInvitation.app.id ?? null;
    },
    [mutationFunction]
  );
  return { acceptCollaboratorInvitation, error, loading };
}
