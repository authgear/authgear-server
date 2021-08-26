import { gql, QueryResult, useQuery } from "@apollo/client";
import { useMemo } from "react";
import { client } from "../../portal/apollo";
import {
  DomainsQuery,
  DomainsQueryVariables,
  DomainsQuery_node_App_domains,
} from "./__generated__/DomainsQuery";

export const domainsQuery = gql`
  query DomainsQuery($appID: ID!) {
    node(id: $appID) {
      ... on App {
        id
        domains {
          id
          createdAt
          apexDomain
          domain
          cookieDomain
          isCustom
          isVerified
          verificationDNSRecord
        }
      }
    }
  }
`;

export type Domain = DomainsQuery_node_App_domains;
interface DomainsQueryResult
  extends Pick<
    QueryResult<DomainsQuery, DomainsQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  domains: Domain[] | null;
}

export function useDomainsQuery(appID: string): DomainsQueryResult {
  const { data, loading, error, refetch } = useQuery<DomainsQuery>(
    domainsQuery,
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
