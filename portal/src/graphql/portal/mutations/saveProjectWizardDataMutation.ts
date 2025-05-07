import React from "react";
import { useMutation } from "@apollo/client";

import { usePortalClient } from "../../portal/apollo";
import {
  SaveProjectWizardDataMutationDocument,
  SaveProjectWizardDataMutationMutationVariables,
  SaveProjectWizardDataMutationMutation,
} from "./saveProjectWizardDataMutation.generated";

export function useSaveProjectWizardDataMutation(): {
  mutate: (appID: string, projectWizardData: unknown) => Promise<string | null>;
  loading: boolean;
  error: unknown;
} {
  const client = usePortalClient();
  const [mutationFunction, { error, loading }] = useMutation<
    SaveProjectWizardDataMutationMutation,
    SaveProjectWizardDataMutationMutationVariables
  >(SaveProjectWizardDataMutationDocument, {
    client,
  });
  const mutate = React.useCallback(
    async (appID: string, projectWizardData: unknown) => {
      const result = await mutationFunction({
        variables: { appID, data: projectWizardData },
      });
      return result.data?.saveProjectWizardData.app.id ?? null;
    },
    [mutationFunction]
  );
  return { mutate, error, loading };
}
