import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";

import { client } from "../apollo";
import {
  DeleteCollaboratorInvitationMutation,
  DeleteCollaboratorInvitationMutationVariables,
} from "./__generated__/DeleteCollaboratorInvitationMutation";

const deleteCollaboratorInvitationMutation = gql`
  mutation DeleteCollaboratorInvitationMutation(
    $collaboratorInvitationID: String!
  ) {
    deleteCollaboratorInvitation(
      input: { collaboratorInvitationID: $collaboratorInvitationID }
    ) {
      app {
        id
        collaboratorInvitations {
          id
          createdAt
          expireAt
          invitedBy
          inviteeEmail
        }
      }
    }
  }
`;

export function useDeleteCollaboratorInvitationMutation(): {
  deleteCollaboratorInvitation: (
    collaboratorInvitationID: string
  ) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    DeleteCollaboratorInvitationMutation,
    DeleteCollaboratorInvitationMutationVariables
  >(deleteCollaboratorInvitationMutation, {
    client,
  });

  const deleteCollaboratorInvitation = useCallback(
    async (collaboratorInvitationID: string) => {
      const result = await mutationFunction({
        variables: {
          collaboratorInvitationID,
        },
      });

      return result.data != null;
    },
    [mutationFunction]
  );

  return { deleteCollaboratorInvitation, error, loading };
}
