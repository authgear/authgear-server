import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SaveProjectWizardDataMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['String']['input'];
  data?: Types.InputMaybe<Types.Scalars['ProjectWizardData']['input']>;
}>;


export type SaveProjectWizardDataMutationMutation = { __typename?: 'Mutation', saveProjectWizardData: { __typename?: 'SaveProjectWizardDataPayload', app: { __typename?: 'App', id: string } } };


export const SaveProjectWizardDataMutationDocument = gql`
    mutation saveProjectWizardDataMutation($appID: String!, $data: ProjectWizardData) {
  saveProjectWizardData(input: {id: $appID, data: $data}) {
    app {
      id
    }
  }
}
    `;
export type SaveProjectWizardDataMutationMutationFn = Apollo.MutationFunction<SaveProjectWizardDataMutationMutation, SaveProjectWizardDataMutationMutationVariables>;

/**
 * __useSaveProjectWizardDataMutationMutation__
 *
 * To run a mutation, you first call `useSaveProjectWizardDataMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSaveProjectWizardDataMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [saveProjectWizardDataMutationMutation, { data, loading, error }] = useSaveProjectWizardDataMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      data: // value for 'data'
 *   },
 * });
 */
export function useSaveProjectWizardDataMutationMutation(baseOptions?: Apollo.MutationHookOptions<SaveProjectWizardDataMutationMutation, SaveProjectWizardDataMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SaveProjectWizardDataMutationMutation, SaveProjectWizardDataMutationMutationVariables>(SaveProjectWizardDataMutationDocument, options);
      }
export type SaveProjectWizardDataMutationMutationHookResult = ReturnType<typeof useSaveProjectWizardDataMutationMutation>;
export type SaveProjectWizardDataMutationMutationResult = Apollo.MutationResult<SaveProjectWizardDataMutationMutation>;
export type SaveProjectWizardDataMutationMutationOptions = Apollo.BaseMutationOptions<SaveProjectWizardDataMutationMutation, SaveProjectWizardDataMutationMutationVariables>;