import { useCallback } from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../apollo";
import {
  SetSubscriptionCancelledStatusMutationMutation,
  SetSubscriptionCancelledStatusMutationDocument,
} from "./setSubscriptionCancelledStatusMutation.generated";

export function useSetSubscriptionCancelledStatusMutation(appID: string): {
  setSubscriptionCancelledStatus: (cancelled: boolean) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] =
    useMutation<SetSubscriptionCancelledStatusMutationMutation>(
      SetSubscriptionCancelledStatusMutationDocument,
      {
        client,
      }
    );

  const setSubscriptionCancelledStatus = useCallback(
    async (cancelled: boolean) => {
      await mutationFunction({
        variables: { appID, cancelled },
      });
    },
    [mutationFunction, appID]
  );

  return { setSubscriptionCancelledStatus, error, loading };
}
