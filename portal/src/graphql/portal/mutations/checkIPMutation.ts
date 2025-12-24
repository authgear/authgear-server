import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import { usePortalClient } from "../../portal/apollo";
import {
  CheckIpMutationMutation,
  CheckIpMutationMutationVariables,
  CheckIpMutationDocument,
} from "./checkIPMutation.generated";

export interface UseCheckIPMutationReturnType {
  checkIP: (
    ipAddress: string,
    cidrs: string[],
    countryCodes: string[]
  ) => Promise<boolean | null | undefined>;
  loading: boolean;
  error: unknown;
  reset: () => void;
}

export function useCheckIPMutation(
  appID: string
): UseCheckIPMutationReturnType {
  const client = usePortalClient();
  const [mutationFunction, { error, loading, reset }] = useMutation<
    CheckIpMutationMutation,
    CheckIpMutationMutationVariables
  >(CheckIpMutationDocument, {
    client,
  });
  const checkIP = useCallback(
    async (ipAddress: string, cidrs: string[], countryCodes: string[]) => {
      const result = await mutationFunction({
        variables: {
          appID,
          ipAddress,
          cidrs,
          countryCodes,
        },
      });
      return result.data?.checkIP;
    },
    [mutationFunction, appID]
  );

  return { checkIP, error, loading, reset };
}
