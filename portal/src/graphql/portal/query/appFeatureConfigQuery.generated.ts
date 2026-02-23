import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AppFeatureConfigFragment = { __typename?: 'App', id: string, effectiveFeatureConfig: any, planName: string };

export type AppFeatureConfigQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
}>;


export type AppFeatureConfigQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, effectiveFeatureConfig: any, planName: string } | { __typename: 'User' } | { __typename: 'Viewer' } | null };

export const AppFeatureConfigFragmentDoc = gql`
    fragment AppFeatureConfig on App {
  id
  effectiveFeatureConfig
  planName
}
    `;
export const AppFeatureConfigQueryDocument = gql`
    query appFeatureConfigQuery($id: ID!) {
  node(id: $id) {
    __typename
    ...AppFeatureConfig
  }
}
    ${AppFeatureConfigFragmentDoc}`;

/**
 * __useAppFeatureConfigQueryQuery__
 *
 * To run a query within a React component, call `useAppFeatureConfigQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useAppFeatureConfigQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAppFeatureConfigQueryQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useAppFeatureConfigQueryQuery(baseOptions: Apollo.QueryHookOptions<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables> & ({ variables: AppFeatureConfigQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>(AppFeatureConfigQueryDocument, options);
      }
export function useAppFeatureConfigQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>(AppFeatureConfigQueryDocument, options);
        }
// @ts-ignore
export function useAppFeatureConfigQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>;
export function useAppFeatureConfigQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AppFeatureConfigQueryQuery | undefined, AppFeatureConfigQueryQueryVariables>;
export function useAppFeatureConfigQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>(AppFeatureConfigQueryDocument, options);
        }
export type AppFeatureConfigQueryQueryHookResult = ReturnType<typeof useAppFeatureConfigQueryQuery>;
export type AppFeatureConfigQueryLazyQueryHookResult = ReturnType<typeof useAppFeatureConfigQueryLazyQuery>;
export type AppFeatureConfigQuerySuspenseQueryHookResult = ReturnType<typeof useAppFeatureConfigQuerySuspenseQuery>;
export type AppFeatureConfigQueryQueryResult = Apollo.QueryResult<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>;