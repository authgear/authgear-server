import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CollaboratorsAndInvitationsQueryQueryVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
}>;


export type CollaboratorsAndInvitationsQueryQuery = { __typename?: 'Query', node?: { __typename: 'App', id: string, collaborators: Array<{ __typename?: 'Collaborator', id: string, role: Types.CollaboratorRole, createdAt: any, user: { __typename?: 'User', id: string, email?: string | null } }>, collaboratorInvitations: Array<{ __typename?: 'CollaboratorInvitation', id: string, createdAt: any, expireAt: any, inviteeEmail: string, invitedBy: { __typename?: 'User', id: string, email?: string | null } }> } | { __typename: 'User' } | { __typename: 'Viewer' } | null };


export const CollaboratorsAndInvitationsQueryDocument = gql`
    query collaboratorsAndInvitationsQuery($appID: ID!) {
  node(id: $appID) {
    __typename
    ... on App {
      id
      collaborators {
        id
        role
        createdAt
        user {
          id
          email
        }
      }
      collaboratorInvitations {
        id
        createdAt
        expireAt
        invitedBy {
          id
          email
        }
        inviteeEmail
      }
    }
  }
}
    `;

/**
 * __useCollaboratorsAndInvitationsQueryQuery__
 *
 * To run a query within a React component, call `useCollaboratorsAndInvitationsQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useCollaboratorsAndInvitationsQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useCollaboratorsAndInvitationsQueryQuery({
 *   variables: {
 *      appID: // value for 'appID'
 *   },
 * });
 */
export function useCollaboratorsAndInvitationsQueryQuery(baseOptions: Apollo.QueryHookOptions<CollaboratorsAndInvitationsQueryQuery, CollaboratorsAndInvitationsQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<CollaboratorsAndInvitationsQueryQuery, CollaboratorsAndInvitationsQueryQueryVariables>(CollaboratorsAndInvitationsQueryDocument, options);
      }
export function useCollaboratorsAndInvitationsQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<CollaboratorsAndInvitationsQueryQuery, CollaboratorsAndInvitationsQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<CollaboratorsAndInvitationsQueryQuery, CollaboratorsAndInvitationsQueryQueryVariables>(CollaboratorsAndInvitationsQueryDocument, options);
        }
export function useCollaboratorsAndInvitationsQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<CollaboratorsAndInvitationsQueryQuery, CollaboratorsAndInvitationsQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<CollaboratorsAndInvitationsQueryQuery, CollaboratorsAndInvitationsQueryQueryVariables>(CollaboratorsAndInvitationsQueryDocument, options);
        }
export type CollaboratorsAndInvitationsQueryQueryHookResult = ReturnType<typeof useCollaboratorsAndInvitationsQueryQuery>;
export type CollaboratorsAndInvitationsQueryLazyQueryHookResult = ReturnType<typeof useCollaboratorsAndInvitationsQueryLazyQuery>;
export type CollaboratorsAndInvitationsQuerySuspenseQueryHookResult = ReturnType<typeof useCollaboratorsAndInvitationsQuerySuspenseQuery>;
export type CollaboratorsAndInvitationsQueryQueryResult = Apollo.QueryResult<CollaboratorsAndInvitationsQueryQuery, CollaboratorsAndInvitationsQueryQueryVariables>;