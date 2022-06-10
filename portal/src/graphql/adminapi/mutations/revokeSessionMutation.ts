import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  RevokeSessionMutationMutation,
  RevokeSessionMutationDocument,
} from "./revokeSessionMutation.generated";

export function useRevokeSessionMutation(): {
  revokeSession: (sessionID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<RevokeSessionMutationMutation>(RevokeSessionMutationDocument);

  const revokeSession = useCallback(
    async (sessionID: string) => {
      const result = await mutationFunction({
        variables: {
          sessionID,
        },
      });
      return !!result.data?.revokeSession;
    },
    [mutationFunction]
  );
  return { revokeSession, error, loading };
}
