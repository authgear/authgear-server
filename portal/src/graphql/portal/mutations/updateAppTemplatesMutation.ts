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
import { TemplateLocale } from "../../../templates";

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

export type AppTemplatesUpdater = (
  paths: string[],
  updates: AppResourceUpdate[]
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
    async (paths: string[], updates: AppResourceUpdate[]) => {
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
