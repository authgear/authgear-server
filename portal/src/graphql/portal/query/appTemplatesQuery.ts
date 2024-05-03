import { useMemo } from "react";
import { QueryResult, useQuery } from "@apollo/client";

import { usePortalClient } from "../apollo";
import {
  AppTemplatesQueryQuery,
  AppTemplatesQueryQueryVariables,
  AppTemplatesQueryDocument,
} from "./appTemplatesQuery.generated";
import {
  Resource,
  ResourceSpecifier,
  decodeForText,
  binary,
  expandSpecifier,
} from "../../../util/resource";

export interface AppTemplatesQueryResult
  extends Pick<
    QueryResult<AppTemplatesQueryQuery, AppTemplatesQueryQueryVariables>,
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
  const client = usePortalClient();
  const pairs: SpecifierPathPair[] = useMemo(() => {
    const pairs = [];
    for (const specifier of specifiers) {
      const path = expandSpecifier(specifier);
      pairs.push({
        specifier,
        path,
      });
    }
    return pairs;
  }, [specifiers]);

  const { data, loading, error, refetch } = useQuery<
    AppTemplatesQueryQuery,
    AppTemplatesQueryQueryVariables
  >(AppTemplatesQueryDocument, {
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

      let value;
      const resource = (appNode?.resources ?? []).find((r) => r.path === path);
      if (resource?.data != null) {
        value = transform(resource.data);
      } else if (
        specifier.def.fallback?.kind === "EffectiveData" &&
        resource?.effectiveData != null
      ) {
        value = transform(resource.effectiveData);
      } else if (specifier.def.fallback?.kind === "Const") {
        value = specifier.def.fallback.fallbackValue;
      }

      const effectiveData =
        resource?.effectiveData && transform(resource.effectiveData);

      resources.push({
        specifier,
        path,
        nullableValue: value,
        effectiveData: effectiveData,
        checksum: resource?.checksum,
      });
    }

    return resources;
  }, [data, pairs]);

  return { resources, loading, error, refetch };
}
