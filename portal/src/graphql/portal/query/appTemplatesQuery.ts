import { useMemo } from "react";
import { gql, QueryResult, useQuery } from "@apollo/client";

import { client } from "../apollo";
import {
  AppTemplatesQuery,
  AppTemplatesQueryVariables,
} from "./__generated__/AppTemplatesQuery";
import { getLocalizedTemplatePath, TemplateLocale } from "../../../templates";
import { ResourcePath } from "../../../util/stringTemplate";

export const appTemplatesQuery = gql`
  query AppTemplatesQuery($id: ID!, $paths: [String!]!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        resources(paths: $paths) {
          path
          languageTag
          effectiveData
        }
      }
    }
  }
`;

export interface Template {
  locale: TemplateLocale;
  resourcePath: ResourcePath<"locale">;
  path: string;
  value: string;
}

export interface AppTemplatesQueryResult
  extends Pick<
    QueryResult<AppTemplatesQuery, AppTemplatesQueryVariables>,
    "loading" | "error" | "refetch"
  > {
  templates: Record<string, Template>;
}

export interface InputPath {
  locale: TemplateLocale;
  resourcePath: ResourcePath<"locale">;
  path: string;
}

export function useAppTemplatesQuery(
  appID: string,
  locales: TemplateLocale[],
  ...resourcePaths: ResourcePath<"locale">[]
): AppTemplatesQueryResult {
  const inputPaths = useMemo<InputPath[]>(() => {
    const output: InputPath[] = [];
    for (const locale of locales) {
      for (const resourcePath of resourcePaths) {
        output.push({
          locale,
          resourcePath,
          path: getLocalizedTemplatePath(locale, resourcePath),
        });
      }
    }
    return output;
  }, [locales, resourcePaths]);

  const paths = useMemo(() => inputPaths.map((inputPath) => inputPath.path), [
    inputPaths,
  ]);

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

  const templates = useMemo(() => {
    const appNode = data?.node?.__typename === "App" ? data.node : null;
    const templates: Record<string, Template> = {};

    for (const inputPath of inputPaths) {
      let found = false;

      for (const resource of appNode?.resources ?? []) {
        if (inputPath.path === resource.path) {
          found = true;
          let value = "";
          if (resource.effectiveData != null) {
            value = atob(resource.effectiveData);
          }
          templates[inputPath.path] = {
            ...inputPath,
            value,
          };
          break;
        }
      }

      if (!found) {
        templates[inputPath.path] = {
          ...inputPath,
          value: "",
        };
      }
    }

    return templates;
  }, [data, inputPaths]);

  return { templates, loading, error, refetch };
}
