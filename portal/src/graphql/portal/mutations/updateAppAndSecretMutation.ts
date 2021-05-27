import React from "react";
import { gql } from "@apollo/client";
import yaml from "js-yaml";

import { client } from "../../portal/apollo";
import {
  PortalAPIApp,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
} from "../../../types";
import {
  UpdateAppAndSecretConfigMutation,
  UpdateAppAndSecretConfigMutationVariables,
} from "./__generated__/UpdateAppAndSecretConfigMutation";
import { useGraphqlMutation } from "../../../hook/graphql";

const APP_CONFIG_PATH = "authgear.yaml";
const SECRET_CONFIG_PATH = "authgear.secrets.yaml";

const updateAppAndSecretConfigMutation = gql`
  mutation UpdateAppAndSecretConfigMutation(
    $appID: ID!
    $updates: [AppResourceUpdate!]!
  ) {
    updateAppResources(input: { appID: $appID, updates: $updates }) {
      app {
        id
        rawAppConfig
        effectiveAppConfig
        rawSecretConfig
      }
    }
  }
`;

export function useUpdateAppAndSecretConfigMutation(appID: string): {
  updateAppAndSecretConfig: (
    appConfig: PortalAPIAppConfig,
    secretConfig: PortalAPISecretConfig
  ) => Promise<PortalAPIApp | null>;
  loading: boolean;
  error: unknown;
  resetError: () => void;
} {
  const [mutationFunction, { error, loading }, resetError] = useGraphqlMutation<
    UpdateAppAndSecretConfigMutation,
    UpdateAppAndSecretConfigMutationVariables
  >(updateAppAndSecretConfigMutation, { client });
  const updateAppAndSecretConfig = React.useCallback(
    async (
      appConfig: PortalAPIAppConfig,
      secretConfig: PortalAPISecretConfig
    ) => {
      const appConfigYaml = yaml.dump(appConfig);
      const secretConfigYaml = yaml.dump(secretConfig);

      const result = await mutationFunction({
        variables: {
          appID,
          updates: [
            { path: APP_CONFIG_PATH, data: btoa(appConfigYaml) },
            { path: SECRET_CONFIG_PATH, data: btoa(secretConfigYaml) },
          ],
        },
      });
      return result.data?.updateAppResources.app ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppAndSecretConfig, error, loading, resetError };
}
