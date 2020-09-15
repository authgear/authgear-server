import { gql, QueryResult, useQuery } from "@apollo/client";
import { client } from "../../portal/apollo";
import { AppAndSecretConfigQuery } from "./__generated__/AppAndSecretConfigQuery";

export const appAndSecretConfigQuery = gql`
  query AppAndSecretConfigQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        effectiveAppConfig
        rawAppConfig
        rawSecretConfig
      }
    }
  }
`;

export const useAppAndSecretConfigQuery = (
  appID: string
): QueryResult<AppAndSecretConfigQuery, Record<string, unknown>> => {
  const appAndSecretConfigQueryResult = useQuery<AppAndSecretConfigQuery>(
    appAndSecretConfigQuery,
    {
      client,
      variables: {
        id: appID,
      },
    }
  );
  return appAndSecretConfigQueryResult;
};
