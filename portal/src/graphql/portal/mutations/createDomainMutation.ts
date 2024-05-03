import { useCallback } from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../../portal/apollo";
import {
  CreateDomainMutationMutation,
  CreateDomainMutationDocument,
} from "./createDomainMutation.generated";

export function useCreateDomainMutation(appID: string): {
  createDomain: (domain: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] =
    useMutation<CreateDomainMutationMutation>(CreateDomainMutationDocument, {
      client,
    });
  const createDomain = useCallback(
    async (domain: string) => {
      const result = await mutationFunction({
        variables: { appID, domain },
      });

      return result.data != null;
    },
    [mutationFunction, appID]
  );

  return { createDomain, error, loading };
}
