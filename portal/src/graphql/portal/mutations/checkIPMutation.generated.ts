import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CheckIpMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  ipAddress: Types.Scalars['String']['input'];
  cidrs: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
  countryCodes: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type CheckIpMutationMutation = { __typename?: 'Mutation', checkIP?: boolean | null };


export const CheckIpMutationDocument = gql`
    mutation CheckIPMutation($appID: ID!, $ipAddress: String!, $cidrs: [String!]!, $countryCodes: [String!]!) {
  checkIP(
    input: {appID: $appID, ipAddress: $ipAddress, cidrs: $cidrs, countryCodes: $countryCodes}
  )
}
    `;
export type CheckIpMutationMutationFn = Apollo.MutationFunction<CheckIpMutationMutation, CheckIpMutationMutationVariables>;

/**
 * __useCheckIpMutationMutation__
 *
 * To run a mutation, you first call `useCheckIpMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCheckIpMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [checkIpMutationMutation, { data, loading, error }] = useCheckIpMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      ipAddress: // value for 'ipAddress'
 *      cidrs: // value for 'cidrs'
 *      countryCodes: // value for 'countryCodes'
 *   },
 * });
 */
export function useCheckIpMutationMutation(baseOptions?: Apollo.MutationHookOptions<CheckIpMutationMutation, CheckIpMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CheckIpMutationMutation, CheckIpMutationMutationVariables>(CheckIpMutationDocument, options);
      }
export type CheckIpMutationMutationHookResult = ReturnType<typeof useCheckIpMutationMutation>;
export type CheckIpMutationMutationResult = Apollo.MutationResult<CheckIpMutationMutation>;
export type CheckIpMutationMutationOptions = Apollo.BaseMutationOptions<CheckIpMutationMutation, CheckIpMutationMutationVariables>;