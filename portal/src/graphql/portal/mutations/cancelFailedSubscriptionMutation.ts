import { useCallback } from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../apollo";
import {
  CancelFailedSubscriptionMutationMutation,
  CancelFailedSubscriptionMutationDocument,
} from "./cancelFailedSubscriptionMutation.generated";

export function useCancelFailedSubscriptionMutation(appID: string): {
  cancelFailedSubscription: () => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] =
    useMutation<CancelFailedSubscriptionMutationMutation>(
      CancelFailedSubscriptionMutationDocument,
      {
        client,
      }
    );

  const cancelFailedSubscription = useCallback(async () => {
    await mutationFunction({
      variables: { appID },
    });
  }, [mutationFunction, appID]);

  return { cancelFailedSubscription, error, loading };
}
