import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateDomainMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  domain: Types.Scalars['String']['input'];
}>;


export type CreateDomainMutationMutation = { __typename?: 'Mutation', createDomain: { __typename?: 'CreateDomainPayload', app: { __typename?: 'App', id: string, domains: Array<{ __typename?: 'Domain', id: string, createdAt: any, domain: string, cookieDomain: string, apexDomain: string, isCustom: boolean, isVerified: boolean, verificationDNSRecord: string }> }, domain: { __typename?: 'Domain', id: string, createdAt: any, domain: string, cookieDomain: string, apexDomain: string, isCustom: boolean, isVerified: boolean, verificationDNSRecord: string } } };


export const CreateDomainMutationDocument = gql`
    mutation createDomainMutation($appID: ID!, $domain: String!) {
  createDomain(input: {appID: $appID, domain: $domain}) {
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
export type CreateDomainMutationMutationFn = Apollo.MutationFunction<CreateDomainMutationMutation, CreateDomainMutationMutationVariables>;

/**
 * __useCreateDomainMutationMutation__
 *
 * To run a mutation, you first call `useCreateDomainMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateDomainMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createDomainMutationMutation, { data, loading, error }] = useCreateDomainMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      domain: // value for 'domain'
 *   },
 * });
 */
export function useCreateDomainMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateDomainMutationMutation, CreateDomainMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateDomainMutationMutation, CreateDomainMutationMutationVariables>(CreateDomainMutationDocument, options);
      }
export type CreateDomainMutationMutationHookResult = ReturnType<typeof useCreateDomainMutationMutation>;
export type CreateDomainMutationMutationResult = Apollo.MutationResult<CreateDomainMutationMutation>;
export type CreateDomainMutationMutationOptions = Apollo.BaseMutationOptions<CreateDomainMutationMutation, CreateDomainMutationMutationVariables>;