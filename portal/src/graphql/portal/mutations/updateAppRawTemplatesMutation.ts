import React from "react";
import { useMutation, gql } from "@apollo/client";

import { client } from "../apollo";
import { AppResourceUpdate } from "../__generated__/globalTypes";
import {
  UpdateAppRawTemplatesMutation,
  UpdateAppRawTemplatesMutationVariables,
} from "./__generated__/UpdateAppRawTemplatesMutation";
import { PortalAPIApp } from "../../../types";

const updateAppRawTemplatesMutation = gql`
  mutation UpdateAppRawTemplatesMutation(
    $appID: ID!
    $updates: [AppResourceUpdate!]!
    $paths: [String!]!
  ) {
    updateAppResources(input: { appID: $appID, updates: $updates }) {
      app {
        id
        resources(paths: $paths) {
          path
          data
        }
      }
    }
  }
`;

export type AppRawTemplatesUpdater<TemplatePath extends string> = (
  updateTemplates: {
    [path in TemplatePath]?: string | null;
  }
) => Promise<PortalAPIApp | null>;

export function useUpdateAppRawTemplatesMutation<TemplatePath extends string>(
  appID: string
): {
  updateAppRawTemplates: AppRawTemplatesUpdater<TemplatePath>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    UpdateAppRawTemplatesMutation,
    UpdateAppRawTemplatesMutationVariables
  >(updateAppRawTemplatesMutation, { client });
  const updateAppRawTemplates = React.useCallback(
    async (updateTemplates: { [path in TemplatePath]?: string | null }) => {
      const paths: string[] = [];
      const updates: AppResourceUpdate[] = [];
      for (const [path, data] of Object.entries(updateTemplates)) {
        if (data === undefined) {
          continue;
        }
        paths.push(path);
        if (typeof data === "string" && data.length > 0) {
          updates.push({ path, data });
        } else {
          updates.push({ path, data: null });
        }
      }

      const result = await mutationFunction({
        variables: {
          appID,
          paths,
          updates: updates.map((update) => {
            return {
              ...update,
              data: update.data == null ? null : btoa(update.data),
            };
          }),
        },
      });
      return result.data?.updateAppResources.app ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppRawTemplates, error, loading };
}
