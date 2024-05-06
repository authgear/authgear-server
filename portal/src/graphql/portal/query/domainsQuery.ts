import { QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";
import { usePortalClient } from "../../portal/apollo";
import { Domain } from "../globalTypes.generated";
import {
  DomainsQueryQuery,
  DomainsQueryQueryVariables,
  DomainsQueryDocument,
} from "./domainsQuery.generated";

interface DomainsQueryResult
  extends Pick<
    QueryResult<DomainsQueryQuery, DomainsQueryQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  domains: Domain[] | null;
}

export function useDomainsQuery(appID: string): DomainsQueryResult {
  const client = usePortalClient();
  const { data, loading, error, refetch } = useQuery<DomainsQueryQuery>(
    DomainsQueryDocument,
    {
      client,
      variables: {
        appID,
      },
    }
  );

  const domains = useMemo(() => {
    const appNode = data?.node?.__typename === "App" ? data.node : null;
    return appNode?.domains ?? null;
  }, [data]);

  return { domains, loading, error, refetch };
}
