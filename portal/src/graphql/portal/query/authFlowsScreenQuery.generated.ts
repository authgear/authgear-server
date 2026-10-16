import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AuthFlowsScreenQueryQueryVariables = Types.Exact<{
  id: Types.Scalars['ID']['input'];
}>;


export type AuthFlowsScreenQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, generatedAuthenticationFlows: { __typename?: 'GeneratedAuthenticationFlows', signupFlow: any, loginFlow: any, signupLoginFlow: any, reauthFlow: any, accountRecoveryFlow?: any | null } } | { __typename: 'User' } | { __typename: 'Viewer' } | null };


export const AuthFlowsScreenQueryDocument = gql`
    query authFlowsScreenQuery($id: ID!) {
  node(id: $id) {
    __typename
    ... on App {
      id
      generatedAuthenticationFlows {
        signupFlow
        loginFlow
        signupLoginFlow
        reauthFlow
        accountRecoveryFlow
      }
    }
  }
}
    `;

/**
 * __useAuthFlowsScreenQueryQuery__
 *
 * To run a query within a React component, call `useAuthFlowsScreenQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useAuthFlowsScreenQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAuthFlowsScreenQueryQuery({
 *   variables: {
 *      id: // value for 'id'
 *   },
 * });
 */
export function useAuthFlowsScreenQueryQuery(baseOptions: Apollo.QueryHookOptions<AuthFlowsScreenQueryQuery, AuthFlowsScreenQueryQueryVariables> & ({ variables: AuthFlowsScreenQueryQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AuthFlowsScreenQueryQuery, AuthFlowsScreenQueryQueryVariables>(AuthFlowsScreenQueryDocument, options);
      }
export function useAuthFlowsScreenQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AuthFlowsScreenQueryQuery, AuthFlowsScreenQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AuthFlowsScreenQueryQuery, AuthFlowsScreenQueryQueryVariables>(AuthFlowsScreenQueryDocument, options);
        }
// @ts-ignore
export function useAuthFlowsScreenQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<AuthFlowsScreenQueryQuery, AuthFlowsScreenQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AuthFlowsScreenQueryQuery, AuthFlowsScreenQueryQueryVariables>;
export function useAuthFlowsScreenQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AuthFlowsScreenQueryQuery, AuthFlowsScreenQueryQueryVariables>): Apollo.UseSuspenseQueryResult<AuthFlowsScreenQueryQuery | undefined, AuthFlowsScreenQueryQueryVariables>;
export function useAuthFlowsScreenQuerySuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<AuthFlowsScreenQueryQuery, AuthFlowsScreenQueryQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AuthFlowsScreenQueryQuery, AuthFlowsScreenQueryQueryVariables>(AuthFlowsScreenQueryDocument, options);
        }
export type AuthFlowsScreenQueryQueryHookResult = ReturnType<typeof useAuthFlowsScreenQueryQuery>;
export type AuthFlowsScreenQueryLazyQueryHookResult = ReturnType<typeof useAuthFlowsScreenQueryLazyQuery>;
export type AuthFlowsScreenQuerySuspenseQueryHookResult = ReturnType<typeof useAuthFlowsScreenQuerySuspenseQuery>;
export type AuthFlowsScreenQueryQueryResult = Apollo.QueryResult<AuthFlowsScreenQueryQuery, AuthFlowsScreenQueryQueryVariables>;