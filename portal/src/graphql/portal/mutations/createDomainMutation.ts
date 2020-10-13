import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";

import { client } from "../../portal/apollo";
import { CreateDomainMutation } from "./__generated__/CreateDomainMutation";

const createDomainMutation = gql`
  mutation CreateDomainMutation($appID: ID!, $domain: String!) {
    createDomain(input: { appID: $appID, domain: $domain }) {
      app {
        id
      }
      domain {
        id
        createdAt
        domain
        apexDomain
        isCustom
        isVerified
        verificationDNSRecord
      }
    }
  }
`;

export function useCreateDomainMutation(
  appID: string
): {
  createDomain: (domain: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    CreateDomainMutation
  >(createDomainMutation, {
    client,
    // TODO: backend return whole list of domains so apollo can
    // automatically update the domain list
    refetchQueries: ["DomainsQuery"],
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
