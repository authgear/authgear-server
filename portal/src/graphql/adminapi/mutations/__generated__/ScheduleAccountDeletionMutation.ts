/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: ScheduleAccountDeletionMutation
// ====================================================

export interface ScheduleAccountDeletionMutation_scheduleAccountDeletion_user {
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

export interface ScheduleAccountDeletionMutation_scheduleAccountDeletion {
  __typename: "ScheduleAccountDeletionPayload";
  user: ScheduleAccountDeletionMutation_scheduleAccountDeletion_user;
}

export interface ScheduleAccountDeletionMutation {
  /**
   * Schedule account deletion
   */
  scheduleAccountDeletion: ScheduleAccountDeletionMutation_scheduleAccountDeletion;
}

export interface ScheduleAccountDeletionMutationVariables {
  userID: string;
}
