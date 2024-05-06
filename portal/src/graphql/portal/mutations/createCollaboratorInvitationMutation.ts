import React from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../apollo";
import {
  CreateCollaboratorInvitationMutationMutation,
  CreateCollaboratorInvitationMutationMutationVariables,
  CreateCollaboratorInvitationMutationDocument,
} from "./createCollaboratorInvitationMutation.generated";
import { CollaboratorsAndInvitationsQueryDocument } from "../query/collaboratorsAndInvitationsQuery.generated";

export function useCreateCollaboratorInvitationMutation(appID: string): {
  createCollaboratorInvitation: (email: string) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] = useMutation<
    CreateCollaboratorInvitationMutationMutation,
    CreateCollaboratorInvitationMutationMutationVariables
  >(CreateCollaboratorInvitationMutationDocument, {
    client,
    refetchQueries: [
      {
        query: CollaboratorsAndInvitationsQueryDocument,
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
