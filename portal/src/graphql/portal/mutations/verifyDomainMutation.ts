import { useCallback } from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../../portal/apollo";
import {
  VerifyDomainMutationMutation,
  VerifyDomainMutationDocument,
} from "./verifyDomainMutation.generated";

export function useVerifyDomainMutation(appID: string): {
  verifyDomain: (domainID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] =
    useMutation<VerifyDomainMutationMutation>(VerifyDomainMutationDocument, {
      client,
    });
  const verifyDomain = useCallback(
    async (domainID: string) => {
      const result = await mutationFunction({
        variables: { appID, domainID },
      });

      return result.data != null;
    },
    [mutationFunction, appID]
  );

  return { verifyDomain, error, loading };
}
