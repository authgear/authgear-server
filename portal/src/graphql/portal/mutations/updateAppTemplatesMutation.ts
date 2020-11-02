import React from "react";
import { gql } from "@apollo/client";

import { useGraphqlMutation } from "../../../hook/graphql";
import { client } from "../apollo";
import { AppResourceUpdate } from "../__generated__/globalTypes";
import {
  UpdateAppTemplatesMutation,
  UpdateAppTemplatesMutationVariables,
} from "./__generated__/UpdateAppTemplatesMutation";
import { PortalAPIApp } from "../../../types";

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

export type UpdateAppTemplatesData<TemplatePath extends string> = {
  [path in TemplatePath]?: string | null;
};

export type AppTemplatesUpdater<TemplatePath extends string> = (
  updateTemplates: UpdateAppTemplatesData<TemplatePath>
) => Promise<PortalAPIApp | null>;

export function useUpdateAppTemplatesMutation<TemplatePath extends string>(
  appID: string,
  ...paths: TemplatePath[]
): {
  updateAppTemplates: AppTemplatesUpdater<TemplatePath>;
  loading: boolean;
  error: unknown;
  resetError: () => void;
} {
  const [mutationFunction, { error, loading }, resetError] = useGraphqlMutation<
    UpdateAppTemplatesMutation,
    UpdateAppTemplatesMutationVariables
  >(updateAppTemplatesMutation, { client });
  const updateAppTemplates = React.useCallback(
    async (updateTemplates: { [path in TemplatePath]?: string | null }) => {
      const updates: AppResourceUpdate[] = [];
      for (const [path, data] of Object.entries(updateTemplates)) {
        if (data === undefined) {
          continue;
        }
        updates.push({ path, data: data as string | null });
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
    [appID, mutationFunction, paths]
  );
  return { updateAppTemplates, error, loading, resetError };
}
