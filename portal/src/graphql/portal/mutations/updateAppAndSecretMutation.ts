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

export function useUpdateAppAndSecretConfigMutation(appID: string): {
  updateAppAndSecretConfig: (
    appConfig: PortalAPIAppConfig,
    appConfigChecksum?: string,
    secretConfigUpdateInstructions?: PortalAPISecretConfigUpdateInstruction,
    secretConfigUpdateInstructionsChecksum?: string,
    ignoreConflict?: boolean
  ) => Promise<PortalAPIApp | null>;
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
    async (
      appConfig: PortalAPIAppConfig,
      appConfigChecksum?: string,
      secretConfigUpdateInstructions?: PortalAPISecretConfigUpdateInstruction,
      secretConfigUpdateInstructionsChecksum?: string,
      ignoreConflict: boolean = false
    ) => {
      const result = await mutationFunction({
        variables: {
          appID,
          appConfig: appConfig,
          appConfigChecksum: !ignoreConflict ? appConfigChecksum : undefined,
          secretConfigUpdateInstructions: secretConfigUpdateInstructions,
          secretConfigUpdateInstructionsChecksum: !ignoreConflict
            ? secretConfigUpdateInstructionsChecksum
            : undefined,
        },
      });
      return result.data?.updateApp.app ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppAndSecretConfig, error, loading, resetError };
}
