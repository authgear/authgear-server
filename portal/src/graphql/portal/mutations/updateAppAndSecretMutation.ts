import React from "react";

import { usePortalClient } from "../../portal/apollo";
import {
  PortalAPIApp,
  PortalAPIAppConfig,
  PortalAPISecretConfigUpdateInstruction,
} from "../../../types";
import {
  UpdateAppAndSecretConfigMutationMutation,
  UpdateAppAndSecretConfigMutationMutationVariables,
  UpdateAppAndSecretConfigMutationDocument,
} from "./updateAppAndSecretMutation.generated";
import { useGraphqlMutation } from "../../../hook/graphql";
import { ResourceUpdate } from "../../../util/resource";
import { toAppResourceUpdates } from "./updateAppTemplatesMutation";

export function useUpdateAppAndSecretConfigMutation(appID: string): {
  updateAppAndSecretConfig: (options: {
    appConfig?: PortalAPIAppConfig;
    appConfigChecksum?: string;
    secretConfigUpdateInstructions?: PortalAPISecretConfigUpdateInstruction;
    secretConfigUpdateInstructionsChecksum?: string;
    // Resource files saved in the same request, so both land or neither.
    resourceUpdates?: ResourceUpdate[];
    ignoreConflict?: boolean;
  }) => Promise<PortalAPIApp | null>;
  loading: boolean;
  error: unknown;
  resetError: () => void;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }, resetError] = useGraphqlMutation<
    UpdateAppAndSecretConfigMutationMutation,
    UpdateAppAndSecretConfigMutationMutationVariables
  >(UpdateAppAndSecretConfigMutationDocument, { client });
  const updateAppAndSecretConfig = React.useCallback(
    async ({
      appConfig,
      appConfigChecksum,
      secretConfigUpdateInstructions,
      secretConfigUpdateInstructionsChecksum,
      resourceUpdates,
      ignoreConflict = false,
    }: {
      appConfig?: PortalAPIAppConfig;
      appConfigChecksum?: string;
      secretConfigUpdateInstructions?: PortalAPISecretConfigUpdateInstruction;
      secretConfigUpdateInstructionsChecksum?: string;
      resourceUpdates?: ResourceUpdate[];
      ignoreConflict?: boolean;
    }) => {
      const result = await mutationFunction({
        variables: {
          appID,
          appConfig: appConfig ? appConfig : undefined,
          appConfigChecksum: !ignoreConflict ? appConfigChecksum : undefined,
          secretConfigUpdateInstructions: secretConfigUpdateInstructions,
          secretConfigUpdateInstructionsChecksum: !ignoreConflict
            ? secretConfigUpdateInstructionsChecksum
            : undefined,
          updates:
            resourceUpdates != null
              ? toAppResourceUpdates(resourceUpdates, ignoreConflict)
              : undefined,
        },
        // The resource forms read their files from appTemplatesQuery.
        refetchQueries:
          resourceUpdates != null
            ? ["appAndSecretConfigQuery", "appTemplatesQuery"]
            : ["appAndSecretConfigQuery"],
        awaitRefetchQueries: resourceUpdates != null,
      });
      return result.data?.updateApp.app ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppAndSecretConfig, error, loading, resetError };
}
