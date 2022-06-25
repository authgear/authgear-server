import * as Apollo from "@apollo/client";
import {
  GenerateStripeCustomerPortalSessionMutationDocument,
  GenerateStripeCustomerPortalSessionMutationMutation,
  GenerateStripeCustomerPortalSessionMutationMutationHookResult,
  GenerateStripeCustomerPortalSessionMutationMutationVariables,
} from "./generateStripeCustomerPortalSessionMutation.generated";
import { client } from "../apollo";

export function useGenerateStripeCustomerPortalSessionMutationMutation(
  baseOptions?: Apollo.MutationHookOptions<
    GenerateStripeCustomerPortalSessionMutationMutation,
    GenerateStripeCustomerPortalSessionMutationMutationVariables
  >
): GenerateStripeCustomerPortalSessionMutationMutationHookResult {
  const options = { ...{ client }, ...baseOptions };
  return Apollo.useMutation<
    GenerateStripeCustomerPortalSessionMutationMutation,
    GenerateStripeCustomerPortalSessionMutationMutationVariables
  >(GenerateStripeCustomerPortalSessionMutationDocument, options);
}
