import * as Types from './globalTypes.generated';

import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
const defaultOptions = {} as const;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
  /** The `AuditLogData` scalar type represents the data of the audit log */
  AuditLogData: any;
  /** The `AuthenticatorClaims` scalar type represents a set of claims belonging to an authenticator */
  AuthenticatorClaims: any;
  /** The `DateTime` scalar type represents a DateTime. The DateTime is serialized as an RFC 3339 quoted string */
  DateTime: any;
  /** The `IdentityClaims` scalar type represents a set of claims belonging to an identity */
  IdentityClaims: any;
  /** The `UserCustomAttributes` scalar type represents the custom attributes of the user */
  UserCustomAttributes: any;
  /** The `UserStandardAttributes` scalar type represents the standard attributes of the user */
  UserStandardAttributes: any;
};

/** Audit log */
export type AuditLog = Node & {
  __typename?: 'AuditLog';
  activityType: AuditLogActivityType;
  clientID?: Maybe<Scalars['String']>;
  createdAt: Scalars['DateTime'];
  data?: Maybe<Scalars['AuditLogData']>;
  /** The ID of an object */
  id: Scalars['ID'];
  ipAddress?: Maybe<Scalars['String']>;
  user?: Maybe<User>;
  userAgent?: Maybe<Scalars['String']>;
};

export enum AuditLogActivityType {
  AuthenticationIdentityAnonymousFailed = 'AUTHENTICATION_IDENTITY_ANONYMOUS_FAILED',
  AuthenticationIdentityBiometricFailed = 'AUTHENTICATION_IDENTITY_BIOMETRIC_FAILED',
  AuthenticationIdentityLoginIdFailed = 'AUTHENTICATION_IDENTITY_LOGIN_ID_FAILED',
  AuthenticationPrimaryOobOtpEmailFailed = 'AUTHENTICATION_PRIMARY_OOB_OTP_EMAIL_FAILED',
  AuthenticationPrimaryOobOtpSmsFailed = 'AUTHENTICATION_PRIMARY_OOB_OTP_SMS_FAILED',
  AuthenticationPrimaryPasswordFailed = 'AUTHENTICATION_PRIMARY_PASSWORD_FAILED',
  AuthenticationSecondaryOobOtpEmailFailed = 'AUTHENTICATION_SECONDARY_OOB_OTP_EMAIL_FAILED',
  AuthenticationSecondaryOobOtpSmsFailed = 'AUTHENTICATION_SECONDARY_OOB_OTP_SMS_FAILED',
  AuthenticationSecondaryPasswordFailed = 'AUTHENTICATION_SECONDARY_PASSWORD_FAILED',
  AuthenticationSecondaryRecoveryCodeFailed = 'AUTHENTICATION_SECONDARY_RECOVERY_CODE_FAILED',
  AuthenticationSecondaryTotpFailed = 'AUTHENTICATION_SECONDARY_TOTP_FAILED',
  EmailSent = 'EMAIL_SENT',
  IdentityEmailAdded = 'IDENTITY_EMAIL_ADDED',
  IdentityEmailRemoved = 'IDENTITY_EMAIL_REMOVED',
  IdentityEmailUpdated = 'IDENTITY_EMAIL_UPDATED',
  IdentityOauthConnected = 'IDENTITY_OAUTH_CONNECTED',
  IdentityOauthDisconnected = 'IDENTITY_OAUTH_DISCONNECTED',
  IdentityPhoneAdded = 'IDENTITY_PHONE_ADDED',
  IdentityPhoneRemoved = 'IDENTITY_PHONE_REMOVED',
  IdentityPhoneUpdated = 'IDENTITY_PHONE_UPDATED',
  IdentityUsernameAdded = 'IDENTITY_USERNAME_ADDED',
  IdentityUsernameRemoved = 'IDENTITY_USERNAME_REMOVED',
  IdentityUsernameUpdated = 'IDENTITY_USERNAME_UPDATED',
  SmsSent = 'SMS_SENT',
  UserAnonymousPromoted = 'USER_ANONYMOUS_PROMOTED',
  UserAuthenticated = 'USER_AUTHENTICATED',
  UserCreated = 'USER_CREATED',
  UserDeleted = 'USER_DELETED',
  UserDeletionScheduled = 'USER_DELETION_SCHEDULED',
  UserDeletionUnscheduled = 'USER_DELETION_UNSCHEDULED',
  UserDisabled = 'USER_DISABLED',
  UserProfileUpdated = 'USER_PROFILE_UPDATED',
  UserReenabled = 'USER_REENABLED',
  UserSignedOut = 'USER_SIGNED_OUT'
}

