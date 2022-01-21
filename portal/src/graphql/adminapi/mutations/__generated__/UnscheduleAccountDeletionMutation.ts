/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UnscheduleAccountDeletionMutation
// ====================================================

export interface UnscheduleAccountDeletionMutation_unscheduleAccountDeletion_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
  isDisabled: boolean;
  disableReason: string | null;
  isDeactivated: boolean;
  /**
   * The scheduled deletion time of the user
   */
  deleteAt: GQL_DateTime | null;
}

export interface UnscheduleAccountDeletionMutation_unscheduleAccountDeletion {
  __typename: "UnscheduleAccountDeletionPayload";
  user: UnscheduleAccountDeletionMutation_unscheduleAccountDeletion_user;
}

export interface UnscheduleAccountDeletionMutation {
  /**
   * Unschedule account deletion
   */
  unscheduleAccountDeletion: UnscheduleAccountDeletionMutation_unscheduleAccountDeletion;
}

export interface UnscheduleAccountDeletionMutationVariables {
  userID: string;
}
