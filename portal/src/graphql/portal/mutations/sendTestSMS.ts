import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import { usePortalClient } from "../../portal/apollo";
import {
  SendTestSmsMutationMutation,
  SendTestSmsMutationDocument,
  SendTestSmsMutationMutationVariables,
} from "./sendTestSMS.generated";
import { SmsProviderConfigurationInput } from "../globalTypes.generated";

export interface SendTestSMSOptions {
  to: string;
  config: SmsProviderConfigurationInput;
}

export interface UseSendTestSMSMutationReturnType {
  sendTestSMS: (opts: SendTestSMSOptions) => Promise<void>;
  loading: boolean;
  error: unknown;
}

export function useSendTestSMSMutation(
  appID: string
): UseSendTestSMSMutationReturnType {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] = useMutation<
    SendTestSmsMutationMutation,
    SendTestSmsMutationMutationVariables
  >(SendTestSmsMutationDocument, {
    client,
  });
  const sendTestSMS = useCallback(
    async (options: SendTestSMSOptions) => {
      await mutationFunction({
        variables: {
          appID,
          to: options.to,
          config: options.config,
        },
      });
    },
    [mutationFunction, appID]
  );
  return { sendTestSMS, error, loading };
}
