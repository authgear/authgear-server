import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";

import { client } from "../../portal/apollo";
import { CreateDomainMutation } from "./__generated__/CreateDomainMutation";

const createDomainMutation = gql`
  mutation CreateDomainMutation($appID: ID!, $domain: String!) {
    createDomain(input: { appID: $appID, domain: $domain }) {
      app {
        id
        domains {
          id
          createdAt
          domain
          cookieDomain
          apexDomain
          isCustom
          isVerified
          verificationDNSRecord
        }
      }
      domain {
        id
        createdAt
        domain
        cookieDomain
        apexDomain
        isCustom
        isVerified
        verificationDNSRecord
      }
    }
  }
`;

export function useCreateDomainMutation(appID: string): {
  createDomain: (domain: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<CreateDomainMutation>(createDomainMutation, {
      client,
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
