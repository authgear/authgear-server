import React from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../../portal/apollo";
import {
  CreateAppMutationMutation,
  CreateAppMutationMutationVariables,
  CreateAppMutationDocument,
} from "./createAppMutation.generated";
import { AppListQueryDocument } from "../query/appListQuery.generated";

export function useCreateAppMutation(): {
  createApp: (
    appID: string,
    projectWizardData: unknown
  ) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] = useMutation<
    CreateAppMutationMutation,
    CreateAppMutationMutationVariables
  >(CreateAppMutationDocument, {
    client,
    refetchQueries: [{ query: AppListQueryDocument }],
  });
  const createApp = React.useCallback(
    async (appID: string, projectWizardData: unknown) => {
      const result = await mutationFunction({
        variables: { appID, projectWizardData },
      });
      return result.data?.createApp.app.id ?? null;
    },
    [mutationFunction]
  );
  return { createApp, error, loading };
}
