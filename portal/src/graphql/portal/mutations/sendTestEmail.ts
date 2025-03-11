import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import { usePortalClient } from "../../portal/apollo";
import {
  SendTestEmailMutationMutation,
  SendTestEmailMutationDocument,
  SendTestEmailMutationMutationVariables,
} from "./sendTestEmail.generated";

export interface SendTestEmailOptions {
  smtpHost: string;
  smtpPort: number;
  smtpUsername: string;
  smtpPassword: string;
  smtpSender: string;
  to: string;
}

export interface UseSendTestEmailMutationReturnType {
  sendTestEmail: (opts: SendTestEmailOptions) => Promise<void>;
  loading: boolean;
  error: unknown;
}

export function useSendTestEmailMutation(
  appID: string
): UseSendTestEmailMutationReturnType {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] = useMutation<
    SendTestEmailMutationMutation,
    SendTestEmailMutationMutationVariables
  >(SendTestEmailMutationDocument, {
    client,
  });
  const sendTestEmail = useCallback(
    async (options: SendTestEmailOptions) => {
      await mutationFunction({
        variables: {
          ...options,
          appID,
        },
      });
    },
    [mutationFunction, appID]
  );
  return { sendTestEmail, error, loading };
}