/** A connection to a list of items. */
export type AuditLogConnection = {
  __typename?: 'AuditLogConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<AuditLogEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']>;
};

/** An edge in a connection */
export type AuditLogEdge = {
  __typename?: 'AuditLogEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of the edge */
  node?: Maybe<AuditLog>;
};

export type Authenticator = Entity & Node & {
  __typename?: 'Authenticator';
  claims: Scalars['AuthenticatorClaims'];
  /** The creation time of entity */
  createdAt: Scalars['DateTime'];
  /** The ID of an object */
  id: Scalars['ID'];
  isDefault: Scalars['Boolean'];
  kind: AuthenticatorKind;
  type: AuthenticatorType;
  /** The update time of entity */
  updatedAt: Scalars['DateTime'];
};


export type AuthenticatorClaimsArgs = {
  names?: InputMaybe<Array<Scalars['String']>>;
};

/** A connection to a list of items. */
export type AuthenticatorConnection = {
  __typename?: 'AuthenticatorConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<AuthenticatorEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']>;
};

/** An edge in a connection */
export type AuthenticatorEdge = {
  __typename?: 'AuthenticatorEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of the edge */
  node?: Maybe<Authenticator>;
};

export enum AuthenticatorKind {
  Primary = 'PRIMARY',
  Secondary = 'SECONDARY'
}

export enum AuthenticatorType {
  OobOtpEmail = 'OOB_OTP_EMAIL',
  OobOtpSms = 'OOB_OTP_SMS',
  Password = 'PASSWORD',
  Totp = 'TOTP'
}

export type Claim = {
  __typename?: 'Claim';
  name: Scalars['String'];
  value: Scalars['String'];
};

export type CreateIdentityInput = {
  /** Definition of the new identity. */
  definition: IdentityDefinition;
  /** Password for the user if required. */
  password?: InputMaybe<Scalars['String']>;
  /** Target user ID. */
  userID: Scalars['ID'];
};

export type CreateIdentityPayload = {
  __typename?: 'CreateIdentityPayload';
  identity: Identity;
  user: User;
};

export type CreateUserInput = {
  /** Definition of the identity of new user. */
  definition: IdentityDefinition;
  /** Password for the user if required. */
  password?: InputMaybe<Scalars['String']>;
};

export type CreateUserPayload = {
  __typename?: 'CreateUserPayload';
  user: User;
};

export type DeleteAuthenticatorInput = {
  /** Target authenticator ID. */
  authenticatorID: Scalars['ID'];
};

export type DeleteAuthenticatorPayload = {
  __typename?: 'DeleteAuthenticatorPayload';
  user: User;
};

export type DeleteIdentityInput = {
  /** Target identity ID. */
  identityID: Scalars['ID'];
};

export type DeleteIdentityPayload = {
  __typename?: 'DeleteIdentityPayload';
  user: User;
};

export type DeleteUserInput = {
  /** Target user ID. */
  userID: Scalars['ID'];
};

export type DeleteUserPayload = {
  __typename?: 'DeleteUserPayload';
  deletedUserID: Scalars['ID'];
};

export type Entity = {
  /** The creation time of entity */
  createdAt: Scalars['DateTime'];
  /** The ID of entity */
  id: Scalars['ID'];
  /** The update time of entity */
  updatedAt: Scalars['DateTime'];
};

export type Identity = Entity & Node & {
  __typename?: 'Identity';
  claims: Scalars['IdentityClaims'];
  /** The creation time of entity */
  createdAt: Scalars['DateTime'];
  /** The ID of an object */
  id: Scalars['ID'];
  type: IdentityType;
  /** The update time of entity */
  updatedAt: Scalars['DateTime'];
};


export type IdentityClaimsArgs = {
  names?: InputMaybe<Array<Scalars['String']>>;
};

/** A connection to a list of items. */
export type IdentityConnection = {
  __typename?: 'IdentityConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<IdentityEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']>;
};

/** Definition of an identity. This is a union object, exactly one of the available fields must be present. */
export type IdentityDefinition = {
  /** Login ID identity definition. */
  loginID?: InputMaybe<IdentityDefinitionLoginId>;
};

