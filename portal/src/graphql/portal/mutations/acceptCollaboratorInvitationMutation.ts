import React from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../../portal/apollo";
import {
  AcceptCollaboratorInvitationMutationMutation,
  AcceptCollaboratorInvitationMutationMutationVariables,
  AcceptCollaboratorInvitationMutationDocument,
} from "./acceptCollaboratorInvitationMutation.generated";

export function useAcceptCollaboratorInvitationMutation(): {
  acceptCollaboratorInvitation: (code: string) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] = useMutation<
    AcceptCollaboratorInvitationMutationMutation,
    AcceptCollaboratorInvitationMutationMutationVariables
  >(AcceptCollaboratorInvitationMutationDocument, {
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
