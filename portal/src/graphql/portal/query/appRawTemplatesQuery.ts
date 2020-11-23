import { useMemo } from "react";
import { gql, QueryResult, useQuery } from "@apollo/client";

import { client } from "../apollo";
import {
  AppRawTemplatesQuery,
  AppRawTemplatesQueryVariables,
} from "./__generated__/AppRawTemplatesQuery";

export const appRawTemplatesQuery = gql`
  query AppRawTemplatesQuery($id: ID!, $paths: [String!]!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        resources(paths: $paths) {
          path
          data
        }
      }
    }
  }
`;

export interface AppRawTemplatesQueryResult<TemplatePath extends string>
  extends Pick<
    QueryResult<AppRawTemplatesQuery, AppRawTemplatesQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  templates: Record<TemplatePath, string | null>;
}

export const useAppRawTemplatesQuery = <TemplatePath extends string>(
  appID: string,
  ...paths: TemplatePath[]
): AppRawTemplatesQueryResult<TemplatePath> => {
  const { data, loading, error, refetch } = useQuery<
    AppRawTemplatesQuery,
    AppRawTemplatesQueryVariables
  >(appRawTemplatesQuery, {
    client,
    variables: {
      id: appID,
      paths,
    },
  });

  const queryData = useMemo(() => {
    const appNode = data?.node?.__typename === "App" ? data.node : null;
    const templates = {} as Record<TemplatePath, string | null>;
    for (const { path, data } of appNode?.resources ?? []) {
      let value = "";
      if (data != null) {
        value = atob(data);
      }
      templates[path as TemplatePath] = value;
    }
    for (const path of paths) {
      if (!(path in templates)) {
        templates[path] = null;
      }
    }
    return { templates };
  }, [data, paths]);

  return { ...queryData, loading, error, refetch };
};
