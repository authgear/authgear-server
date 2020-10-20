import { useMemo } from "react";
import { gql, QueryResult, useQuery } from "@apollo/client";

import { client } from "../apollo";
import {
  AppTemplatesQuery,
  AppTemplatesQueryVariables,
} from "./__generated__/AppTemplatesQuery";

export const appTemplatesQuery = gql`
  query AppTemplatesQuery($id: ID!, $paths: [String!]!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        resources(paths: $paths) {
          path
          effectiveData
        }
      }
    }
  }
`;

export interface AppTemplatesQueryResult<TemplatePath extends string>
  extends Pick<
    QueryResult<AppTemplatesQuery, AppTemplatesQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  templates: Record<TemplatePath, string>;
}

export const useAppTemplatesQuery = <TemplatePath extends string>(
  appID: string,
  ...paths: TemplatePath[]
): AppTemplatesQueryResult<TemplatePath> => {
  const { data, loading, error, refetch } = useQuery<
    AppTemplatesQuery,
    AppTemplatesQueryVariables
  >(appTemplatesQuery, {
    client,
    variables: {
      id: appID,
      paths,
    },
  });

  const queryData = useMemo(() => {
    const appNode = data?.node?.__typename === "App" ? data.node : null;
    const templates = {} as Record<TemplatePath, string>;
    for (const { path, effectiveData } of appNode?.resources ?? []) {
      templates[path as TemplatePath] = effectiveData ?? "";
    }
    for (const path of paths) {
      if (!(path in templates)) {
        templates[path] = "";
      }
    }
    return { templates };
  }, [data, paths]);

  return { ...queryData, loading, error, refetch };
};
