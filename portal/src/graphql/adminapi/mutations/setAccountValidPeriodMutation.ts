import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import {
  SetAccountValidPeriodMutationMutation,
  SetAccountValidPeriodMutationDocument,
} from "./setAccountValidPeriodMutation.generated";

export function useSetAccountValidPeriodMutation(): {
  setAccountValidPeriod: (opts: {
    userID: string;
    accountValidFrom: Date | null;
    accountValidUntil: Date | null;
  }) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { loading, error }] =
    useMutation<SetAccountValidPeriodMutationMutation>(
      SetAccountValidPeriodMutationDocument,
      {
        // Setting account valid period could terminate all sessions.
        // So we have to refetch queries that fetch sessions.
        refetchQueries: ["UserQuery"],
      }
    );

  const setAccountValidPeriod = useCallback(
    async (opts: {
      userID: string;
      accountValidFrom: Date | null;
      accountValidUntil: Date | null;
    }) => {
      const result = await mutationFunction({
        variables: {
          userID: opts.userID,
          accountValidFrom: opts.accountValidFrom,
          accountValidUntil: opts.accountValidUntil,
        },
      });

      return !!result.data;
    },
    [mutationFunction]
  );

  return { setAccountValidPeriod, loading, error };
}
