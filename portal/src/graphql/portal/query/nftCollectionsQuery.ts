import { QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";
import { usePortalClient } from "../../portal/apollo";
import { NftCollection } from "../globalTypes.generated";
import {
  NftCollectionsQueryQuery,
  NftCollectionsQueryQueryVariables,
  NftCollectionsQueryDocument,
} from "./nftCollectionsQuery.generated";

interface NftCollectionsQueryResult
  extends Pick<
    QueryResult<NftCollectionsQueryQuery, NftCollectionsQueryQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  collections: NftCollection[];
}

export function useNftCollectionsQuery(
  appID: string
): NftCollectionsQueryResult {
  const client = usePortalClient();
  const { data, loading, error, refetch } = useQuery<NftCollectionsQueryQuery>(
    NftCollectionsQueryDocument,
    {
      client,
      variables: {
        appID,
      },
    }
  );

  const collections = useMemo(() => {
    const appNode = data?.node?.__typename === "App" ? data.node : null;
    return appNode?.nftCollections ?? [];
  }, [data]);

  return { collections, loading, error, refetch };
}
