import React from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../../portal/apollo";
import {
  CreateCheckoutSessionMutationMutation,
  CreateCheckoutSessionMutationDocument,
} from "./createCheckoutSessionMutation.generated";

export function useCreateCheckoutSessionMutation(): {
  createCheckoutSession: (
    appID: string,
    planName: string
  ) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] =
    useMutation<CreateCheckoutSessionMutationMutation>(
      CreateCheckoutSessionMutationDocument,
      {
        client,
      }
    );
  const createCheckoutSession = React.useCallback(
    async (appID: string, planName: string) => {
      const result = await mutationFunction({
        variables: { appID, planName },
      });
      return result.data?.createCheckoutSession.url ?? null;
    },
    [mutationFunction]
  );
  return { createCheckoutSession, error, loading };
}
