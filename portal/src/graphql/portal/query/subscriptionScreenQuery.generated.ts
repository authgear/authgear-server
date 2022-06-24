import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AppFragmentFragment = { __typename?: 'App', id: string, effectiveFeatureConfig: any, planName: string, previousMonth?: { __typename?: 'SubscriptionUsage', nextBillingDate: any, items: Array<{ __typename?: 'SubscriptionUsageItem', type: Types.SubscriptionItemPriceType, usageType: Types.SubscriptionItemPriceUsageType, smsRegion: Types.SubscriptionItemPriceSmsRegion, quantity: number, currency?: string | null, unitAmount?: number | null, totalAmount?: number | null }> } | null, thisMonth?: { __typename?: 'SubscriptionUsage', nextBillingDate: any, items: Array<{ __typename?: 'SubscriptionUsageItem', type: Types.SubscriptionItemPriceType, usageType: Types.SubscriptionItemPriceUsageType, smsRegion: Types.SubscriptionItemPriceSmsRegion, quantity: number, currency?: string | null, unitAmount?: number | null, totalAmount?: number | null }> } | null };

export type SubscriptionScreenQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID'];
  thisMonth: Types.Scalars['DateTime'];
  previousMonth: Types.Scalars['DateTime'];
}>;


export type SubscriptionScreenQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, effectiveFeatureConfig: any, planName: string, previousMonth?: { __typename?: 'SubscriptionUsage', nextBillingDate: any, items: Array<{ __typename?: 'SubscriptionUsageItem', type: Types.SubscriptionItemPriceType, usageType: Types.SubscriptionItemPriceUsageType, smsRegion: Types.SubscriptionItemPriceSmsRegion, quantity: number, currency?: string | null, unitAmount?: number | null, totalAmount?: number | null }> } | null, thisMonth?: { __typename?: 'SubscriptionUsage', nextBillingDate: any, items: Array<{ __typename?: 'SubscriptionUsageItem', type: Types.SubscriptionItemPriceType, usageType: Types.SubscriptionItemPriceUsageType, smsRegion: Types.SubscriptionItemPriceSmsRegion, quantity: number, currency?: string | null, unitAmount?: number | null, totalAmount?: number | null }> } | null } | { __typename: 'User' } | null, subscriptionPlans: Array<{ __typename?: 'SubscriptionPlan', name: string, prices: Array<{ __typename?: 'SubscriptionItemPrice', currency: string, smsRegion: Types.SubscriptionItemPriceSmsRegion, type: Types.SubscriptionItemPriceType, unitAmount: number, usageType: Types.SubscriptionItemPriceUsageType }> }> };

export const AppFragmentFragmentDoc = gql`
    fragment AppFragment on App {
  id
  effectiveFeatureConfig
  planName
  previousMonth: subscriptionUsage(date: $previousMonth) {
    nextBillingDate
    items {
      type
      usageType
      smsRegion
      quantity
      currency
      unitAmount
      totalAmount
    }
  }
  thisMonth: subscriptionUsage(date: $thisMonth) {
    nextBillingDate
    items {
      type
      usageType
      smsRegion
      quantity
      currency
      unitAmount
      totalAmount
    }
  }
}
    `;
export const SubscriptionScreenQueryDocument = gql`
    query subscriptionScreenQuery($id: ID!, $thisMonth: DateTime!, $previousMonth: DateTime!) {
  node(id: $id) {
    __typename
    ...AppFragment
  }
  subscriptionPlans {
    name
    prices {
      currency
      smsRegion
      type
      unitAmount
      usageType
    }
  }
}
    ${AppFragmentFragmentDoc}`;

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
 *      thisMonth: // value for 'thisMonth'
 *      previousMonth: // value for 'previousMonth'
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