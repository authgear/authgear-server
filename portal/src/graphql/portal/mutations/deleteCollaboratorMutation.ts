import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";

import { client } from "../apollo";
import {
  DeleteCollaboratorMutation,
  DeleteCollaboratorMutationVariables,
} from "./__generated__/DeleteCollaboratorMutation";

const deleteCollaboratorMutation = gql`
  mutation DeleteCollaboratorMutation($collaboratorID: String!) {
    deleteCollaborator(input: { collaboratorID: $collaboratorID }) {
      app {
        id
        collaborators {
          id
          createdAt
          userID
        }
      }
    }
  }
`;

export function useDeleteCollaboratorMutation(): {
  deleteCollaborator: (collaboratorID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    DeleteCollaboratorMutation,
    DeleteCollaboratorMutationVariables
  >(deleteCollaboratorMutation, {
    client,
  });

  const deleteCollaborator = useCallback(
    async (collaboratorID: string) => {
      const result = await mutationFunction({
        variables: {
          collaboratorID,
        },
      });

      return result.data != null;
    },
    [mutationFunction]
  );

  return { deleteCollaborator, error, loading };
}
