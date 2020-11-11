import { useCallback } from "react";
import { gql } from "@apollo/client";

import { useGraphqlMutation } from "../../../hook/graphql";
import { client } from "../apollo";
import { AppResourceUpdate } from "../__generated__/globalTypes";
import {
  UpdateAppTemplatesMutation,
  UpdateAppTemplatesMutationVariables,
} from "./__generated__/UpdateAppTemplatesMutation";
import { PortalAPIApp } from "../../../types";
import {
  getLocalizedTemplatePath,
  PathTemplate,
  TemplateLocale,
} from "../../../templates";

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
          effectiveData
        }
      }
    }
  }
`;

export type UpdateAppTemplatesData = Partial<Record<string, string | null>>;

export type AppTemplatesUpdater = (
  updateTemplates: UpdateAppTemplatesData
) => Promise<PortalAPIApp | null>;

export function useUpdateAppTemplatesMutation(
  appID: string,
  locale: TemplateLocale,
  ...pathTemplates: PathTemplate[]
): {
  updateAppTemplates: AppTemplatesUpdater;
  loading: boolean;
  error: unknown;
  resetError: () => void;
} {
  const [mutationFunction, { error, loading }, resetError] = useGraphqlMutation<
    UpdateAppTemplatesMutation,
    UpdateAppTemplatesMutationVariables
  >(updateAppTemplatesMutation, { client });
  const updateAppTemplates = useCallback(
    async (updateTemplates: UpdateAppTemplatesData) => {
      const updates: AppResourceUpdate[] = [];
      for (const [path, data] of Object.entries(updateTemplates)) {
        if (data === undefined) {
          continue;
        }
        updates.push({
          path,
          data,
        });
      }

      const paths = pathTemplates.map((pathTemplate) =>
        getLocalizedTemplatePath(locale, pathTemplate)
      );

      const result = await mutationFunction({
        variables: {
          appID,
          paths,
          updates,
        },
      });
      return result.data?.updateAppResources.app ?? null;
    },
    [appID, mutationFunction, locale, pathTemplates]
  );
  return { updateAppTemplates, error, loading, resetError };
}
