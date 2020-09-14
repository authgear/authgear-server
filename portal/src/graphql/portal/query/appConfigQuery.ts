import { gql, QueryResult, useQuery } from "@apollo/client";
import { client } from "../../portal/apollo";
import { AppConfigQuery } from "./__generated__/AppConfigQuery";

export const appConfigQuery = gql`
  query AppConfigQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        effectiveAppConfig
      }
    }
  }
`;

export const useAppConfigQuery = (
  appID: string
): QueryResult<AppConfigQuery, Record<string, unknown>> => {
  const appConfigQueryResult = useQuery<AppConfigQuery>(appConfigQuery, {
    client,
    variables: {
      id: appID,
    },
  });
  return appConfigQueryResult;
};
