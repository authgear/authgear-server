import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type VerifyDomainMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  domainID: Types.Scalars['String']['input'];
}>;


export type VerifyDomainMutationMutation = { __typename?: 'Mutation', verifyDomain: { __typename?: 'VerifyDomainPayload', app: { __typename?: 'App', id: string, domains: Array<{ __typename?: 'Domain', id: string, createdAt: any, domain: string, cookieDomain: string, apexDomain: string, isCustom: boolean, isVerified: boolean, verificationDNSRecord: string }> }, domain: { __typename?: 'Domain', id: string, createdAt: any, domain: string, cookieDomain: string, apexDomain: string, isCustom: boolean, isVerified: boolean, verificationDNSRecord: string } } };


export const VerifyDomainMutationDocument = gql`
    mutation verifyDomainMutation($appID: ID!, $domainID: String!) {
  verifyDomain(input: {appID: $appID, domainID: $domainID}) {
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
    }
    domain {
      id
      createdAt
      domain
      cookieDomain
      apexDomain
      isCustom
      isVerified
      verificationDNSRecord
    }
  }
}
    `;
export type VerifyDomainMutationMutationFn = Apollo.MutationFunction<VerifyDomainMutationMutation, VerifyDomainMutationMutationVariables>;

/**
 * __useVerifyDomainMutationMutation__
 *
 * To run a mutation, you first call `useVerifyDomainMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useVerifyDomainMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [verifyDomainMutationMutation, { data, loading, error }] = useVerifyDomainMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      domainID: // value for 'domainID'
 *   },
 * });
 */
export function useVerifyDomainMutationMutation(baseOptions?: Apollo.MutationHookOptions<VerifyDomainMutationMutation, VerifyDomainMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<VerifyDomainMutationMutation, VerifyDomainMutationMutationVariables>(VerifyDomainMutationDocument, options);
      }
export type VerifyDomainMutationMutationHookResult = ReturnType<typeof useVerifyDomainMutationMutation>;
export type VerifyDomainMutationMutationResult = Apollo.MutationResult<VerifyDomainMutationMutation>;
export type VerifyDomainMutationMutationOptions = Apollo.BaseMutationOptions<VerifyDomainMutationMutation, VerifyDomainMutationMutationVariables>;