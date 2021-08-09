import { useCallback } from "react";
import { gql, useMutation } from "@apollo/client";
import { client } from "../../portal/apollo";
import { SendTestEmailMutation } from "./__generated__/SendTestEmailMutation";

const sendTestEmailMutation = gql`
  mutation SendTestEmailMutation(
    $appID: ID!
    $smtpHost: String!
    $smtpPort: Int!
    $smtpUsername: String!
    $smtpPassword: String!
    $to: String!
  ) {
    sendTestSMTPConfigurationEmail(
      input: {
        appID: $appID
        smtpHost: $smtpHost
        smtpPort: $smtpPort
        smtpUsername: $smtpUsername
        smtpPassword: $smtpPassword
        to: $to
      }
    )
  }
`;

export interface SendTestEmailOptions {
  smtpHost: string;
  smtpPort: number;
  smtpUsername: string;
  smtpPassword: string;
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
  const [mutationFunction, { error, loading }] =
    useMutation<SendTestEmailMutation>(sendTestEmailMutation, { client });
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
