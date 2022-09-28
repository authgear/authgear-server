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
    secretConfigUpdateInstructions?: PortalAPISecretConfigUpdateInstruction
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
      secretConfigUpdateInstructions?: PortalAPISecretConfigUpdateInstruction
    ) => {
      const result = await mutationFunction({
        variables: {
          appID,
          appConfig: appConfig,
          secretConfigUpdateInstructions: secretConfigUpdateInstructions,
        },
      });
      return result.data?.updateApp.app ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppAndSecretConfig, error, loading, resetError };
}
