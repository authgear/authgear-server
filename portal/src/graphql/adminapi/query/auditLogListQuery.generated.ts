import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type AuditLogEdgesNodeFragment = { __typename?: 'AuditLog', id: string, createdAt: any, activityType: Types.AuditLogActivityType, data?: any | null, user?: { __typename?: 'User', id: string } | null };

export type AuditLogListFragment = { __typename?: 'AuditLogConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'AuditLogEdge', node?: { __typename?: 'AuditLog', id: string, createdAt: any, activityType: Types.AuditLogActivityType, data?: any | null, user?: { __typename?: 'User', id: string } | null } | null } | null> | null };

export type AuditLogListQueryQueryVariables = Types.Exact<{
  pageSize: Types.Scalars['Int']['input'];
  cursor?: Types.InputMaybe<Types.Scalars['String']['input']>;
  activityTypes?: Types.InputMaybe<Array<Types.AuditLogActivityType> | Types.AuditLogActivityType>;
  userIDs?: Types.InputMaybe<Array<Types.Scalars['ID']['input']> | Types.Scalars['ID']['input']>;
  emailAddresses?: Types.InputMaybe<Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input']>;
  phoneNumbers?: Types.InputMaybe<Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input']>;
  rangeFrom?: Types.InputMaybe<Types.Scalars['DateTime']['input']>;
  rangeTo?: Types.InputMaybe<Types.Scalars['DateTime']['input']>;
  sortDirection?: Types.InputMaybe<Types.SortDirection>;
}>;


export type AuditLogListQueryQuery = { __typename?: 'Query', auditLogs?: { __typename?: 'AuditLogConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'AuditLogEdge', node?: { __typename?: 'AuditLog', id: string, createdAt: any, activityType: Types.AuditLogActivityType, data?: any | null, user?: { __typename?: 'User', id: string } | null } | null } | null> | null } | null };

export const AuditLogEdgesNodeFragmentDoc = gql`
    fragment AuditLogEdgesNode on AuditLog {
  id
  createdAt
  activityType
  user {
    id
  }
  data
}
    `;
export const AuditLogListFragmentDoc = gql`
    fragment AuditLogList on AuditLogConnection {
  edges {
    node {
      ...AuditLogEdgesNode
    }
  }
  totalCount
}
    ${AuditLogEdgesNodeFragmentDoc}`;
export const AuditLogListQueryDocument = gql`
    query AuditLogListQuery($pageSize: Int!, $cursor: String, $activityTypes: [AuditLogActivityType!], $userIDs: [ID!], $emailAddresses: [String!], $phoneNumbers: [String!], $rangeFrom: DateTime, $rangeTo: DateTime, $sortDirection: SortDirection) {
  auditLogs(
    first: $pageSize
    after: $cursor
    activityTypes: $activityTypes
    userIDs: $userIDs
    emailAddresses: $emailAddresses
    phoneNumbers: $phoneNumbers
    rangeFrom: $rangeFrom
    rangeTo: $rangeTo
    sortDirection: $sortDirection
  ) {
    ...AuditLogList
  }
}
    ${AuditLogListFragmentDoc}`;

/**
 * __useAuditLogListQueryQuery__
 *
 * To run a query within a React component, call `useAuditLogListQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useAuditLogListQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useAuditLogListQueryQuery({
 *   variables: {
 *      pageSize: // value for 'pageSize'
 *      cursor: // value for 'cursor'
 *      activityTypes: // value for 'activityTypes'
 *      userIDs: // value for 'userIDs'
 *      emailAddresses: // value for 'emailAddresses'
 *      phoneNumbers: // value for 'phoneNumbers'
 *      rangeFrom: // value for 'rangeFrom'
 *      rangeTo: // value for 'rangeTo'
 *      sortDirection: // value for 'sortDirection'
 *   },
 * });
 */
export function useAuditLogListQueryQuery(baseOptions: Apollo.QueryHookOptions<AuditLogListQueryQuery, AuditLogListQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<AuditLogListQueryQuery, AuditLogListQueryQueryVariables>(AuditLogListQueryDocument, options);
      }
export function useAuditLogListQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<AuditLogListQueryQuery, AuditLogListQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<AuditLogListQueryQuery, AuditLogListQueryQueryVariables>(AuditLogListQueryDocument, options);
        }
export function useAuditLogListQuerySuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<AuditLogListQueryQuery, AuditLogListQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<AuditLogListQueryQuery, AuditLogListQueryQueryVariables>(AuditLogListQueryDocument, options);
        }
export type AuditLogListQueryQueryHookResult = ReturnType<typeof useAuditLogListQueryQuery>;
export type AuditLogListQueryLazyQueryHookResult = ReturnType<typeof useAuditLogListQueryLazyQuery>;
export type AuditLogListQuerySuspenseQueryHookResult = ReturnType<typeof useAuditLogListQuerySuspenseQuery>;
export type AuditLogListQueryQueryResult = Apollo.QueryResult<AuditLogListQueryQuery, AuditLogListQueryQueryVariables>;