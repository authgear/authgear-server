import * as Apollo from "@apollo/client";
import {
  SubscriptionPlansQueryQuery,
  SubscriptionPlansQueryQueryVariables,
  SubscriptionPlansQueryDocument,
  SubscriptionPlansQueryQueryHookResult,
} from "./subscriptionPlansQuery.generated";
import { client } from "../apollo";

export function useSubscriptionPlansQueryQuery(
  baseOptions?: Apollo.QueryHookOptions<
    SubscriptionPlansQueryQuery,
    SubscriptionPlansQueryQueryVariables
  >
): SubscriptionPlansQueryQueryHookResult {
  const options = { ...{ client }, ...baseOptions };
  return Apollo.useQuery<
    SubscriptionPlansQueryQuery,
    SubscriptionPlansQueryQueryVariables
  >(SubscriptionPlansQueryDocument, options);
}
