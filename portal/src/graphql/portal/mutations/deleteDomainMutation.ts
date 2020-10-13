import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";

import { client } from "../../portal/apollo";
import { DeleteDomainMutation } from "./__generated__/DeleteDomainMutation";

const deleteDomainMutation = gql`
  mutation DeleteDomainMutation($appID: ID!, $domainID: String!) {
    deleteDomain(input: { appID: $appID, domainID: $domainID }) {
      app {
        id
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
    // TODO: backend return whole list of domains so apollo can
    // automatically update the domain list
    refetchQueries: ["DomainsQuery"],
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
