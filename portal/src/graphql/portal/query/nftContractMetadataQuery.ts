import { useLazyQuery } from "@apollo/client";
import { useCallback } from "react";
import { client } from "../apollo";
import { NftCollection } from "../globalTypes.generated";
import {
  NftContractMetadataQueryQuery,
  NftContractMetadataQueryDocument,
} from "./nftContractMetadataQuery.generated";

interface NftContractMetadataQueryResult {
  fetch: (contractId: string) => Promise<NftCollection | null>;
  loading: boolean;
  error: unknown;
}

export function useNftContractMetadataLazyQuery(): NftContractMetadataQueryResult {
  const [fetch, { loading, error }] =
    useLazyQuery<NftContractMetadataQueryQuery>(
      NftContractMetadataQueryDocument,
      {
        client,
      }
    );

  const fetchData = useCallback(
    async (contractId: string) => {
      const res = await fetch({
        variables: {
          contractID: contractId,
        },
      });

      return res.data?.nftContractMetadata ?? null;
    },
    [fetch]
  );

  return { fetch: fetchData, loading, error };
}
