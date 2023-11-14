import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SkipAppTutorialProgressMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['String']['input'];
  progress: Types.Scalars['String']['input'];
}>;


export type SkipAppTutorialProgressMutationMutation = { __typename?: 'Mutation', skipAppTutorialProgress: { __typename?: 'SkipAppTutorialProgressPayload', app: { __typename?: 'App', id: string, tutorialStatus: { __typename?: 'TutorialStatus', data: any } } } };


export const SkipAppTutorialProgressMutationDocument = gql`
    mutation skipAppTutorialProgressMutation($appID: String!, $progress: String!) {
  skipAppTutorialProgress(input: {id: $appID, progress: $progress}) {
    app {
      id
      tutorialStatus {
        data
      }
    }
  }
}
    `;
export type SkipAppTutorialProgressMutationMutationFn = Apollo.MutationFunction<SkipAppTutorialProgressMutationMutation, SkipAppTutorialProgressMutationMutationVariables>;

/**
 * __useSkipAppTutorialProgressMutationMutation__
 *
 * To run a mutation, you first call `useSkipAppTutorialProgressMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSkipAppTutorialProgressMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [skipAppTutorialProgressMutationMutation, { data, loading, error }] = useSkipAppTutorialProgressMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      progress: // value for 'progress'
 *   },
 * });
 */
export function useSkipAppTutorialProgressMutationMutation(baseOptions?: Apollo.MutationHookOptions<SkipAppTutorialProgressMutationMutation, SkipAppTutorialProgressMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SkipAppTutorialProgressMutationMutation, SkipAppTutorialProgressMutationMutationVariables>(SkipAppTutorialProgressMutationDocument, options);
      }
export type SkipAppTutorialProgressMutationMutationHookResult = ReturnType<typeof useSkipAppTutorialProgressMutationMutation>;
export type SkipAppTutorialProgressMutationMutationResult = Apollo.MutationResult<SkipAppTutorialProgressMutationMutation>;
export type SkipAppTutorialProgressMutationMutationOptions = Apollo.BaseMutationOptions<SkipAppTutorialProgressMutationMutation, SkipAppTutorialProgressMutationMutationVariables>;