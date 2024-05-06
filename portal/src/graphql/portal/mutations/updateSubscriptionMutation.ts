import * as Apollo from "@apollo/client";
import {
  UpdateSubscriptionMutationDocument,
  UpdateSubscriptionMutationMutation,
  UpdateSubscriptionMutationMutationHookResult,
  UpdateSubscriptionMutationMutationVariables,
} from "./updateSubscriptionMutation.generated";
import { usePortalClient } from "../apollo";

export function useUpdateSubscriptionMutation(
  baseOptions?: Apollo.MutationHookOptions<
    UpdateSubscriptionMutationMutation,
    UpdateSubscriptionMutationMutationVariables
  >
): UpdateSubscriptionMutationMutationHookResult {
  const client = usePortalClient();
  const options = {
    ...{ client },
    ...baseOptions,
    refetchQueries: ["subscriptionScreenQuery"],
  };
  return Apollo.useMutation<
    UpdateSubscriptionMutationMutation,
    UpdateSubscriptionMutationMutationVariables
  >(UpdateSubscriptionMutationDocument, options);
}
