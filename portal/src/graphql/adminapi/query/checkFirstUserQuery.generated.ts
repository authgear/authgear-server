import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CheckFirstUserQueryVariables = Types.Exact<{ [key: string]: never; }>;


export type CheckFirstUserQuery = { __typename?: 'Query', users?: { __typename?: 'UserConnection', totalCount?: number | null } | null };


export const CheckFirstUserDocument = gql`
    query CheckFirstUser {
  users(first: 1) {
    totalCount
  }
}
    `;

/**
 * __useCheckFirstUserQuery__
 *
 * To run a query within a React component, call `useCheckFirstUserQuery` and pass it any options that fit your needs.
 * When your component renders, `useCheckFirstUserQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useCheckFirstUserQuery({
 *   variables: {
 *   },
 * });
 */
export function useCheckFirstUserQuery(baseOptions?: Apollo.QueryHookOptions<CheckFirstUserQuery, CheckFirstUserQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<CheckFirstUserQuery, CheckFirstUserQueryVariables>(CheckFirstUserDocument, options);
      }
export function useCheckFirstUserLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<CheckFirstUserQuery, CheckFirstUserQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<CheckFirstUserQuery, CheckFirstUserQueryVariables>(CheckFirstUserDocument, options);
        }
export function useCheckFirstUserSuspenseQuery(baseOptions?: Apollo.SuspenseQueryHookOptions<CheckFirstUserQuery, CheckFirstUserQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<CheckFirstUserQuery, CheckFirstUserQueryVariables>(CheckFirstUserDocument, options);
        }
export type CheckFirstUserQueryHookResult = ReturnType<typeof useCheckFirstUserQuery>;
export type CheckFirstUserLazyQueryHookResult = ReturnType<typeof useCheckFirstUserLazyQuery>;
export type CheckFirstUserSuspenseQueryHookResult = ReturnType<typeof useCheckFirstUserSuspenseQuery>;
export type CheckFirstUserQueryResult = Apollo.QueryResult<CheckFirstUserQuery, CheckFirstUserQueryVariables>;