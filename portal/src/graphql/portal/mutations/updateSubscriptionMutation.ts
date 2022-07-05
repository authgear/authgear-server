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
