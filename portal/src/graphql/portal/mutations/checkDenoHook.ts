import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import { usePortalClient } from "../../portal/apollo";
import {
  CheckDenoHookMutationDocument,
  CheckDenoHookMutationMutation,
} from "./checkDenoHook.generated";

export interface UseCheckDenoHookMutationReturnType {
  checkDenoHook: (content: string) => Promise<void>;
  loading: boolean;
  error: unknown;
  reset: () => void;
}

export function useCheckDenoHookMutation(
  appID: string
): UseCheckDenoHookMutationReturnType {
  const client = usePortalClient();
  const [mutationFunction, { error, loading, reset }] =
    useMutation<CheckDenoHookMutationMutation>(CheckDenoHookMutationDocument, {
      client,
    });
  const checkDenoHook = useCallback(
    async (content: string) => {
      await mutationFunction({
        variables: {
          appID,
          content,
        },
      });
    },
    [mutationFunction, appID]
  );

  return { checkDenoHook, error, loading, reset };
}
