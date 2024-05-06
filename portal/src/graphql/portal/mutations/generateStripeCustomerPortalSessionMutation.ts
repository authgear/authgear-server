import * as Apollo from "@apollo/client";
import {
  GenerateStripeCustomerPortalSessionMutationDocument,
  GenerateStripeCustomerPortalSessionMutationMutation,
  GenerateStripeCustomerPortalSessionMutationMutationHookResult,
  GenerateStripeCustomerPortalSessionMutationMutationVariables,
} from "./generateStripeCustomerPortalSessionMutation.generated";
import { usePortalClient } from "../apollo";

export function useGenerateStripeCustomerPortalSessionMutationMutation(
  baseOptions?: Apollo.MutationHookOptions<
    GenerateStripeCustomerPortalSessionMutationMutation,
    GenerateStripeCustomerPortalSessionMutationMutationVariables
  >
): GenerateStripeCustomerPortalSessionMutationMutationHookResult {
  const client = usePortalClient();
  const options = { ...{ client }, ...baseOptions };
  return Apollo.useMutation<
    GenerateStripeCustomerPortalSessionMutationMutation,
    GenerateStripeCustomerPortalSessionMutationMutationVariables
  >(GenerateStripeCustomerPortalSessionMutationDocument, options);
}
