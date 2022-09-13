import {
  LazyQueryResult,
  OperationVariables,
  useLazyQuery,
} from "@apollo/client";
import { useCallback } from "react";
import { client } from "../apollo";
import {
  NftContractMetadataQueryQuery,
  NftContractMetadataQueryDocument,
} from "./nftContractMetadataQuery.generated";

interface NftContractMetadataQueryResult {
  fetch: (
    contractId: string
  ) => Promise<
    LazyQueryResult<NftContractMetadataQueryQuery, OperationVariables>
  >;
  loading: boolean;
  error: unknown;
}

export function useNftContractMetadataLazyQuery(
  appID: string
): NftContractMetadataQueryResult {
  const [fetch, { loading, error }] =
    useLazyQuery<NftContractMetadataQueryQuery>(
      NftContractMetadataQueryDocument,
      {
        client,
      }
    );

  const fetchData = useCallback(
    async (contractId: string) => {
      return fetch({
        variables: {
          appID,
          contractID: contractId,
        },
      });
    },
    [appID, fetch]
  );

  return { fetch: fetchData, loading, error };
}
