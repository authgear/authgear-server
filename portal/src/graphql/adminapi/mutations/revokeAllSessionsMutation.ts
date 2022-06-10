import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  RevokeAllSessionsMutationMutation,
  RevokeAllSessionsMutationDocument,
} from "./revokeAllSessionsMutation.generated";

export function useRevokeAllSessionsMutation(): {
  revokeAllSessions: (userID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<RevokeAllSessionsMutationMutation>(
      RevokeAllSessionsMutationDocument
    );

  const revokeAllSessions = useCallback(
    async (userID: string) => {
      const result = await mutationFunction({
        variables: {
          userID,
        },
      });
      return !!result.data?.revokeAllSessions;
    },
    [mutationFunction]
  );
  return { revokeAllSessions, error, loading };
}