export type IdentityDefinitionLoginId = {
  /** The login ID key. */
  key: Scalars['String'];
  /** The login ID. */
  value: Scalars['String'];
};

/** An edge in a connection */
export type IdentityEdge = {
  __typename?: 'IdentityEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of the edge */
  node?: Maybe<Identity>;
};

export enum IdentityType {
  Anonymous = 'ANONYMOUS',
  Biometric = 'BIOMETRIC',
  LoginId = 'LOGIN_ID',
  Oauth = 'OAUTH'
}

export type Mutation = {
  __typename?: 'Mutation';
  /** Create new identity for user */
  createIdentity: CreateIdentityPayload;
  /** Create new user */
  createUser: CreateUserPayload;
  /** Delete authenticator of user */
  deleteAuthenticator: DeleteAuthenticatorPayload;
  /** Delete identity of user */
  deleteIdentity: DeleteIdentityPayload;
  /** Delete specified user */
  deleteUser: DeleteUserPayload;
  /** Reset password of user */
  resetPassword: ResetPasswordPayload;
  /** Revoke all sessions of user */
  revokeAllSessions: RevokeAllSessionsPayload;
  /** Revoke session of user */
  revokeSession: RevokeSessionPayload;
  /** Schedule account deletion */
  scheduleAccountDeletion: ScheduleAccountDeletionPayload;
  /** Set disabled status of user */
  setDisabledStatus: SetDisabledStatusPayload;
  /** Set verified status of a claim of user */
  setVerifiedStatus: SetVerifiedStatusPayload;
  /** Unschedule account deletion */
  unscheduleAccountDeletion: UnscheduleAccountDeletionPayload;
  /** Update user */
  updateUser: UpdateUserPayload;
};


export type MutationCreateIdentityArgs = {
  input: CreateIdentityInput;
};


export type MutationCreateUserArgs = {
  input: CreateUserInput;
};


export type MutationDeleteAuthenticatorArgs = {
  input: DeleteAuthenticatorInput;
};


export type MutationDeleteIdentityArgs = {
  input: DeleteIdentityInput;
};


export type MutationDeleteUserArgs = {
  input: DeleteUserInput;
};


export type MutationResetPasswordArgs = {
  input: ResetPasswordInput;
};


export type MutationRevokeAllSessionsArgs = {
  input: RevokeAllSessionsInput;
};


export type MutationRevokeSessionArgs = {
  input: RevokeSessionInput;
};


export type MutationScheduleAccountDeletionArgs = {
  input: ScheduleAccountDeletionInput;
};


export type MutationSetDisabledStatusArgs = {
  input: SetDisabledStatusInput;
};


export type MutationSetVerifiedStatusArgs = {
  input: SetVerifiedStatusInput;
};


export type MutationUnscheduleAccountDeletionArgs = {
  input: UnscheduleAccountDeletionInput;
};


export type MutationUpdateUserArgs = {
  input: UpdateUserInput;
};

/** An object with an ID */
export type Node = {
  /** The id of the object */
  id: Scalars['ID'];
};

/** Information about pagination in a connection. */
export type PageInfo = {
  __typename?: 'PageInfo';
  /** When paginating forwards, the cursor to continue. */
  endCursor?: Maybe<Scalars['String']>;
  /** When paginating forwards, are there more items? */
  hasNextPage: Scalars['Boolean'];
  /** When paginating backwards, are there more items? */
  hasPreviousPage: Scalars['Boolean'];
  /** When paginating backwards, the cursor to continue. */
  startCursor?: Maybe<Scalars['String']>;
};

export type Query = {
  __typename?: 'Query';
  /** Audit logs */
  auditLogs?: Maybe<AuditLogConnection>;
  /** Fetches an object given its ID */
  node?: Maybe<Node>;
  /** Lookup nodes by a list of IDs. */
  nodes: Array<Maybe<Node>>;
  /** All users */
  users?: Maybe<UserConnection>;
};


export type QueryAuditLogsArgs = {
  activityTypes?: InputMaybe<Array<AuditLogActivityType>>;
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  rangeFrom?: InputMaybe<Scalars['DateTime']>;
  rangeTo?: InputMaybe<Scalars['DateTime']>;
};


export type QueryNodeArgs = {
  id: Scalars['ID'];
};


export type QueryNodesArgs = {
  ids: Array<Scalars['ID']>;
};


