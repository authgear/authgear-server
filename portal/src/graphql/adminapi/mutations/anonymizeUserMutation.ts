import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  AnonymizeUserMutationMutation,
  AnonymizeUserMutationDocument,
} from "./anonymizeUserMutation.generated";

export function useAnonymizeUserMutation(): {
  anonymizeUser: (userID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<AnonymizeUserMutationMutation>(AnonymizeUserMutationDocument, {
      refetchQueries: ["userQuery"],
    });

  const anonymizeUser = useCallback(
    async (userID: string) => {
      const result = await mutationFunction({
        variables: {
          userID,
        },
      });

      return !!result.data;
    },
    [mutationFunction]
  );

  return { anonymizeUser, loading, error };
}
