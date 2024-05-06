import { useCallback } from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../apollo";
import {
  ProbeNftCollectionMutationMutation,
  ProbeNftCollectionMutationDocument,
} from "./probeNFTCollectionMutation.generated";

export function useProbeNFTCollectionMutation(): {
  probeNFTCollection: (contractId: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] =
    useMutation<ProbeNftCollectionMutationMutation>(
      ProbeNftCollectionMutationDocument,
      {
        client,
      }
    );
  const probeNFTCollection = useCallback(
    async (contractId: string) => {
      const res = await mutationFunction({
        variables: { contractID: contractId },
      });
      return res.data?.probeNFTCollection.isLargeCollection ?? false;
    },
    [mutationFunction]
  );

  return { probeNFTCollection, error, loading };
}
