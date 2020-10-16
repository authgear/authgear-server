/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: DeleteCollaboratorMutation
// ====================================================

export interface DeleteCollaboratorMutation_deleteCollaborator_app_collaborators {
  __typename: "Collaborator";
  id: string;
  createdAt: GQL_DateTime;
  userID: string;
}

export interface DeleteCollaboratorMutation_deleteCollaborator_app {
  __typename: "App";
  /**
   * The ID of an object
   */
  id: string;
  collaborators: DeleteCollaboratorMutation_deleteCollaborator_app_collaborators[];
}

export interface DeleteCollaboratorMutation_deleteCollaborator {
  __typename: "DeleteCollaboratorPayload";
  app: DeleteCollaboratorMutation_deleteCollaborator_app;
}

export interface DeleteCollaboratorMutation {
  /**
   * Delete collaborator of target app.
   */
  deleteCollaborator: DeleteCollaboratorMutation_deleteCollaborator;
}

export interface DeleteCollaboratorMutationVariables {
  collaboratorID: string;
}
