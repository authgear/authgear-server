import { useMemo } from "react";
import { gql, QueryResult, useQuery } from "@apollo/client";

import { client } from "../apollo";
import {
  AppTemplatesQuery,
  AppTemplatesQueryVariables,
} from "./__generated__/AppTemplatesQuery";
import {
  getLocalizedTemplatePath,
  PathTemplate,
  TemplateLocale,
  TemplateMap,
} from "../../../templates";

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

export interface AppTemplatesQueryResult
  extends Pick<
    QueryResult<AppTemplatesQuery, AppTemplatesQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  templates: Record<string, string>;
}

export function useAppTemplatesQuery(
  appID: string,
  locale: TemplateLocale,
  ...pathTemplates: PathTemplate[]
): AppTemplatesQueryResult {
  const paths = useMemo(
    () =>
      pathTemplates.map((pathTemplate) =>
        getLocalizedTemplatePath(locale, pathTemplate)
      ),
    [locale, pathTemplates]
  );

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
    const templates: TemplateMap = {};
    for (const { path, effectiveData } of appNode?.resources ?? []) {
      templates[path] = effectiveData ?? "";
    }
    for (const path of paths) {
      if (!(path in templates)) {
        templates[path] = "";
      }
    }
    return { templates };
  }, [data, paths]);

  return { ...queryData, loading, error, refetch };
}
