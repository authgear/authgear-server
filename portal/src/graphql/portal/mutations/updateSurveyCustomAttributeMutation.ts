import { useCallback } from "react";
import { useMutation } from "@apollo/client";
import { usePortalClient } from "../../portal/apollo";
import {
  UpdateSurveyCustomAttributeMutationDocument,
  UpdateSurveyCustomAttributeMutationMutation,
} from "./updateSurveyCustomAttributeMutation.generated";

export interface UseUpdateSurveyCustomAttributeMutationReturnType {
  updateSurveyCustomAttributeHook: (surveyJson: string) => Promise<void>;
  loading: boolean;
  error: unknown;
  reset: () => void;
}

export function useUpdateSurveyCustomAttributeMutation(): UseUpdateSurveyCustomAttributeMutationReturnType {
  const client = usePortalClient();
  const [mutationFunction, { error, loading, reset }] =
    useMutation<UpdateSurveyCustomAttributeMutationMutation>(
      UpdateSurveyCustomAttributeMutationDocument,
      {
        client,
      }
    );
  const updateSurveyCustomAttributeHook = useCallback(
    async (surveyJson: string) => {
      await mutationFunction({
        variables: { surveyJSON: surveyJson },
      });
    },
    [mutationFunction]
  );

  return { updateSurveyCustomAttributeHook, error, loading, reset };
}
