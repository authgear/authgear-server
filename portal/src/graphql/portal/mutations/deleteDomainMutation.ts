import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";

import { client } from "../../portal/apollo";
import { DeleteDomainMutation } from "./__generated__/DeleteDomainMutation";

const deleteDomainMutation = gql`
  mutation DeleteDomainMutation($appID: ID!, $domainID: String!) {
    deleteDomain(input: { appID: $appID, domainID: $domainID }) {
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
        rawAppConfig
        effectiveAppConfig
      }
    }
  }
`;

export function useDeleteDomainMutation(
  appID: string
): {
  deleteDomain: (domainID: string) => Promise<boolean>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    DeleteDomainMutation
  >(deleteDomainMutation, {
    client,
  });

  const deleteDomain = useCallback(
    async (domainID: string) => {
      const result = await mutationFunction({
        variables: { appID, domainID },
      });

      return result.data != null;
    },
    [mutationFunction, appID]
  );

  return { deleteDomain, error, loading };
}
