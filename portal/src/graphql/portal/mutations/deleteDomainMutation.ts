import { useCallback } from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../../portal/apollo";
import {
  DeleteDomainMutationMutation,
  DeleteDomainMutationDocument,
} from "./deleteDomainMutation.generated";

export function useDeleteDomainMutation(appID: string): {
  deleteDomain: (domainID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] =
    useMutation<DeleteDomainMutationMutation>(DeleteDomainMutationDocument, {
      client,
    });

  const deleteDomain = useCallback(
    async (domainID: string) => {
      const result = await mutationFunction({
        variables: { appID, domainID },
      });

      return result.data != null;
    },
    [mutationFunction, appID]
  );

  return { deleteDomain, error, loading };
}
