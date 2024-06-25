import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SaveOnboardingSurveyMutationMutationVariables = Types.Exact<{
  surveyJSON: Types.Scalars['String']['input'];
}>;


export type SaveOnboardingSurveyMutationMutation = { __typename?: 'Mutation', saveOnboardingSurvey?: boolean | null };


export const SaveOnboardingSurveyMutationDocument = gql`
    mutation saveOnboardingSurveyMutation($surveyJSON: String!) {
  saveOnboardingSurvey(input: {surveyJSON: $surveyJSON})
}
    `;
export type SaveOnboardingSurveyMutationMutationFn = Apollo.MutationFunction<SaveOnboardingSurveyMutationMutation, SaveOnboardingSurveyMutationMutationVariables>;

/**
 * __useSaveOnboardingSurveyMutationMutation__
 *
 * To run a mutation, you first call `useSaveOnboardingSurveyMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSaveOnboardingSurveyMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [saveOnboardingSurveyMutationMutation, { data, loading, error }] = useSaveOnboardingSurveyMutationMutation({
 *   variables: {
 *      surveyJSON: // value for 'surveyJSON'
 *   },
 * });
 */
export function useSaveOnboardingSurveyMutationMutation(baseOptions?: Apollo.MutationHookOptions<SaveOnboardingSurveyMutationMutation, SaveOnboardingSurveyMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SaveOnboardingSurveyMutationMutation, SaveOnboardingSurveyMutationMutationVariables>(SaveOnboardingSurveyMutationDocument, options);
      }
export type SaveOnboardingSurveyMutationMutationHookResult = ReturnType<typeof useSaveOnboardingSurveyMutationMutation>;
export type SaveOnboardingSurveyMutationMutationResult = Apollo.MutationResult<SaveOnboardingSurveyMutationMutation>;
export type SaveOnboardingSurveyMutationMutationOptions = Apollo.BaseMutationOptions<SaveOnboardingSurveyMutationMutation, SaveOnboardingSurveyMutationMutationVariables>;