import * as Apollo from "@apollo/client";
import {
  SubscriptionScreenQueryQuery,
  SubscriptionScreenQueryQueryVariables,
  SubscriptionScreenQueryDocument,
  SubscriptionScreenQueryQueryHookResult,
} from "./subscriptionScreenQuery.generated";
import { usePortalClient } from "../apollo";

export function useSubscriptionScreenQueryQuery(
  baseOptions?: Apollo.QueryHookOptions<
    SubscriptionScreenQueryQuery,
    SubscriptionScreenQueryQueryVariables
  >
): SubscriptionScreenQueryQueryHookResult {
  const client = usePortalClient();
  const options = { ...{ client }, ...baseOptions };
  return Apollo.useQuery<
    SubscriptionScreenQueryQuery,
    SubscriptionScreenQueryQueryVariables
  >(SubscriptionScreenQueryDocument, options);
}
