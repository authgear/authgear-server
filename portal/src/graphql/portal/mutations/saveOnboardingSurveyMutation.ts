import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import { usePortalClient } from "../apollo";
import {
  SaveOnboardingSurveyMutationDocument,
  SaveOnboardingSurveyMutationMutation,
} from "./saveOnboardingSurveyMutation.generated";

export interface UseSaveOnboardingSurveyMutationReturnType {
  saveOnboardingSurveyHook: (surveyJson: string) => Promise<void>;
  loading: boolean;
  error: unknown;
  reset: () => void;
}

export function useSaveOnboardingSurveyMutation(): UseSaveOnboardingSurveyMutationReturnType {
  const client = usePortalClient();
  // eslint-disable-next-line @typescript-eslint/unbound-method
  const [mutationFunction, { error, loading, reset }] =
    useMutation<SaveOnboardingSurveyMutationMutation>(
      SaveOnboardingSurveyMutationDocument,
      {
        client,
      }
    );
  const saveOnboardingSurveyHook = useCallback(
    async (surveyJson: string) => {
      const result = await mutationFunction({
        variables: { surveyJSON: surveyJson },
      });
      if (result.errors != null && result.errors.length > 0) {
        // eslint-disable-next-line @typescript-eslint/only-throw-error
        throw result.errors;
      }
    },
    [mutationFunction]
  );

  return { saveOnboardingSurveyHook, error, loading, reset };
}
