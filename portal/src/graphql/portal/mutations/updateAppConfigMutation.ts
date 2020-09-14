import React from "react";
import { useMutation, gql } from "@apollo/client";
import yaml from "js-yaml";

import { client } from "../../portal/apollo";
import { PortalAPIApp, PortalAPIAppConfig } from "../../../types";
import { UpdateAppConfigMutation } from "./__generated__/UpdateAppConfigMutation";

// relative to project root
const APP_CONFIG_PATH = "./authgear.yaml";

const updateAppConfigMutation = gql`
  mutation UpdateAppConfigMutation(
    $appID: String!
    $updateFile: AppConfigFile!
  ) {
    updateAppConfig(
      input: { appID: $appID, updateFiles: [$updateFile], deleteFiles: [] }
    ) {
      id
      rawAppConfig
      effectiveAppConfig
    }
  }
`;

export function useUpdateAppConfigMutation(
  appID: string
): {
  updateAppConfig: (
    appConfig: PortalAPIAppConfig
  ) => Promise<PortalAPIApp | null>;
} {
  const [mutationFunction] = useMutation<UpdateAppConfigMutation>(
    updateAppConfigMutation,
    { client }
  );
  const updateAppConfig = React.useCallback(
    async (appConfig: PortalAPIAppConfig) => {
      const appConfigYaml = yaml.safeDump(appConfig);

      const result = await mutationFunction({
        variables: {
          appID,
          updateFile: { path: APP_CONFIG_PATH, content: appConfigYaml },
        },
      });
      return result.data?.updateAppConfig ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppConfig };
}
