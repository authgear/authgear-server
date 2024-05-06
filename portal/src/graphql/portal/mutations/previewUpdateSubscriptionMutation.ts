import * as Apollo from "@apollo/client";
import {
  PreviewUpdateSubscriptionMutationDocument,
  PreviewUpdateSubscriptionMutationMutation,
  PreviewUpdateSubscriptionMutationMutationHookResult,
  PreviewUpdateSubscriptionMutationMutationVariables,
} from "./previewUpdateSubscriptionMutation.generated";
import { usePortalClient } from "../apollo";

export function usePreviewUpdateSubscriptionMutation(
  baseOptions?: Apollo.MutationHookOptions<
    PreviewUpdateSubscriptionMutationMutation,
    PreviewUpdateSubscriptionMutationMutationVariables
  >
): PreviewUpdateSubscriptionMutationMutationHookResult {
  const client = usePortalClient();
  const options = {
    ...{ client },
    ...baseOptions,
  };
  return Apollo.useMutation<
    PreviewUpdateSubscriptionMutationMutation,
    PreviewUpdateSubscriptionMutationMutationVariables
  >(PreviewUpdateSubscriptionMutationDocument, options);
}
