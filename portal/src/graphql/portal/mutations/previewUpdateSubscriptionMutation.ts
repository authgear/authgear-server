import * as Apollo from "@apollo/client";
import {
  PreviewUpdateSubscriptionMutationDocument,
  PreviewUpdateSubscriptionMutationMutation,
  PreviewUpdateSubscriptionMutationMutationHookResult,
  PreviewUpdateSubscriptionMutationMutationVariables,
} from "./previewUpdateSubscriptionMutation.generated";
import { client } from "../apollo";

export function usePreviewUpdateSubscriptionMutation(
  baseOptions?: Apollo.MutationHookOptions<
    PreviewUpdateSubscriptionMutationMutation,
    PreviewUpdateSubscriptionMutationMutationVariables
  >
): PreviewUpdateSubscriptionMutationMutationHookResult {
  const options = {
    ...{ client },
    ...baseOptions,
  };
  return Apollo.useMutation<
    PreviewUpdateSubscriptionMutationMutation,
    PreviewUpdateSubscriptionMutationMutationVariables
  >(PreviewUpdateSubscriptionMutationDocument, options);
}
