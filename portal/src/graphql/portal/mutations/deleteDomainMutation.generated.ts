import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DeleteDomainMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  domainID: Types.Scalars['String']['input'];
}>;


export type DeleteDomainMutationMutation = { __typename?: 'Mutation', deleteDomain: { __typename?: 'DeleteDomainPayload', app: { __typename?: 'App', id: string, rawAppConfig: any, effectiveAppConfig: any, domains: Array<{ __typename?: 'Domain', id: string, createdAt: any, domain: string, cookieDomain: string, apexDomain: string, isCustom: boolean, isVerified: boolean, verificationDNSRecord: string }> } } };


export const DeleteDomainMutationDocument = gql`
    mutation deleteDomainMutation($appID: ID!, $domainID: String!) {
  deleteDomain(input: {appID: $appID, domainID: $domainID}) {
    app {
      id
      domains {
        id
        createdAt
        domain
        cookieDomain
        apexDomain
        isCustom
        isVerified
        verificationDNSRecord
      }
      rawAppConfig
      effectiveAppConfig
    }
  }
}
    `;
export type DeleteDomainMutationMutationFn = Apollo.MutationFunction<DeleteDomainMutationMutation, DeleteDomainMutationMutationVariables>;

/**
 * __useDeleteDomainMutationMutation__
 *
 * To run a mutation, you first call `useDeleteDomainMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteDomainMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteDomainMutationMutation, { data, loading, error }] = useDeleteDomainMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      domainID: // value for 'domainID'
 *   },
 * });
 */
export function useDeleteDomainMutationMutation(baseOptions?: Apollo.MutationHookOptions<DeleteDomainMutationMutation, DeleteDomainMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteDomainMutationMutation, DeleteDomainMutationMutationVariables>(DeleteDomainMutationDocument, options);
      }
export type DeleteDomainMutationMutationHookResult = ReturnType<typeof useDeleteDomainMutationMutation>;
export type DeleteDomainMutationMutationResult = Apollo.MutationResult<DeleteDomainMutationMutation>;
export type DeleteDomainMutationMutationOptions = Apollo.BaseMutationOptions<DeleteDomainMutationMutation, DeleteDomainMutationMutationVariables>;