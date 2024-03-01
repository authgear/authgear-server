import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateAppTemplatesMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  updates: Array<Types.AppResourceUpdate> | Types.AppResourceUpdate;
  paths: Array<Types.Scalars['String']['input']> | Types.Scalars['String']['input'];
}>;


export type UpdateAppTemplatesMutationMutation = { __typename?: 'Mutation', updateApp: { __typename?: 'UpdateAppPayload', app: { __typename?: 'App', id: string, resources: Array<{ __typename?: 'AppResource', path: string, languageTag?: string | null, data?: string | null, effectiveData?: string | null, checksum?: string | null }> } } };


export const UpdateAppTemplatesMutationDocument = gql`
    mutation updateAppTemplatesMutation($appID: ID!, $updates: [AppResourceUpdate!]!, $paths: [String!]!) {
  updateApp(input: {appID: $appID, updates: $updates}) {
    app {
      id
      resources(paths: $paths) {
        path
        languageTag
        data
        effectiveData
        checksum
      }
    }
  }
}
    `;
export type UpdateAppTemplatesMutationMutationFn = Apollo.MutationFunction<UpdateAppTemplatesMutationMutation, UpdateAppTemplatesMutationMutationVariables>;

/**
 * __useUpdateAppTemplatesMutationMutation__
 *
 * To run a mutation, you first call `useUpdateAppTemplatesMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateAppTemplatesMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateAppTemplatesMutationMutation, { data, loading, error }] = useUpdateAppTemplatesMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      updates: // value for 'updates'
 *      paths: // value for 'paths'
 *   },
 * });
 */
export function useUpdateAppTemplatesMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateAppTemplatesMutationMutation, UpdateAppTemplatesMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateAppTemplatesMutationMutation, UpdateAppTemplatesMutationMutationVariables>(UpdateAppTemplatesMutationDocument, options);
      }
export type UpdateAppTemplatesMutationMutationHookResult = ReturnType<typeof useUpdateAppTemplatesMutationMutation>;
export type UpdateAppTemplatesMutationMutationResult = Apollo.MutationResult<UpdateAppTemplatesMutationMutation>;
export type UpdateAppTemplatesMutationMutationOptions = Apollo.BaseMutationOptions<UpdateAppTemplatesMutationMutation, UpdateAppTemplatesMutationMutationVariables>;