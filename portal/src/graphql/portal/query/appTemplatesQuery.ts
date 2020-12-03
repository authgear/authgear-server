import { useMemo } from "react";
import { gql, QueryResult, useQuery } from "@apollo/client";

import { client } from "../apollo";
import {
  AppTemplatesQuery,
  AppTemplatesQueryVariables,
} from "./__generated__/AppTemplatesQuery";
import { renderPath } from "../../../resources";
import {
  Resource,
  ResourceSpecifier,
  decodeForText,
  binary,
} from "../../../util/resource";

export const appTemplatesQuery = gql`
  query AppTemplatesQuery($id: ID!, $paths: [String!]!) {
    node(id: $id) {
      __typename
      ... on App {
        id
        resources(paths: $paths) {
          path
          languageTag
          data
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
  resources: Resource[];
}

interface SpecifierPathPair {
  specifier: ResourceSpecifier;
  path: string;
}

export function useAppTemplatesQuery(
  appID: string,
  specifiers: ResourceSpecifier[]
): AppTemplatesQueryResult {
  const pairs: SpecifierPathPair[] = useMemo(() => {
    const pairs = [];
    for (const specifier of specifiers) {
      if (specifier.def.extensions.length === 0) {
        pairs.push({
          specifier,
          path: renderPath(specifier.def.resourcePath, {
            locale: specifier.locale,
          }),
        });
      } else {
        for (const extension of specifier.def.extensions) {
          pairs.push({
            specifier,
            path: renderPath(specifier.def.resourcePath, {
              extension,
              locale: specifier.locale,
            }),
          });
        }
      }
    }
    return pairs;
  }, [specifiers]);

  const { data, loading, error, refetch } = useQuery<
    AppTemplatesQuery,
    AppTemplatesQueryVariables
  >(appTemplatesQuery, {
    client,
    variables: {
      id: appID,
      paths: pairs.map((pair) => pair.path),
    },
  });

  // eslint-disable-next-line complexity
  const resources = useMemo(() => {
    const appNode = data?.node?.__typename === "App" ? data.node : null;
    const resources: Resource[] = [];

    for (const { specifier, path } of pairs) {
      let transform: (a: string) => string;
      switch (specifier.def.type) {
        case "text":
          transform = decodeForText;
          break;
        case "binary":
          transform = binary;
          break;
        default:
          throw new Error(
            "unexpected resource type: " + String(specifier.def.type)
          );
      }

      let value = "";
      const resource = (appNode?.resources ?? []).find((r) => r.path === path);
      if (resource?.data != null) {
        value = transform(resource.data);
      } else if (
        resource?.effectiveData != null &&
        specifier.def.usesEffectiveDataAsFallbackValue
      ) {
        value = transform(resource.effectiveData);
      }

      resources.push({
        specifier,
        path,
        value,
      });
    }

    return resources;
  }, [data, pairs]);

  return { resources, loading, error, refetch };
}
