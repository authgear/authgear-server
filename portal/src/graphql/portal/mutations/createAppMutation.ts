import React from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../../portal/apollo";
import {
  CreateAppMutationMutation,
  CreateAppMutationDocument,
} from "./createAppMutation.generated";
import { AppListQueryDocument } from "../query/appListQuery.generated";

export function useCreateAppMutation(): {
  createApp: (appID: string) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] =
    useMutation<CreateAppMutationMutation>(CreateAppMutationDocument, {
      client,
      refetchQueries: [{ query: AppListQueryDocument }],
    });
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
