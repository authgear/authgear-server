import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import { usePortalClient } from "../apollo";
import {
  SaveOnboardingSurveyMutationDocument,
  SaveOnboardingSurveyMutationMutation,
} from "./saveOnboardingSurveyMutation.generated";

export interface UseSaveOnboardingSurveyMutationReturnType {
  saveOnboardingSurvey: (surveyJson: string) => Promise<void>;
  loading: boolean;
  error: unknown;
  reset: () => void;
}

export function useSaveOnboardingSurveyMutation(): UseSaveOnboardingSurveyMutationReturnType {
  const client = usePortalClient();
  const [mutationFunction, { error, loading, reset }] =
    useMutation<SaveOnboardingSurveyMutationMutation>(
      SaveOnboardingSurveyMutationDocument,
      {
        client,
      }
    );
  const saveOnboardingSurvey = useCallback(
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

  return { saveOnboardingSurvey: saveOnboardingSurvey, error, loading, reset };
}
