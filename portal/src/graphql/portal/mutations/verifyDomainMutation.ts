import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";

import { client } from "../../portal/apollo";
import { VerifyDomainMutation } from "./__generated__/VerifyDomainMutation";

const verifyDomainMutation = gql`
  mutation VerifyDomainMutation($appID: ID!, $domainID: String!) {
    verifyDomain(input: { appID: $appID, domainID: $domainID }) {
      app {
        id
        domains {
          id
          createdAt
          domain
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
        apexDomain
        isCustom
        isVerified
        verificationDNSRecord
      }
    }
  }
`;

export function useVerifyDomainMutation(appID: string): {
  verifyDomain: (domainID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] =
    useMutation<VerifyDomainMutation>(verifyDomainMutation, { client });
  const verifyDomain = useCallback(
    async (domainID: string) => {
      const result = await mutationFunction({
        variables: { appID, domainID },
      });

      return result.data != null;
    },
    [mutationFunction, appID]
  );

  return { verifyDomain, error, loading };
}
