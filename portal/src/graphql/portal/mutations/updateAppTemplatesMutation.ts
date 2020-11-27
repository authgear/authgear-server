import { useCallback } from "react";
import { gql } from "@apollo/client";
import { useGraphqlMutation } from "../../../hook/graphql";
import { client } from "../apollo";
import { AppResourceUpdate } from "../__generated__/globalTypes";
import {
  UpdateAppTemplatesMutation,
  UpdateAppTemplatesMutationVariables,
} from "./__generated__/UpdateAppTemplatesMutation";
import { renderPath } from "../../../resources";
import { PortalAPIApp } from "../../../types";
import {
  ResourceUpdate,
  ResourceSpecifier,
  binary,
  encodeForText,
} from "../../../util/resource";

const updateAppTemplatesMutation = gql`
  mutation UpdateAppTemplatesMutation(
    $appID: ID!
    $updates: [AppResourceUpdate!]!
    $paths: [String!]!
  ) {
    updateAppResources(input: { appID: $appID, updates: $updates }) {
      app {
        id
        resources(paths: $paths) {
          path
          languageTag
          data
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

export type AppTemplatesUpdater = (
  specifiers: ResourceSpecifier[],
  updates: ResourceUpdate[]
) => Promise<PortalAPIApp | null>;

export function useUpdateAppTemplatesMutation(
  appID: string
): {
  updateAppTemplates: AppTemplatesUpdater;
  loading: boolean;
  error: unknown;
  resetError: () => void;
} {
  const [mutationFunction, { error, loading }, resetError] = useGraphqlMutation<
    UpdateAppTemplatesMutation,
    UpdateAppTemplatesMutationVariables
  >(updateAppTemplatesMutation, {
    client,
    // FIXME: I cannot figure out the rendered query does not rerender :(
    refetchQueries: ["AppTemplatesQuery", "TemplateLocaleQuery"],
    awaitRefetchQueries: true,
  });
  const updateAppTemplates = useCallback(
    async (specifiers: ResourceSpecifier[], updates: ResourceUpdate[]) => {
      const paths = [];
      for (const specifier of specifiers) {
        if (specifier.def.extensions.length === 0) {
          paths.push(
            renderPath(specifier.def.resourcePath, {
              locale: specifier.locale,
            })
          );
        } else {
          for (const extension of specifier.def.extensions) {
            paths.push(
              renderPath(specifier.def.resourcePath, {
                locale: specifier.locale,
                extension,
              })
            );
          }
        }
      }

      const updatePayload: AppResourceUpdate[] = updates.map((update) => {
        let transform: (a: string) => string;
        switch (update.specifier.def.type) {
          case "text":
            transform = encodeForText;
            break;
          case "binary":
            transform = binary;
            break;
          default:
            throw new Error(
              "unexpected resource type: " + String(update.specifier.def.type)
            );
        }
        return {
          path: update.path,
          data: update.value == null ? null : transform(update.value),
        };
      });

      const result = await mutationFunction({
        variables: {
          appID,
          paths,
          updates: updatePayload,
        },
      });
      return result.data?.updateAppResources.app ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppTemplates, error, loading, resetError };
}
