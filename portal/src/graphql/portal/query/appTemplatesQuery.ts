import { useMemo } from "react";
import { gql, QueryResult, useQuery } from "@apollo/client";

import { client } from "../apollo";
import {
  AppTemplatesQuery,
  AppTemplatesQueryVariables,
} from "./__generated__/AppTemplatesQuery";
import {
  getLocalizedTemplatePath,
  TemplateLocale,
  TemplateMap,
} from "../../../templates";
import { ResourcePath } from "../../../util/stringTemplate";

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
        resourceLocales: resources {
          path
          languageTag
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
  templateLocales: TemplateLocale[];
}

export function useAppTemplatesQuery(
  appID: string,
  locale: TemplateLocale,
  ...pathTemplates: ResourcePath<"locale">[]
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

    const templateLocaleSets = new Set<TemplateLocale>();
    const templateResourceData =
      appNode?.resourceLocales.filter((resourceData) => {
        return resourceData.path.split("/")[0] === "templates";
      }) ?? [];
    for (const resourceData of templateResourceData) {
      const locale = resourceData.languageTag;
      if (locale != null) templateLocaleSets.add(locale);
    }
    return { templates, templateLocales: Array.from(templateLocaleSets) };
  }, [data, paths]);

  return { ...queryData, loading, error, refetch };
}
