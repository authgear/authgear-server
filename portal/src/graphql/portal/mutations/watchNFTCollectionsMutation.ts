import { useCallback } from "react";
import { useMutation } from "@apollo/client";

import { client } from "../apollo";
import {
  WatchNftCollectionsMutationMutation,
  WatchNftCollectionsMutationDocument,
} from "./watchNFTCollectionsMutation.generated";

export function useWatchNFTCollectionsMutation(appID: string): {
  watchNFTCollections: (contractIds: string[]) => Promise<void>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<WatchNftCollectionsMutationMutation>(
      WatchNftCollectionsMutationDocument,
      {
        client,
      }
    );
  const watchNFTCollections = useCallback(
    async (contractIds: string[]) => {
      await mutationFunction({
        variables: { appID, contractIDs: contractIds },
      });
    },
    [mutationFunction, appID]
  );

  return { watchNFTCollections, error, loading };
}
