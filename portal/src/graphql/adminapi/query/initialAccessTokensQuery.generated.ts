import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type InitialAccessTokensQueryQueryVariables = Types.Exact<{ [key: string]: never; }>;


export type InitialAccessTokensQueryQuery = { __typename?: 'Query', initialAccessTokens: Array<{ __typename?: 'InitialAccessToken', id: string, createdAt: any, expiresAt: any, type: Types.InitialAccessTokenType }> };


export const InitialAccessTokensQueryDocument = gql`
    query initialAccessTokensQuery {
  initialAccessTokens {
    id
    createdAt
    expiresAt
    type
  }
}
    `;

/**
 * __useInitialAccessTokensQueryQuery__
 *
 * To run a query within a React component, call `useInitialAccessTokensQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useInitialAccessTokensQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useInitialAccessTokensQueryQuery({
 *   variables: {
 *   },
 * });
 */
export function useInitialAccessTokensQueryQuery(baseOptions?: Apollo.QueryHookOptions<InitialAccessTokensQueryQuery, InitialAccessTokensQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<InitialAccessTokensQueryQuery, InitialAccessTokensQueryQueryVariables>(InitialAccessTokensQueryDocument, options);
      }
export function useInitialAccessTokensQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<InitialAccessTokensQueryQuery, InitialAccessTokensQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<InitialAccessTokensQueryQuery, InitialAccessTokensQueryQueryVariables>(InitialAccessTokensQueryDocument, options);
        }
// @ts-ignore
export function useInitialAccessTokensQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<InitialAccessTokensQueryQuery, InitialAccessTokensQueryQueryVariables>): Apollo.UseSuspenseQueryResult<InitialAccessTokensQueryQuery, InitialAccessTokensQueryQueryVariables>;
export function useInitialAccessTokensQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<InitialAccessTokensQueryQuery, InitialAccessTokensQueryQueryVariables>): Apollo.UseSuspenseQueryResult<InitialAccessTokensQueryQuery | undefined, InitialAccessTokensQueryQueryVariables>;
export function useInitialAccessTokensQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<InitialAccessTokensQueryQuery, InitialAccessTokensQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<InitialAccessTokensQueryQuery, InitialAccessTokensQueryQueryVariables>(InitialAccessTokensQueryDocument, options);
        }
export type InitialAccessTokensQueryQueryHookResult = ReturnType<typeof useInitialAccessTokensQueryQuery>;
export type InitialAccessTokensQueryLazyQueryHookResult = ReturnType<typeof useInitialAccessTokensQueryLazyQuery>;
export type InitialAccessTokensQuerySuspenseQueryHookResult = ReturnType<typeof useInitialAccessTokensQuerySuspenseQuery>;
export type InitialAccessTokensQueryQueryResult = Apollo.QueryResult<InitialAccessTokensQueryQuery, InitialAccessTokensQueryQueryVariables>;