/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { IdentityDefinitionLoginID } from "./../../__generated__/globalTypes";

// ====================================================
// GraphQL mutation operation: CreateUserMutation
// ====================================================

export interface CreateUserMutation_createUser_user {
  __typename: "User";
  /**
   * The ID of an object
   */
  id: string;
}

export interface CreateUserMutation_createUser {
  __typename: "CreateUserPayload";
  user: CreateUserMutation_createUser_user;
}

export interface CreateUserMutation {
  /**
   * Create new user
   */
  createUser: CreateUserMutation_createUser;
}

export interface CreateUserMutationVariables {
  identityDefinition: IdentityDefinitionLoginID;
  password?: string | null;
}
