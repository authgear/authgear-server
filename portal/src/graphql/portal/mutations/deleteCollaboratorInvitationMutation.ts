import { useCallback } from "react";
import { useMutation } from "@apollo/client";

import { client } from "../apollo";
import {
  DeleteCollaboratorInvitationMutationMutation,
  DeleteCollaboratorInvitationMutationMutationVariables,
  DeleteCollaboratorInvitationMutationDocument,
} from "./deleteCollaboratorInvitationMutation.generated";

export function useDeleteCollaboratorInvitationMutation(): {
  deleteCollaboratorInvitation: (
    collaboratorInvitationID: string
  ) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    DeleteCollaboratorInvitationMutationMutation,
    DeleteCollaboratorInvitationMutationMutationVariables
  >(DeleteCollaboratorInvitationMutationDocument, {
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