export type QueryUsersArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
  searchKeyword?: InputMaybe<Scalars['String']>;
  sortBy?: InputMaybe<UserSortBy>;
  sortDirection?: InputMaybe<SortDirection>;
};

export type ResetPasswordInput = {
  /** New password. */
  password: Scalars['String'];
  /** Target user ID. */
  userID: Scalars['ID'];
};

export type ResetPasswordPayload = {
  __typename?: 'ResetPasswordPayload';
  user: User;
};

export type RevokeAllSessionsInput = {
  /** Target user ID. */
  userID: Scalars['ID'];
};

export type RevokeAllSessionsPayload = {
  __typename?: 'RevokeAllSessionsPayload';
  user: User;
};

export type RevokeSessionInput = {
  /** Target session ID. */
  sessionID: Scalars['ID'];
};

export type RevokeSessionPayload = {
  __typename?: 'RevokeSessionPayload';
  user: User;
};

export type ScheduleAccountDeletionInput = {
  /** Target user ID. */
  userID: Scalars['ID'];
};

export type ScheduleAccountDeletionPayload = {
  __typename?: 'ScheduleAccountDeletionPayload';
  user: User;
};

export type Session = Entity & Node & {
  __typename?: 'Session';
  acr: Scalars['String'];
  amr: Array<Scalars['String']>;
  /** The creation time of entity */
  createdAt: Scalars['DateTime'];
  createdByIP: Scalars['String'];
  displayName: Scalars['String'];
  /** The ID of an object */
  id: Scalars['ID'];
  lastAccessedAt: Scalars['DateTime'];
  lastAccessedByIP: Scalars['String'];
  type: SessionType;
  /** The update time of entity */
  updatedAt: Scalars['DateTime'];
};

/** A connection to a list of items. */
export type SessionConnection = {
  __typename?: 'SessionConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<SessionEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']>;
};

/** An edge in a connection */
export type SessionEdge = {
  __typename?: 'SessionEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of the edge */
  node?: Maybe<Session>;
};

export enum SessionType {
  Idp = 'IDP',
  OfflineGrant = 'OFFLINE_GRANT'
}

export type SetDisabledStatusInput = {
  /** Indicate whether the target user is disabled. */
  isDisabled: Scalars['Boolean'];
  /** Indicate the disable reason; If not provided, the user will be disabled with no reason. */
  reason?: InputMaybe<Scalars['String']>;
  /** Target user ID. */
  userID: Scalars['ID'];
};

export type SetDisabledStatusPayload = {
  __typename?: 'SetDisabledStatusPayload';
  user: User;
};

export type SetVerifiedStatusInput = {
  /** Name of the claim to set verified status. */
  claimName: Scalars['String'];
  /** Value of the claim. */
  claimValue: Scalars['String'];
  /** Indicate whether the target claim is verified. */
  isVerified: Scalars['Boolean'];
  /** Target user ID. */
  userID: Scalars['ID'];
};

export type SetVerifiedStatusPayload = {
  __typename?: 'SetVerifiedStatusPayload';
  user: User;
};

export enum SortDirection {
  Asc = 'ASC',
  Desc = 'DESC'
}

export type UnscheduleAccountDeletionInput = {
  /** Target user ID. */
  userID: Scalars['ID'];
};

export type UnscheduleAccountDeletionPayload = {
  __typename?: 'UnscheduleAccountDeletionPayload';
  user: User;
};

export type UpdateUserInput = {
  /** Whole custom attributes to be set on the user. */
  customAttributes?: InputMaybe<Scalars['UserCustomAttributes']>;
  /** Whole standard attributes to be set on the user. */
  standardAttributes?: InputMaybe<Scalars['UserStandardAttributes']>;
  /** Target user ID. */
  userID: Scalars['ID'];
};

export type UpdateUserPayload = {
  __typename?: 'UpdateUserPayload';
  user: User;
};

