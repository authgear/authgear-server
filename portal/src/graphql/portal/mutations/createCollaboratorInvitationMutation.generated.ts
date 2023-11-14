import * as Types from '../globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
const defaultOptions = {} as const;
export type CreateCollaboratorInvitationMutationMutationVariables = Types.Exact<{
  appID: Types.Scalars['ID']['input'];
  email: Types.Scalars['String']['input'];
}>;


export type CreateCollaboratorInvitationMutationMutation = { __typename?: 'Mutation', createCollaboratorInvitation: { __typename?: 'CreateCollaboratorInvitationPayload', collaboratorInvitation: { __typename?: 'CollaboratorInvitation', id: string, createdAt: any, expireAt: any, inviteeEmail: string, invitedBy: { __typename?: 'User', id: string, email?: string | null } } } };


export const CreateCollaboratorInvitationMutationDocument = gql`
    mutation createCollaboratorInvitationMutation($appID: ID!, $email: String!) {
  createCollaboratorInvitation(input: {appID: $appID, inviteeEmail: $email}) {
    collaboratorInvitation {
      id
      createdAt
      expireAt
      invitedBy {
        id
        email
      }
      inviteeEmail
    }
  }
}
    `;
export type CreateCollaboratorInvitationMutationMutationFn = Apollo.MutationFunction<CreateCollaboratorInvitationMutationMutation, CreateCollaboratorInvitationMutationMutationVariables>;

/**
 * __useCreateCollaboratorInvitationMutationMutation__
 *
 * To run a mutation, you first call `useCreateCollaboratorInvitationMutationMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateCollaboratorInvitationMutationMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createCollaboratorInvitationMutationMutation, { data, loading, error }] = useCreateCollaboratorInvitationMutationMutation({
 *   variables: {
 *      appID: // value for 'appID'
 *      email: // value for 'email'
 *   },
 * });
 */
export function useCreateCollaboratorInvitationMutationMutation(baseOptions?: Apollo.MutationHookOptions<CreateCollaboratorInvitationMutationMutation, CreateCollaboratorInvitationMutationMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateCollaboratorInvitationMutationMutation, CreateCollaboratorInvitationMutationMutationVariables>(CreateCollaboratorInvitationMutationDocument, options);
      }
export type CreateCollaboratorInvitationMutationMutationHookResult = ReturnType<typeof useCreateCollaboratorInvitationMutationMutation>;
export type CreateCollaboratorInvitationMutationMutationResult = Apollo.MutationResult<CreateCollaboratorInvitationMutationMutation>;
export type CreateCollaboratorInvitationMutationMutationOptions = Apollo.BaseMutationOptions<CreateCollaboratorInvitationMutationMutation, CreateCollaboratorInvitationMutationMutationVariables>;