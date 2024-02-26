import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AppTemplatesQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
  paths: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type AppTemplatesQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, resources: Array<{ __typename?: 'AppResource', path: string, languageTag?: string | null, data?: string | null, effectiveData?: string | null, checksum?: string | null }> } | { __typename: 'User' } | { __typename: 'Viewer' } | null };


export const AppTemplatesQueryDocument = gql`
    query appTemplatesQuery($id: ID!, $paths: [String!]!) {
  node(id: $id) {
    __typename
    ... on App {
      id
      resources(paths: $paths) {
        path
        languageTag
        data
        effectiveData
        checksum
      }
    }
  }
}
    `;

/**
 * __useAppTemplatesQueryQuery__
 *
 * To run a query within a React component, call `useAppTemplatesQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useAppTemplatesQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAppTemplatesQueryQuery({
 *   variables: {
 *      id: // value for 'id'
 *      paths: // value for 'paths'
 *   },
 * });
 */
export function useAppTemplatesQueryQuery(baseOptions: Apollo.QueryHookOptions<AppTemplatesQueryQuery, AppTemplatesQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AppTemplatesQueryQuery, AppTemplatesQueryQueryVariables>(AppTemplatesQueryDocument, options);
      }
export function useAppTemplatesQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AppTemplatesQueryQuery, AppTemplatesQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AppTemplatesQueryQuery, AppTemplatesQueryQueryVariables>(AppTemplatesQueryDocument, options);
        }
export function useAppTemplatesQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<AppTemplatesQueryQuery, AppTemplatesQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AppTemplatesQueryQuery, AppTemplatesQueryQueryVariables>(AppTemplatesQueryDocument, options);
        }
export type AppTemplatesQueryQueryHookResult = ReturnType<typeof useAppTemplatesQueryQuery>;
export type AppTemplatesQueryLazyQueryHookResult = ReturnType<typeof useAppTemplatesQueryLazyQuery>;
export type AppTemplatesQuerySuspenseQueryHookResult = ReturnType<typeof useAppTemplatesQuerySuspenseQuery>;
export type AppTemplatesQueryQueryResult = Apollo.QueryResult<AppTemplatesQueryQuery, AppTemplatesQueryQueryVariables>;