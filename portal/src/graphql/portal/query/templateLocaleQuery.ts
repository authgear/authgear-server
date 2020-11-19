import { useMemo } from "react";
import { gql, QueryResult, useQuery } from "@apollo/client";

import { client } from "../apollo";
import {
  TemplateLocaleQuery,
  TemplateLocaleQueryVariables,
} from "./__generated__/TemplateLocaleQuery";
import { TemplateLocale } from "../../../templates";

export const templateLocaleQuery = gql`
  query TemplateLocaleQuery($id: ID!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        resourceLocales: resources {
          path
          languageTag
        }
      }
    }
  }
`;

export interface TemplateLocaleQueryResult
  extends Pick<
    QueryResult<TemplateLocaleQuery, TemplateLocaleQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  templateLocales: TemplateLocale[];
}

export function useTemplateLocaleQuery(
  appID: string
): TemplateLocaleQueryResult {
  const { data, loading, error, refetch } = useQuery<
    TemplateLocaleQuery,
    TemplateLocaleQueryVariables
  >(templateLocaleQuery, {
    client,
    variables: {
      id: appID,
    },
  });

  const queryData = useMemo(() => {
    const appNode = data?.node?.__typename === "App" ? data.node : null;
    const templateLocaleSets = new Set<TemplateLocale>();
    const templateResourceData =
      appNode?.resourceLocales.filter((resourceData) => {
        return resourceData.path.split("/")[0] === "templates";
      }) ?? [];
    for (const resourceData of templateResourceData) {
      const locale = resourceData.languageTag;
      if (locale != null) templateLocaleSets.add(locale);
    }
    return { templateLocales: Array.from(templateLocaleSets) };
  }, [data]);

  return { ...queryData, loading, error, refetch };
}
