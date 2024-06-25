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
  const [mutationFunction, { error, loading, reset }] =
    useMutation<SaveOnboardingSurveyMutationMutation>(
      SaveOnboardingSurveyMutationDocument,
      {
        client,
      }
    );
  const saveOnboardingSurveyHook = useCallback(
    async (surveyJson: string) => {
      await mutationFunction({
        variables: { surveyJSON: surveyJson },
      });
    },
    [mutationFunction]
  );

  return { saveOnboardingSurveyHook, error, loading, reset };
}
