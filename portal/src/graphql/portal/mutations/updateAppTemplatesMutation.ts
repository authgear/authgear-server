import { useCallback } from "react";
import { useGraphqlMutation } from "../../../hook/graphql";
import { usePortalClient } from "../apollo";
import { AppResourceUpdate } from "../globalTypes.generated";
import {
  UpdateAppTemplatesMutationMutation,
  UpdateAppTemplatesMutationMutationVariables,
  UpdateAppTemplatesMutationDocument,
} from "./updateAppTemplatesMutation.generated";
import { PortalAPIApp } from "../../../types";
import {
  ResourceUpdate,
  binary,
  encodeForText,
  expandSpecifier,
} from "../../../util/resource";

export type AppTemplatesUpdater = (
  updates: ResourceUpdate[],
  ignoreConflict?: boolean
) => Promise<PortalAPIApp | null>;

export function useUpdateAppTemplatesMutation(appID: string): {
  updateAppTemplates: AppTemplatesUpdater;
  loading: boolean;
  error: unknown;
  resetError: () => void;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }, resetError] = useGraphqlMutation<
    UpdateAppTemplatesMutationMutation,
    UpdateAppTemplatesMutationMutationVariables
  >(UpdateAppTemplatesMutationDocument, {
    client,
  });
  const updateAppTemplates = useCallback(
    async (updates: ResourceUpdate[], ignoreConflict: boolean = false) => {
      const paths = [];
      for (const specifier of updates.map((u) => u.specifier)) {
        paths.push(expandSpecifier(specifier));
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
          checksum: !ignoreConflict ? update.checksum : undefined,
        };
      });

      const result = await mutationFunction({
        variables: {
          appID,
          paths,
          updates: updatePayload,
        },
      });
      return result.data?.updateApp.app ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppTemplates, error, loading, resetError };
}
