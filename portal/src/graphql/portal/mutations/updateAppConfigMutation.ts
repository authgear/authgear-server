import React from "react";
import { useMutation, gql } from "@apollo/client";

import { client } from "../../portal/apollo";
import { PortalAPIApp } from "../../../types";
import { UpdateAppConfigMutation } from "./__generated__/UpdateAppConfigMutation";

// relative to project root
const UPDATE_FILE_PATH = "./authgear.yaml";

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
  updateAppConfig: (appConfigYaml: string) => Promise<PortalAPIApp | null>;
} {
  const [mutationFunction] = useMutation<UpdateAppConfigMutation>(
    updateAppConfigMutation,
    { client }
  );
  const updateAppConfig = React.useCallback(
    async (appConfigYaml: string) => {
      const result = await mutationFunction({
        variables: {
          appID,
          updateFile: { path: UPDATE_FILE_PATH, content: appConfigYaml },
        },
      });
      return result.data?.updateAppConfig ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppConfig };
}
