import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type ScreenNavFragment = { __typename?: 'App', id: string, effectiveFeatureConfig: any, planName: string, tutorialStatus: { __typename?: 'TutorialStatus', appID: string, data: any } };

export type ScreenNavQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
}>;


export type ScreenNavQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, effectiveFeatureConfig: any, planName: string, tutorialStatus: { __typename?: 'TutorialStatus', appID: string, data: any } } | { __typename: 'User' } | { __typename: 'Viewer' } | null };

export const ScreenNavFragmentDoc = gql`
    fragment ScreenNav on App {
  id
  effectiveFeatureConfig
  planName
  tutorialStatus {
    appID
    data
  }
}
    `;
export const ScreenNavQueryDocument = gql`
    query screenNavQuery($id: ID!) {
  node(id: $id) {
    __typename
    ...ScreenNav
  }
}
    ${ScreenNavFragmentDoc}`;

/**
 * __useScreenNavQueryQuery__
 *
 * To run a query within a React component, call `useScreenNavQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useScreenNavQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useScreenNavQueryQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useScreenNavQueryQuery(baseOptions: Apollo.QueryHookOptions<ScreenNavQueryQuery, ScreenNavQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ScreenNavQueryQuery, ScreenNavQueryQueryVariables>(ScreenNavQueryDocument, options);
      }
export function useScreenNavQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ScreenNavQueryQuery, ScreenNavQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ScreenNavQueryQuery, ScreenNavQueryQueryVariables>(ScreenNavQueryDocument, options);
        }
export function useScreenNavQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<ScreenNavQueryQuery, ScreenNavQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<ScreenNavQueryQuery, ScreenNavQueryQueryVariables>(ScreenNavQueryDocument, options);
        }
export type ScreenNavQueryQueryHookResult = ReturnType<typeof useScreenNavQueryQuery>;
export type ScreenNavQueryLazyQueryHookResult = ReturnType<typeof useScreenNavQueryLazyQuery>;
export type ScreenNavQuerySuspenseQueryHookResult = ReturnType<typeof useScreenNavQuerySuspenseQuery>;
export type ScreenNavQueryQueryResult = Apollo.QueryResult<ScreenNavQueryQuery, ScreenNavQueryQueryVariables>;