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
import { TemplateLocale, ALL_TEMPLATE_PATHS } from "../../../templates";

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
        resourceLocales: resources {
          path
          languageTag
        }
      }
    }
  }
`;

export type UpdateAppTemplatesData = Partial<Record<string, string | null>>;

export type AppTemplatesUpdater = (
  updateTemplates: UpdateAppTemplatesData
) => Promise<PortalAPIApp | null>;

export type TemplateLocaleRemover = (
  locales: TemplateLocale[]
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
  >(updateAppTemplatesMutation, { client });
  const updateAppTemplates = useCallback(
    async (updateTemplates: UpdateAppTemplatesData) => {
      const updates: AppResourceUpdate[] = [];
      const paths: string[] = [];
      for (const [path, data] of Object.entries(updateTemplates)) {
        if (data === undefined) {
          continue;
        }
        updates.push({
          path,
          data,
        });
        paths.push(path);
      }

      const result = await mutationFunction({
        variables: {
          appID,
          paths,
          updates,
        },
      });
      return result.data?.updateAppResources.app ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppTemplates, error, loading, resetError };
}

export function useRemoveTemplateLocalesMutation(
  appID: string
): {
  removeTemplateLocales: TemplateLocaleRemover;
  loading: boolean;
  error: unknown;
  resetError: () => void;
} {
  const [mutationFunction, { error, loading }, resetError] = useGraphqlMutation<
    UpdateAppTemplatesMutation,
    UpdateAppTemplatesMutationVariables
  >(updateAppTemplatesMutation, { client });
  const removeTemplateLocales = useCallback<TemplateLocaleRemover>(
    async (locales: TemplateLocale[]) => {
      // all message template path
      const updates: AppResourceUpdate[] = [];
      const paths: string[] = [];
      for (const templatePath of ALL_TEMPLATE_PATHS) {
        for (const locale of locales) {
          const path = templatePath.render({ locale });
          updates.push({
            path,
            data: null,
          });
          paths.push(path);
        }
      }
      const result = await mutationFunction({
        variables: {
          appID,
          paths,
          updates,
        },
      });
      return result.data?.updateAppResources.app ?? null;
    },
    [appID, mutationFunction]
  );
  return { removeTemplateLocales, error, loading, resetError };
}
