import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import { AppFeatureConfigFragmentDoc } from './appFeatureConfigQuery.generated';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SubscriptionScreenQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID'];
}>;


export type SubscriptionScreenQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, effectiveFeatureConfig: any, planName: string } | { __typename: 'User' } | null, subscriptionPlans: Array<{ __typename?: 'SubscriptionPlan', name: string, prices: Array<{ __typename?: 'SubscriptionItemPrice', currency: string, smsRegion: Types.SubscriptionItemPriceSmsRegion, stripePriceID: string, stripeProductID: string, type: Types.SubscriptionItemPriceType, unitAmount: number, usageType: Types.SubscriptionItemPriceUsageType } | null> }> };


export const SubscriptionScreenQueryDocument = gql`
    query subscriptionScreenQuery($id: ID!) {
  node(id: $id) {
    __typename
    ...AppFeatureConfig
  }
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
    ${AppFeatureConfigFragmentDoc}`;

/**
 * __useSubscriptionScreenQueryQuery__
 *
 * To run a query within a React component, call `useSubscriptionScreenQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useSubscriptionScreenQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSubscriptionScreenQueryQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useSubscriptionScreenQueryQuery(baseOptions: Apollo.QueryHookOptions<SubscriptionScreenQueryQuery, SubscriptionScreenQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<SubscriptionScreenQueryQuery, SubscriptionScreenQueryQueryVariables>(SubscriptionScreenQueryDocument, options);
      }
export function useSubscriptionScreenQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<SubscriptionScreenQueryQuery, SubscriptionScreenQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<SubscriptionScreenQueryQuery, SubscriptionScreenQueryQueryVariables>(SubscriptionScreenQueryDocument, options);
        }
export type SubscriptionScreenQueryQueryHookResult = ReturnType<typeof useSubscriptionScreenQueryQuery>;
export type SubscriptionScreenQueryLazyQueryHookResult = ReturnType<typeof useSubscriptionScreenQueryLazyQuery>;
export type SubscriptionScreenQueryQueryResult = Apollo.QueryResult<SubscriptionScreenQueryQuery, SubscriptionScreenQueryQueryVariables>;