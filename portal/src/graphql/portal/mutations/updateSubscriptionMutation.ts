import * as Apollo from "@apollo/client";
import {
  UpdateSubscriptionMutationDocument,
  UpdateSubscriptionMutationMutation,
  UpdateSubscriptionMutationMutationHookResult,
  UpdateSubscriptionMutationMutationVariables,
} from "./updateSubscriptionMutation.generated";
import { client } from "../apollo";

export function useUpdateSubscriptionMutation(
  baseOptions?: Apollo.MutationHookOptions<
    UpdateSubscriptionMutationMutation,
    UpdateSubscriptionMutationMutationVariables
  >
): UpdateSubscriptionMutationMutationHookResult {
  return Apollo.useMutation<
    UpdateSubscriptionMutationMutation,
    UpdateSubscriptionMutationMutationVariables
  >(UpdateSubscriptionMutationDocument, {
    ...{ client },
    ...baseOptions,
    refetchQueries: ["subscriptionScreenQuery"],
    awaitRefetchQueries: true,
  });
}
