import React from "react";
import { gql, useMutation } from "@apollo/client";

import { client } from "../../portal/apollo";
import { CreateAppMutation } from "./__generated__/CreateAppMutation";
import { appListQuery } from "../query/appListQuery";

const createAppMutation = gql`
  mutation CreateAppMutation($appID: String!) {
    createApp(input: { id: $appID }) {
      app {
        id
      }
    }
  }
`;

export function useCreateAppMutation(): {
  createApp: (appID: string) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const [mutationFunction, { error, loading }] = useMutation<CreateAppMutation>(
    createAppMutation,
    {
      client,
      refetchQueries: [{ query: appListQuery }],
    }
  );
  const createApp = React.useCallback(
    async (appID: string) => {
      const result = await mutationFunction({
        variables: { appID },
      });
      return result.data?.createApp.app.id ?? null;
    },
    [mutationFunction]
  );
  return { createApp, error, loading };
}
