import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type UpdateSurveyCustomAttributeMutationMutationVariables = Types.Exact<{
  surveyJSON: Types.Scalars['String']['input'];
}>;


export type UpdateSurveyCustomAttributeMutationMutation = { __typename?: 'Mutation', updateSurveyCustomAttribute?: boolean | null };


export const UpdateSurveyCustomAttributeMutationDocument = gql`
    mutation updateSurveyCustomAttributeMutation($surveyJSON: String!) {
  updateSurveyCustomAttribute(input: {surveyJSON: $surveyJSON})
}
    `;
export type UpdateSurveyCustomAttributeMutationMutationFn = Apollo.MutationFunction<UpdateSurveyCustomAttributeMutationMutation, UpdateSurveyCustomAttributeMutationMutationVariables>;

/**
 * __useUpdateSurveyCustomAttributeMutationMutation__
 *
 * To run a mutation, you first call `useUpdateSurveyCustomAttributeMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useUpdateSurveyCustomAttributeMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [updateSurveyCustomAttributeMutationMutation, { data, loading, error }] = useUpdateSurveyCustomAttributeMutationMutation({
 *   variables: {
 *      surveyJSON: // value for 'surveyJSON'
 *   },
 * });
 */
export function useUpdateSurveyCustomAttributeMutationMutation(baseOptions?: Apollo.MutationHookOptions<UpdateSurveyCustomAttributeMutationMutation, UpdateSurveyCustomAttributeMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<UpdateSurveyCustomAttributeMutationMutation, UpdateSurveyCustomAttributeMutationMutationVariables>(UpdateSurveyCustomAttributeMutationDocument, options);
      }
export type UpdateSurveyCustomAttributeMutationMutationHookResult = ReturnType<typeof useUpdateSurveyCustomAttributeMutationMutation>;
export type UpdateSurveyCustomAttributeMutationMutationResult = Apollo.MutationResult<UpdateSurveyCustomAttributeMutationMutation>;
export type UpdateSurveyCustomAttributeMutationMutationOptions = Apollo.BaseMutationOptions<UpdateSurveyCustomAttributeMutationMutation, UpdateSurveyCustomAttributeMutationMutationVariables>;