/** Authgear user */
export type User = Entity & Node & {
  __typename?: 'User';
  authenticators?: Maybe<AuthenticatorConnection>;
  /** The creation time of entity */
  createdAt: Scalars['DateTime'];
  customAttributes: Scalars['UserCustomAttributes'];
  /** The scheduled deletion time of the user */
  deleteAt?: Maybe<Scalars['DateTime']>;
  disableReason?: Maybe<Scalars['String']>;
  endUserAccountID?: Maybe<Scalars['String']>;
  formattedName?: Maybe<Scalars['String']>;
  /** The ID of an object */
  id: Scalars['ID'];
  identities?: Maybe<IdentityConnection>;
  isAnonymous: Scalars['Boolean'];
  isDeactivated: Scalars['Boolean'];
  isDisabled: Scalars['Boolean'];
  /** The last login time of user */
  lastLoginAt?: Maybe<Scalars['DateTime']>;
  sessions?: Maybe<SessionConnection>;
  standardAttributes: Scalars['UserStandardAttributes'];
  /** The update time of entity */
  updatedAt: Scalars['DateTime'];
  verifiedClaims: Array<Claim>;
};


/** Authgear user */
export type UserAuthenticatorsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


/** Authgear user */
export type UserIdentitiesArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};


/** Authgear user */
export type UserSessionsArgs = {
  after?: InputMaybe<Scalars['String']>;
  before?: InputMaybe<Scalars['String']>;
  first?: InputMaybe<Scalars['Int']>;
  last?: InputMaybe<Scalars['Int']>;
};

/** A connection to a list of items. */
export type UserConnection = {
  __typename?: 'UserConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<UserEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']>;
};

/** An edge in a connection */
export type UserEdge = {
  __typename?: 'UserEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String'];
  /** The item at the end of the edge */
  node?: Maybe<User>;
};

export enum UserSortBy {
  CreatedAt = 'CREATED_AT',
  LastLoginAt = 'LAST_LOGIN_AT'
}

export type UsersListQueryQueryVariables = Types.Exact<{
  searchKeyword: Types.Scalars['String'];
  pageSize: Types.Scalars['Int'];
  cursor?: Types.InputMaybe<Types.Scalars['String']>;
  sortBy?: Types.InputMaybe<Types.UserSortBy>;
  sortDirection?: Types.InputMaybe<Types.SortDirection>;
}>;


export type UsersListQueryQuery = { __typename?: 'Query', users?: { __typename?: 'UserConnection', totalCount?: number | null, edges?: Array<{ __typename?: 'UserEdge', node?: { __typename?: 'User', id: string, createdAt: any, lastLoginAt?: any | null, isAnonymous: boolean, isDisabled: boolean, disableReason?: string | null, isDeactivated: boolean, deleteAt?: any | null, standardAttributes: any, formattedName?: string | null, endUserAccountID?: string | null } | null } | null> | null } | null };


export const UsersListQueryDocument = gql`
    query UsersListQuery($searchKeyword: String!, $pageSize: Int!, $cursor: String, $sortBy: UserSortBy, $sortDirection: SortDirection) {
  users(
    first: $pageSize
    after: $cursor
    searchKeyword: $searchKeyword
    sortBy: $sortBy
    sortDirection: $sortDirection
  ) {
    edges {
      node {
        id
        createdAt
        lastLoginAt
        isAnonymous
        isDisabled
        disableReason
        isDeactivated
        deleteAt
        standardAttributes
        formattedName
        endUserAccountID
      }
    }
    totalCount
  }
}
    `;

/**
 * __useUsersListQueryQuery__
 *
 * To run a query within a React component, call `useUsersListQueryQuery` and pass it any options that fit your needs.
 * When your component renders, `useUsersListQueryQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUsersListQueryQuery({
 *   variables: {
 *      searchKeyword: // value for 'searchKeyword'
 *      pageSize: // value for 'pageSize'
 *      cursor: // value for 'cursor'
 *      sortBy: // value for 'sortBy'
 *      sortDirection: // value for 'sortDirection'
 *   },
 * });
 */
export function useUsersListQueryQuery(baseOptions: Apollo.QueryHookOptions<UsersListQueryQuery, UsersListQueryQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<UsersListQueryQuery, UsersListQueryQueryVariables>(UsersListQueryDocument, options);
      }
export function useUsersListQueryLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<UsersListQueryQuery, UsersListQueryQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<UsersListQueryQuery, UsersListQueryQueryVariables>(UsersListQueryDocument, options);
        }
export type UsersListQueryQueryHookResult = ReturnType<typeof useUsersListQueryQuery>;
export type UsersListQueryLazyQueryHookResult = ReturnType<typeof useUsersListQueryLazyQuery>;
export type UsersListQueryQueryResult = Apollo.QueryResult<UsersListQueryQuery, UsersListQueryQueryVariables>;