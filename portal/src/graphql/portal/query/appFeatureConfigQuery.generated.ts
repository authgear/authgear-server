import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AppFeatureConfigFragment = { __typename?: 'App', id: string, effectiveFeatureConfig: any, planName: string };

export type AppFeatureConfigQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID'];
}>;


export type AppFeatureConfigQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, effectiveFeatureConfig: any, planName: string } | { __typename: 'User' } | null };


export const AppFeatureConfigQueryDocument = gql`
    query appFeatureConfigQuery($id: ID!) {
  node(id: $id) {
    __typename
    ... on App {
      id
      effectiveFeatureConfig
      planName
    }
  }
}
    `;

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
export function useAppFeatureConfigQueryQuery(baseOptions: Apollo.QueryHookOptions<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>(AppFeatureConfigQueryDocument, options);
      }
export function useAppFeatureConfigQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>(AppFeatureConfigQueryDocument, options);
        }
export type AppFeatureConfigQueryQueryHookResult = ReturnType<typeof useAppFeatureConfigQueryQuery>;
export type AppFeatureConfigQueryLazyQueryHookResult = ReturnType<typeof useAppFeatureConfigQueryLazyQuery>;
export type AppFeatureConfigQueryQueryResult = Apollo.QueryResult<AppFeatureConfigQueryQuery, AppFeatureConfigQueryQueryVariables>;