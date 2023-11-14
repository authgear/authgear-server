import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AuthenticatedForInvitationQueryQueryVariables = Types.Exact<{
  code: Types.Scalars['String']['input'];
}>;


export type AuthenticatedForInvitationQueryQuery = { __typename?: 'Query', viewer?: { __typename?: 'Viewer', email?: string | null } | null, checkCollaboratorInvitation?: { __typename?: 'CheckCollaboratorInvitationPayload', isInvitee: boolean, appID: string } | null };


export const AuthenticatedForInvitationQueryDocument = gql`
    query authenticatedForInvitationQuery($code: String!) {
  viewer {
    email
  }
  checkCollaboratorInvitation(code: $code) {
    isInvitee
    appID
  }
}
    `;

/**
 * __useAuthenticatedForInvitationQueryQuery__
 *
 * To run a query within a React component, call `useAuthenticatedForInvitationQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useAuthenticatedForInvitationQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAuthenticatedForInvitationQueryQuery({
 *   variables: {
 *      code: // value for 'code'
 *   },
 * });
 */
export function useAuthenticatedForInvitationQueryQuery(baseOptions: Apollo.QueryHookOptions<AuthenticatedForInvitationQueryQuery, AuthenticatedForInvitationQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AuthenticatedForInvitationQueryQuery, AuthenticatedForInvitationQueryQueryVariables>(AuthenticatedForInvitationQueryDocument, options);
      }
export function useAuthenticatedForInvitationQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AuthenticatedForInvitationQueryQuery, AuthenticatedForInvitationQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AuthenticatedForInvitationQueryQuery, AuthenticatedForInvitationQueryQueryVariables>(AuthenticatedForInvitationQueryDocument, options);
        }
export function useAuthenticatedForInvitationQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<AuthenticatedForInvitationQueryQuery, AuthenticatedForInvitationQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AuthenticatedForInvitationQueryQuery, AuthenticatedForInvitationQueryQueryVariables>(AuthenticatedForInvitationQueryDocument, options);
        }
export type AuthenticatedForInvitationQueryQueryHookResult = ReturnType<typeof useAuthenticatedForInvitationQueryQuery>;
export type AuthenticatedForInvitationQueryLazyQueryHookResult = ReturnType<typeof useAuthenticatedForInvitationQueryLazyQuery>;
export type AuthenticatedForInvitationQuerySuspenseQueryHookResult = ReturnType<typeof useAuthenticatedForInvitationQuerySuspenseQuery>;
export type AuthenticatedForInvitationQueryQueryResult = Apollo.QueryResult<AuthenticatedForInvitationQueryQuery, AuthenticatedForInvitationQueryQueryVariables>;