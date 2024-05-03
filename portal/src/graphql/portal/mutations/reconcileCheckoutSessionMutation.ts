import React from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../../portal/apollo";
import {
  ReconcileCheckoutSessionMutationMutation,
  ReconcileCheckoutSessionMutationDocument,
} from "./reconcileCheckoutSessionMutation.generated";
import { AppListQueryDocument } from "../query/appListQuery.generated";

export function useReconcileCheckoutSessionMutation(): {
  reconcileCheckoutSession: (
    appID: string,
    checkoutSessionID: string
  ) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] =
    useMutation<ReconcileCheckoutSessionMutationMutation>(
      ReconcileCheckoutSessionMutationDocument,
      {
        client,
        refetchQueries: [{ query: AppListQueryDocument }],
      }
    );
  const reconcileCheckoutSession = React.useCallback(
    async (appID: string, checkoutSessionID: string) => {
      await mutationFunction({
        variables: { appID, checkoutSessionID },
      });
    },
    [mutationFunction]
  );
  return { reconcileCheckoutSession, error, loading };
}
