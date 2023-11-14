import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type DeleteCollaboratorMutationMutationVariables = Types.Exact<{
  collaboratorID: Types.Scalars['String']['input'];
}>;


export type DeleteCollaboratorMutationMutation = { __typename?: 'Mutation', deleteCollaborator: { __typename?: 'DeleteCollaboratorPayload', app: { __typename?: 'App', id: string, collaborators: Array<{ __typename?: 'Collaborator', id: string, createdAt: any, user: { __typename?: 'User', id: string, email?: string | null } }> } } };


export const DeleteCollaboratorMutationDocument = gql`
    mutation deleteCollaboratorMutation($collaboratorID: String!) {
  deleteCollaborator(input: {collaboratorID: $collaboratorID}) {
    app {
      id
      collaborators {
        id
        createdAt
        user {
          id
          email
        }
      }
    }
  }
}
    `;
export type DeleteCollaboratorMutationMutationFn = Apollo.MutationFunction<DeleteCollaboratorMutationMutation, DeleteCollaboratorMutationMutationVariables>;

/**
 * __useDeleteCollaboratorMutationMutation__
 *
 * To run a mutation, you first call `useDeleteCollaboratorMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteCollaboratorMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteCollaboratorMutationMutation, { data, loading, error }] = useDeleteCollaboratorMutationMutation({
 *   variables: {
 *      collaboratorID: // value for 'collaboratorID'
 *   },
 * });
 */
export function useDeleteCollaboratorMutationMutation(baseOptions?: Apollo.MutationHookOptions<DeleteCollaboratorMutationMutation, DeleteCollaboratorMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteCollaboratorMutationMutation, DeleteCollaboratorMutationMutationVariables>(DeleteCollaboratorMutationDocument, options);
      }
export type DeleteCollaboratorMutationMutationHookResult = ReturnType<typeof useDeleteCollaboratorMutationMutation>;
export type DeleteCollaboratorMutationMutationResult = Apollo.MutationResult<DeleteCollaboratorMutationMutation>;
export type DeleteCollaboratorMutationMutationOptions = Apollo.BaseMutationOptions<DeleteCollaboratorMutationMutation, DeleteCollaboratorMutationMutationVariables>;