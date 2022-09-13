import React from "react";

import { client } from "../../portal/apollo";
import {
  PortalAPIApp,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
} from "../../../types";
import {
  UpdateAppAndSecretConfigMutationMutation,
  UpdateAppAndSecretConfigMutationMutationVariables,
  UpdateAppAndSecretConfigMutationDocument,
} from "./updateAppAndSecretMutation.generated";
import { useGraphqlMutation } from "../../../hook/graphql";

// sanitizeSecretConfig makes sure the return value does not contain fields like __typename.
// The GraphQL runtime will complain about unknown fields.
function sanitizeSecretConfig(
  secretConfig: PortalAPISecretConfig
): PortalAPISecretConfig {
  return {
    oauthSSOProviderClientSecrets:
      secretConfig.oauthSSOProviderClientSecrets?.map((clientSecret) => {
        return {
          alias: clientSecret.alias,
          clientSecret: clientSecret.clientSecret,
        };
      }) ?? null,
    smtpSecret:
      secretConfig.smtpSecret != null
        ? {
            host: secretConfig.smtpSecret.host,
            port: secretConfig.smtpSecret.port,
            username: secretConfig.smtpSecret.username,
            password: secretConfig.smtpSecret.password,
          }
        : null,
  };
}

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
    UpdateAppAndSecretConfigMutationMutation,
    UpdateAppAndSecretConfigMutationMutationVariables
  >(UpdateAppAndSecretConfigMutationDocument, { client });
  const updateAppAndSecretConfig = React.useCallback(
    async (
      appConfig: PortalAPIAppConfig,
      secretConfig: PortalAPISecretConfig
    ) => {
      const result = await mutationFunction({
        variables: {
          appID,
          appConfig: appConfig,
          secretConfig: sanitizeSecretConfig(secretConfig),
        },
      });
      return result.data?.updateApp.app ?? null;
    },
    [appID, mutationFunction]
  );
  return { updateAppAndSecretConfig, error, loading, resetError };
}
