import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type SkipAppTutorialMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['String']['input'];
}>;


export type SkipAppTutorialMutationMutation = { __typename?: 'Mutation', skipAppTutorial: { __typename?: 'SkipAppTutorialPayload', app: { __typename?: 'App', id: string } } };


export const SkipAppTutorialMutationDocument = gql`
    mutation skipAppTutorialMutation($appID: String!) {
  skipAppTutorial(input: {id: $appID}) {
    app {
      id
    }
  }
}
    `;
export type SkipAppTutorialMutationMutationFn = Apollo.MutationFunction<SkipAppTutorialMutationMutation, SkipAppTutorialMutationMutationVariables>;

/**
 * __useSkipAppTutorialMutationMutation__
 *
 * To run a mutation, you first call `useSkipAppTutorialMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useSkipAppTutorialMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [skipAppTutorialMutationMutation, { data, loading, error }] = useSkipAppTutorialMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *   },
 * });
 */
export function useSkipAppTutorialMutationMutation(baseOptions?: Apollo.MutationHookOptions<SkipAppTutorialMutationMutation, SkipAppTutorialMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<SkipAppTutorialMutationMutation, SkipAppTutorialMutationMutationVariables>(SkipAppTutorialMutationDocument, options);
      }
export type SkipAppTutorialMutationMutationHookResult = ReturnType<typeof useSkipAppTutorialMutationMutation>;
export type SkipAppTutorialMutationMutationResult = Apollo.MutationResult<SkipAppTutorialMutationMutation>;
export type SkipAppTutorialMutationMutationOptions = Apollo.BaseMutationOptions<SkipAppTutorialMutationMutation, SkipAppTutorialMutationMutationVariables>;