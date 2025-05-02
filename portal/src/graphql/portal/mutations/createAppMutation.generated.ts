import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateAppMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['String']['input'];
  projectWizardData: Types.Scalars['ProjectWizardData']['input'];
}>;


export type CreateAppMutationMutation = { __typename?: 'Mutation', createApp: { __typename?: 'CreateAppPayload', app: { __typename?: 'App', id: string } } };


export const CreateAppMutationDocument = gql`
    mutation createAppMutation($appID: String!, $projectWizardData: ProjectWizardData!) {
  createApp(input: {id: $appID, projectWizardData: $projectWizardData}) {
    app {
      id
    }
  }
}
    `;
export type CreateAppMutationMutationFn = Apollo.MutationFunction<CreateAppMutationMutation, CreateAppMutationMutationVariables>;

/**
 * __useCreateAppMutationMutation__
 *
 * To run a mutation, you first call `useCreateAppMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateAppMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createAppMutationMutation, { data, loading, error }] = useCreateAppMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      projectWizardData: // value for 'projectWizardData'
 *   },
 * });
 */
export function useCreateAppMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateAppMutationMutation, CreateAppMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateAppMutationMutation, CreateAppMutationMutationVariables>(CreateAppMutationDocument, options);
      }
export type CreateAppMutationMutationHookResult = ReturnType<typeof useCreateAppMutationMutation>;
export type CreateAppMutationMutationResult = Apollo.MutationResult<CreateAppMutationMutation>;
export type CreateAppMutationMutationOptions = Apollo.BaseMutationOptions<CreateAppMutationMutation, CreateAppMutationMutationVariables>;