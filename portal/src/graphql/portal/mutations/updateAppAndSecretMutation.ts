import React from "react";
import { useMutation, gql } from "@apollo/client";
import yaml from "js-yaml";

import { client } from "../../portal/apollo";
import {
  PortalAPIApp,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
} from "../../../types";
import { UpdateAppAndSecretConfigMutation } from "./__generated__/UpdateAppAndSecretConfigMutation";

// relative to project root
const APP_CONFIG_PATH = "./authgear.yaml";
const SECRET_CONFIG_PATH = "./authgear.secrets.yaml";

const updateAppAndSecretConfigMutation = gql`
  mutation UpdateAppAndSecretConfigMutation(
    $appID: String!
    $appConfigFile: AppConfigFile!
    $secretConfigFile: AppConfigFile!
  ) {
    updateAppConfig(
      input: {
        appID: $appID
        updateFiles: [$appConfigFile, $secretConfigFile]
        deleteFiles: []
      }
    ) {
      id
      rawAppConfig
      effectiveAppConfig
      rawSecretConfig
    }
  }
`;

export function useUpdateAppAndSecretConfigMutation(
  appID: string
): {
  updateAppAndSecretConfig: (
    appConfig: PortalAPIAppConfig,
    secretConfig: PortalAPISecretConfig
  ) => Promise<PortalAPIApp | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<
    UpdateAppAndSecretConfigMutation
  >(updateAppAndSecretConfigMutation, { client });
  const updateAppAndSecretConfig = React.useCallback(
    async (
      appConfig: PortalAPIAppConfig,
      secretConfig: PortalAPISecretConfig
    ) => {
      const appConfigYaml = yaml.safeDump(appConfig);
      const secretConfigYaml = yaml.safeDump(secretConfig);

      const result = await mutationFunction({
        variables: {
          appID,
          appConfigFile: { path: APP_CONFIG_PATH, content: appConfigYaml },
          secretConfigFile: {
            path: SECRET_CONFIG_PATH,
            content: secretConfigYaml,
          },
        },
      });
      return result.data?.updateAppConfig ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppAndSecretConfig, error, loading };
}
