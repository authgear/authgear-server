import { useCallback } from "react";
import { useMutation } from "@apollo/client";

import { client } from "../apollo";
import {
  DeleteCollaboratorMutationMutation,
  DeleteCollaboratorMutationMutationVariables,
  DeleteCollaboratorMutationDocument,
} from "./deleteCollaboratorMutation.generated";

export function useDeleteCollaboratorMutation(): {
  deleteCollaborator: (collaboratorID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    DeleteCollaboratorMutationMutation,
    DeleteCollaboratorMutationMutationVariables
  >(DeleteCollaboratorMutationDocument, {
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
