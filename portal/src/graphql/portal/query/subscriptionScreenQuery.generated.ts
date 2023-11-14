import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AppFragmentFragment = { __typename?: 'App', id: string, effectiveAppConfig: any, effectiveFeatureConfig: any, isProcessingSubscription: boolean, lastStripeError?: any | null, planName: string, previousMonth?: { __typename?: 'SubscriptionUsage', nextBillingDate: any, items: Array<{ __typename?: 'SubscriptionUsageItem', type: Types.SubscriptionItemPriceType, usageType: Types.SubscriptionItemPriceUsageType, smsRegion: Types.SubscriptionItemPriceSmsRegion, whatsappRegion: Types.SubscriptionItemPriceWhatsappRegion, quantity: number, currency?: string | null, unitAmount?: number | null, totalAmount?: number | null, freeQuantity?: number | null, transformQuantityDivideBy?: number | null, transformQuantityRound: Types.TransformQuantityRound }> } | null, thisMonth?: { __typename?: 'SubscriptionUsage', nextBillingDate: any, items: Array<{ __typename?: 'SubscriptionUsageItem', type: Types.SubscriptionItemPriceType, usageType: Types.SubscriptionItemPriceUsageType, smsRegion: Types.SubscriptionItemPriceSmsRegion, whatsappRegion: Types.SubscriptionItemPriceWhatsappRegion, quantity: number, currency?: string | null, unitAmount?: number | null, totalAmount?: number | null, freeQuantity?: number | null, transformQuantityDivideBy?: number | null, transformQuantityRound: Types.TransformQuantityRound }> } | null, subscription?: { __typename?: 'Subscription', id: string, createdAt: any, updatedAt: any, cancelledAt?: any | null, endedAt?: any | null } | null };

export type SubscriptionScreenQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
  thisMonth: Types.Scalars['DateTime']['input'];
  previousMonth: Types.Scalars['DateTime']['input'];
}>;


export type SubscriptionScreenQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, effectiveAppConfig: any, effectiveFeatureConfig: any, isProcessingSubscription: boolean, lastStripeError?: any | null, planName: string, previousMonth?: { __typename?: 'SubscriptionUsage', nextBillingDate: any, items: Array<{ __typename?: 'SubscriptionUsageItem', type: Types.SubscriptionItemPriceType, usageType: Types.SubscriptionItemPriceUsageType, smsRegion: Types.SubscriptionItemPriceSmsRegion, whatsappRegion: Types.SubscriptionItemPriceWhatsappRegion, quantity: number, currency?: string | null, unitAmount?: number | null, totalAmount?: number | null, freeQuantity?: number | null, transformQuantityDivideBy?: number | null, transformQuantityRound: Types.TransformQuantityRound }> } | null, thisMonth?: { __typename?: 'SubscriptionUsage', nextBillingDate: any, items: Array<{ __typename?: 'SubscriptionUsageItem', type: Types.SubscriptionItemPriceType, usageType: Types.SubscriptionItemPriceUsageType, smsRegion: Types.SubscriptionItemPriceSmsRegion, whatsappRegion: Types.SubscriptionItemPriceWhatsappRegion, quantity: number, currency?: string | null, unitAmount?: number | null, totalAmount?: number | null, freeQuantity?: number | null, transformQuantityDivideBy?: number | null, transformQuantityRound: Types.TransformQuantityRound }> } | null, subscription?: { __typename?: 'Subscription', id: string, createdAt: any, updatedAt: any, cancelledAt?: any | null, endedAt?: any | null } | null } | { __typename: 'User' } | { __typename: 'Viewer' } | null, subscriptionPlans: Array<{ __typename?: 'SubscriptionPlan', name: string, prices: Array<{ __typename?: 'SubscriptionItemPrice', currency: string, smsRegion: Types.SubscriptionItemPriceSmsRegion, whatsappRegion: Types.SubscriptionItemPriceWhatsappRegion, type: Types.SubscriptionItemPriceType, unitAmount: number, usageType: Types.SubscriptionItemPriceUsageType, freeQuantity?: number | null, transformQuantityDivideBy?: number | null, transformQuantityRound: Types.TransformQuantityRound }> }> };

export const AppFragmentFragmentDoc = gql`
    fragment AppFragment on App {
  id
  effectiveAppConfig
  effectiveFeatureConfig
  isProcessingSubscription
  lastStripeError
  planName
  previousMonth: subscriptionUsage(date: $previousMonth) {
    nextBillingDate
    items {
      type
      usageType
      smsRegion
      whatsappRegion
      quantity
      currency
      unitAmount
      totalAmount
      freeQuantity
      transformQuantityDivideBy
      transformQuantityRound
    }
  }
  thisMonth: subscriptionUsage(date: $thisMonth) {
    nextBillingDate
    items {
      type
      usageType
      smsRegion
      whatsappRegion
      quantity
      currency
      unitAmount
      totalAmount
      freeQuantity
      transformQuantityDivideBy
      transformQuantityRound
    }
  }
  subscription {
    id
    createdAt
    updatedAt
    cancelledAt
    endedAt
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
      whatsappRegion
      type
      unitAmount
      usageType
      freeQuantity
      transformQuantityDivideBy
      transformQuantityRound
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
export function useSubscriptionScreenQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<SubscriptionScreenQueryQuery, SubscriptionScreenQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<SubscriptionScreenQueryQuery, SubscriptionScreenQueryQueryVariables>(SubscriptionScreenQueryDocument, options);
        }
export type SubscriptionScreenQueryQueryHookResult = ReturnType<typeof useSubscriptionScreenQueryQuery>;
export type SubscriptionScreenQueryLazyQueryHookResult = ReturnType<typeof useSubscriptionScreenQueryLazyQuery>;
export type SubscriptionScreenQuerySuspenseQueryHookResult = ReturnType<typeof useSubscriptionScreenQuerySuspenseQuery>;
export type SubscriptionScreenQueryQueryResult = Apollo.QueryResult<SubscriptionScreenQueryQuery, SubscriptionScreenQueryQueryVariables>;