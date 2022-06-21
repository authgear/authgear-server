import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SubscriptionPlansQueryQueryVariables = Types.Exact<{ [key: string]: never; }>;


export type SubscriptionPlansQueryQuery = { __typename?: 'Query', subscriptionPlans: Array<{ __typename?: 'SubscriptionPlan', name: string, prices: Array<{ __typename?: 'SubscriptionItemPrice', currency: string, smsRegion: Types.SubscriptionItemPriceSmsRegion, stripePriceID: string, stripeProductID: string, type: Types.SubscriptionItemPriceType, unitAmount: number, usageType: Types.SubscriptionItemPriceUsageType } | null> }> };


export const SubscriptionPlansQueryDocument = gql`
    query subscriptionPlansQuery {
  subscriptionPlans {
    name
    prices {
      currency
      smsRegion
      stripePriceID
      stripeProductID
      type
      unitAmount
      usageType
    }
  }
}
    `;

/**
 * __useSubscriptionPlansQueryQuery__
 *
 * To run a query within a React component, call `useSubscriptionPlansQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useSubscriptionPlansQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSubscriptionPlansQueryQuery({
 *   variables: {
 *   },
 * });
 */
export function useSubscriptionPlansQueryQuery(baseOptions?: Apollo.QueryHookOptions<SubscriptionPlansQueryQuery, SubscriptionPlansQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SubscriptionPlansQueryQuery, SubscriptionPlansQueryQueryVariables>(SubscriptionPlansQueryDocument, options);
      }
export function useSubscriptionPlansQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SubscriptionPlansQueryQuery, SubscriptionPlansQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SubscriptionPlansQueryQuery, SubscriptionPlansQueryQueryVariables>(SubscriptionPlansQueryDocument, options);
        }
export type SubscriptionPlansQueryQueryHookResult = ReturnType<typeof useSubscriptionPlansQueryQuery>;
export type SubscriptionPlansQueryLazyQueryHookResult = ReturnType<typeof useSubscriptionPlansQueryLazyQuery>;
export type SubscriptionPlansQueryQueryResult = Apollo.QueryResult<SubscriptionPlansQueryQuery, SubscriptionPlansQueryQueryVariables>;