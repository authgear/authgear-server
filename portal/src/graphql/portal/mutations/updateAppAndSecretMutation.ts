import React from "react";

import { client } from "../../portal/apollo";
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
    withChecksum?: boolean
  ) => Promise<PortalAPIApp | null>;
  loading: boolean;
  error: unknown;
  resetError: () => void;
} {
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
      withChecksum: boolean = true
    ) => {
      const result = await mutationFunction({
        variables: {
          appID,
          appConfig: appConfig,
          appConfigChecksum: withChecksum ? appConfigChecksum : undefined,
          secretConfigUpdateInstructions: secretConfigUpdateInstructions,
          secretConfigUpdateInstructionsChecksum: withChecksum
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
