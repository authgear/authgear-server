import { gql, QueryResult, useQuery } from "@apollo/client";
import { client } from "../../portal/apollo";
import { AppListQuery } from "./__generated__/AppListQuery";

export const appListQuery = gql`
  query AppListQuery {
    apps {
      edges {
        node {
          id
          effectiveAppConfig
        }
      }
    }
  }
`;

export const useAppListQuery = (): QueryResult<AppListQuery> => {
  return useQuery<AppListQuery>(appListQuery, { client });
};
