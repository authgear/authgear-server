export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  /** The `AuditLogData` scalar type represents the data of the audit log */
  AuditLogData: { input: GQL_AuditLogData; output: GQL_AuditLogData; }
  /** The `AuthenticatorClaims` scalar type represents a set of claims belonging to an authenticator */
  AuthenticatorClaims: { input: GQL_AuthenticatorClaims; output: GQL_AuthenticatorClaims; }
  /** The `DateTime` scalar type represents a DateTime. The DateTime is serialized as an RFC 3339 quoted string */
  DateTime: { input: GQL_DateTime; output: GQL_DateTime; }
  /** The `IdentityClaims` scalar type represents a set of claims belonging to an identity */
  IdentityClaims: { input: GQL_IdentityClaims; output: GQL_IdentityClaims; }
  /** The `UserCustomAttributes` scalar type represents the custom attributes of the user */
  UserCustomAttributes: { input: GQL_UserCustomAttributes; output: GQL_UserCustomAttributes; }
  /** The `UserStandardAttributes` scalar type represents the standard attributes of the user */
  UserStandardAttributes: { input: GQL_UserStandardAttributes; output: GQL_UserStandardAttributes; }
  /** The `Web3Claims` scalar type represents the scalar type of the user */
  Web3Claims: { input: GQL_Web3Claims; output: GQL_Web3Claims; }
};

export type AddGroupToRolesInput = {
  /** The key of the group. */
  groupKey: Scalars['String']['input'];
  /** The list of role keys. */
  roleKeys: Array<Scalars['String']['input']>;
};

export type AddGroupToRolesPayload = {
  __typename?: 'AddGroupToRolesPayload';
  group: Group;
};

export type AddGroupToUsersInput = {
  /** The key of the group. */
  groupKey: Scalars['String']['input'];
  /** The list of user ids. */
  userIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
};

export type AddGroupToUsersPayload = {
  __typename?: 'AddGroupToUsersPayload';
  group: Group;
};

export type AddRoleToGroupsInput = {
  /** The list of group keys. */
  groupKeys: Array<Scalars['String']['input']>;
  /** The key of the role. */
  roleKey: Scalars['String']['input'];
};

export type AddRoleToGroupsPayload = {
  __typename?: 'AddRoleToGroupsPayload';
  role: Role;
};

export type AddRoleToUsersInput = {
  /** The key of the role. */
  roleKey: Scalars['String']['input'];
  /** The list of user ids. */
  userIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
};

export type AddRoleToUsersPayload = {
  __typename?: 'AddRoleToUsersPayload';
  role: Role;
};

export type AddUserToGroupsInput = {
  /** The list of group keys. */
  groupKeys?: InputMaybe<Array<Scalars['String']['input']>>;
  /** The ID of the user. */
  userID: Scalars['ID']['input'];
};

export type AddUserToGroupsPayload = {
  __typename?: 'AddUserToGroupsPayload';
  user: User;
};

export type AddUserToRolesInput = {
  /** The list of role keys. */
  roleKeys?: InputMaybe<Array<Scalars['String']['input']>>;
  /** The id of the user. */
  userID: Scalars['ID']['input'];
};

export type AddUserToRolesPayload = {
  __typename?: 'AddUserToRolesPayload';
  user: User;
};

export type AnonymizeUserInput = {
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type AnonymizeUserPayload = {
  __typename?: 'AnonymizeUserPayload';
  anonymizedUserID: Scalars['ID']['output'];
};

/** Audit log */
export type AuditLog = Node & {
  __typename?: 'AuditLog';
  activityType: AuditLogActivityType;
  clientID?: Maybe<Scalars['String']['output']>;
  createdAt: Scalars['DateTime']['output'];
  data?: Maybe<Scalars['AuditLogData']['output']>;
  /** The ID of an object */
  id: Scalars['ID']['output'];
  ipAddress?: Maybe<Scalars['String']['output']>;
  user?: Maybe<User>;
  userAgent?: Maybe<Scalars['String']['output']>;
};

export enum AuditLogActivityType {
  AdminApiMutationAddGroupToRolesExecuted = 'ADMIN_API_MUTATION_ADD_GROUP_TO_ROLES_EXECUTED',
  AdminApiMutationAddGroupToUsersExecuted = 'ADMIN_API_MUTATION_ADD_GROUP_TO_USERS_EXECUTED',
  AdminApiMutationAddRoleToGroupsExecuted = 'ADMIN_API_MUTATION_ADD_ROLE_TO_GROUPS_EXECUTED',
  AdminApiMutationAddRoleToUsersExecuted = 'ADMIN_API_MUTATION_ADD_ROLE_TO_USERS_EXECUTED',
  AdminApiMutationAddUserToGroupsExecuted = 'ADMIN_API_MUTATION_ADD_USER_TO_GROUPS_EXECUTED',
  AdminApiMutationAddUserToRolesExecuted = 'ADMIN_API_MUTATION_ADD_USER_TO_ROLES_EXECUTED',
  AdminApiMutationAnonymizeUserExecuted = 'ADMIN_API_MUTATION_ANONYMIZE_USER_EXECUTED',
  AdminApiMutationCreateAuthenticatorExecuted = 'ADMIN_API_MUTATION_CREATE_AUTHENTICATOR_EXECUTED',
  AdminApiMutationCreateGroupExecuted = 'ADMIN_API_MUTATION_CREATE_GROUP_EXECUTED',
  AdminApiMutationCreateIdentityExecuted = 'ADMIN_API_MUTATION_CREATE_IDENTITY_EXECUTED',
  AdminApiMutationCreateRoleExecuted = 'ADMIN_API_MUTATION_CREATE_ROLE_EXECUTED',
  AdminApiMutationCreateSessionExecuted = 'ADMIN_API_MUTATION_CREATE_SESSION_EXECUTED',
  AdminApiMutationCreateUserExecuted = 'ADMIN_API_MUTATION_CREATE_USER_EXECUTED',
  AdminApiMutationDeleteAuthenticatorExecuted = 'ADMIN_API_MUTATION_DELETE_AUTHENTICATOR_EXECUTED',
  AdminApiMutationDeleteAuthorizationExecuted = 'ADMIN_API_MUTATION_DELETE_AUTHORIZATION_EXECUTED',
  AdminApiMutationDeleteGroupExecuted = 'ADMIN_API_MUTATION_DELETE_GROUP_EXECUTED',
  AdminApiMutationDeleteIdentityExecuted = 'ADMIN_API_MUTATION_DELETE_IDENTITY_EXECUTED',
  AdminApiMutationDeleteRoleExecuted = 'ADMIN_API_MUTATION_DELETE_ROLE_EXECUTED',
  AdminApiMutationDeleteUserExecuted = 'ADMIN_API_MUTATION_DELETE_USER_EXECUTED',
  AdminApiMutationGenerateOobOtpCodeExecuted = 'ADMIN_API_MUTATION_GENERATE_OOB_OTP_CODE_EXECUTED',
  AdminApiMutationRemoveGroupFromRolesExecuted = 'ADMIN_API_MUTATION_REMOVE_GROUP_FROM_ROLES_EXECUTED',
  AdminApiMutationRemoveGroupFromUsersExecuted = 'ADMIN_API_MUTATION_REMOVE_GROUP_FROM_USERS_EXECUTED',
  AdminApiMutationRemoveRoleFromGroupsExecuted = 'ADMIN_API_MUTATION_REMOVE_ROLE_FROM_GROUPS_EXECUTED',
  AdminApiMutationRemoveRoleFromUsersExecuted = 'ADMIN_API_MUTATION_REMOVE_ROLE_FROM_USERS_EXECUTED',
  AdminApiMutationRemoveUserFromGroupsExecuted = 'ADMIN_API_MUTATION_REMOVE_USER_FROM_GROUPS_EXECUTED',
  AdminApiMutationRemoveUserFromRolesExecuted = 'ADMIN_API_MUTATION_REMOVE_USER_FROM_ROLES_EXECUTED',
  AdminApiMutationResetPasswordExecuted = 'ADMIN_API_MUTATION_RESET_PASSWORD_EXECUTED',
  AdminApiMutationRevokeAllSessionsExecuted = 'ADMIN_API_MUTATION_REVOKE_ALL_SESSIONS_EXECUTED',
  AdminApiMutationRevokeSessionExecuted = 'ADMIN_API_MUTATION_REVOKE_SESSION_EXECUTED',
  AdminApiMutationScheduleAccountAnonymizationExecuted = 'ADMIN_API_MUTATION_SCHEDULE_ACCOUNT_ANONYMIZATION_EXECUTED',
  AdminApiMutationScheduleAccountDeletionExecuted = 'ADMIN_API_MUTATION_SCHEDULE_ACCOUNT_DELETION_EXECUTED',
  AdminApiMutationSendResetPasswordMessageExecuted = 'ADMIN_API_MUTATION_SEND_RESET_PASSWORD_MESSAGE_EXECUTED',
  AdminApiMutationSetDisabledStatusExecuted = 'ADMIN_API_MUTATION_SET_DISABLED_STATUS_EXECUTED',
  AdminApiMutationSetPasswordExpiredExecuted = 'ADMIN_API_MUTATION_SET_PASSWORD_EXPIRED_EXECUTED',
  AdminApiMutationSetVerifiedStatusExecuted = 'ADMIN_API_MUTATION_SET_VERIFIED_STATUS_EXECUTED',
  AdminApiMutationUnscheduleAccountAnonymizationExecuted = 'ADMIN_API_MUTATION_UNSCHEDULE_ACCOUNT_ANONYMIZATION_EXECUTED',
  AdminApiMutationUnscheduleAccountDeletionExecuted = 'ADMIN_API_MUTATION_UNSCHEDULE_ACCOUNT_DELETION_EXECUTED',
  AdminApiMutationUpdateGroupExecuted = 'ADMIN_API_MUTATION_UPDATE_GROUP_EXECUTED',
  AdminApiMutationUpdateIdentityExecuted = 'ADMIN_API_MUTATION_UPDATE_IDENTITY_EXECUTED',
  AdminApiMutationUpdateRoleExecuted = 'ADMIN_API_MUTATION_UPDATE_ROLE_EXECUTED',
  AdminApiMutationUpdateUserExecuted = 'ADMIN_API_MUTATION_UPDATE_USER_EXECUTED',
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
  IdentityBiometricDisabled = 'IDENTITY_BIOMETRIC_DISABLED',
  IdentityBiometricEnabled = 'IDENTITY_BIOMETRIC_ENABLED',
  IdentityEmailAdded = 'IDENTITY_EMAIL_ADDED',
  IdentityEmailRemoved = 'IDENTITY_EMAIL_REMOVED',
  IdentityEmailUnverified = 'IDENTITY_EMAIL_UNVERIFIED',
  IdentityEmailUpdated = 'IDENTITY_EMAIL_UPDATED',
  IdentityEmailVerified = 'IDENTITY_EMAIL_VERIFIED',
  IdentityOauthConnected = 'IDENTITY_OAUTH_CONNECTED',
  IdentityOauthDisconnected = 'IDENTITY_OAUTH_DISCONNECTED',
  IdentityPhoneAdded = 'IDENTITY_PHONE_ADDED',
  IdentityPhoneRemoved = 'IDENTITY_PHONE_REMOVED',
  IdentityPhoneUnverified = 'IDENTITY_PHONE_UNVERIFIED',
  IdentityPhoneUpdated = 'IDENTITY_PHONE_UPDATED',
  IdentityPhoneVerified = 'IDENTITY_PHONE_VERIFIED',
  IdentityUsernameAdded = 'IDENTITY_USERNAME_ADDED',
  IdentityUsernameRemoved = 'IDENTITY_USERNAME_REMOVED',
  IdentityUsernameUpdated = 'IDENTITY_USERNAME_UPDATED',
  ProjectAppSecretViewed = 'PROJECT_APP_SECRET_VIEWED',
  ProjectAppUpdated = 'PROJECT_APP_UPDATED',
  ProjectBillingCheckoutCreated = 'PROJECT_BILLING_CHECKOUT_CREATED',
  ProjectBillingSubscriptionCancelled = 'PROJECT_BILLING_SUBSCRIPTION_CANCELLED',
  ProjectBillingSubscriptionStatusUpdated = 'PROJECT_BILLING_SUBSCRIPTION_STATUS_UPDATED',
  ProjectBillingSubscriptionUpdated = 'PROJECT_BILLING_SUBSCRIPTION_UPDATED',
  ProjectCollaboratorDeleted = 'PROJECT_COLLABORATOR_DELETED',
  ProjectCollaboratorInvitationAccepted = 'PROJECT_COLLABORATOR_INVITATION_ACCEPTED',
  ProjectCollaboratorInvitationCreated = 'PROJECT_COLLABORATOR_INVITATION_CREATED',
  ProjectCollaboratorInvitationDeleted = 'PROJECT_COLLABORATOR_INVITATION_DELETED',
  ProjectDomainCreated = 'PROJECT_DOMAIN_CREATED',
  ProjectDomainDeleted = 'PROJECT_DOMAIN_DELETED',
  ProjectDomainVerified = 'PROJECT_DOMAIN_VERIFIED',
  SmsSent = 'SMS_SENT',
  UserAnonymizationScheduled = 'USER_ANONYMIZATION_SCHEDULED',
  UserAnonymizationUnscheduled = 'USER_ANONYMIZATION_UNSCHEDULED',
  UserAnonymized = 'USER_ANONYMIZED',
  UserAnonymousPromoted = 'USER_ANONYMOUS_PROMOTED',
  UserAuthenticated = 'USER_AUTHENTICATED',
  UserCreated = 'USER_CREATED',
  UserDeleted = 'USER_DELETED',
  UserDeletionScheduled = 'USER_DELETION_SCHEDULED',
  UserDeletionUnscheduled = 'USER_DELETION_UNSCHEDULED',
  UserDisabled = 'USER_DISABLED',
  UserProfileUpdated = 'USER_PROFILE_UPDATED',
  UserReenabled = 'USER_REENABLED',
  UserSessionTerminated = 'USER_SESSION_TERMINATED',
  UserSignedOut = 'USER_SIGNED_OUT',
  WhatsappOtpVerified = 'WHATSAPP_OTP_VERIFIED',
  WhatsappSent = 'WHATSAPP_SENT'
}

/** A connection to a list of items. */
export type AuditLogConnection = {
  __typename?: 'AuditLogConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<AuditLogEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']['output']>;
};

/** An edge in a connection */
export type AuditLogEdge = {
  __typename?: 'AuditLogEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of the edge */
  node?: Maybe<AuditLog>;
};

export type Authenticator = Entity & Node & {
  __typename?: 'Authenticator';
  claims: Scalars['AuthenticatorClaims']['output'];
  /** The creation time of entity */
  createdAt: Scalars['DateTime']['output'];
  expireAfter?: Maybe<Scalars['DateTime']['output']>;
  /** The ID of an object */
  id: Scalars['ID']['output'];
  isDefault: Scalars['Boolean']['output'];
  kind: AuthenticatorKind;
  type: AuthenticatorType;
  /** The update time of entity */
  updatedAt: Scalars['DateTime']['output'];
};


export type AuthenticatorClaimsArgs = {
  names?: InputMaybe<Array<Scalars['String']['input']>>;
};

/** A connection to a list of items. */
export type AuthenticatorConnection = {
  __typename?: 'AuthenticatorConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<AuthenticatorEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']['output']>;
};

/** Definition of an authenticator. This is a union object, exactly one of the available fields must be present. */
export type AuthenticatorDefinition = {
  /** Kind of authenticator. */
  kind: AuthenticatorKind;
  /** OOB OTP Email authenticator definition. Must be provided when type is OOB_OTP_EMAIL. */
  oobOtpEmail?: InputMaybe<AuthenticatorDefinitionOobotpEmail>;
  /** OOB OTP SMS authenticator definition. Must be provided when type is OOB_OTP_SMS. */
  oobOtpSMS?: InputMaybe<AuthenticatorDefinitionOobotpsms>;
  /** Password authenticator definition. Must be provided when type is PASSWORD. */
  password?: InputMaybe<AuthenticatorDefinitionPassword>;
  /** Type of authenticator. */
  type: AuthenticatorType;
};

export type AuthenticatorDefinitionOobotpEmail = {
  /** Email of the new oob otp sms authenticator. */
  email: Scalars['String']['input'];
};

export type AuthenticatorDefinitionOobotpsms = {
  /** Phone number of the new oob otp sms authenticator. */
  phone: Scalars['String']['input'];
};

export type AuthenticatorDefinitionPassword = {
  /** Password of the new authenticator. */
  password: Scalars['String']['input'];
};

/** An edge in a connection */
export type AuthenticatorEdge = {
  __typename?: 'AuthenticatorEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String']['output'];
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
  Passkey = 'PASSKEY',
  Password = 'PASSWORD',
  Totp = 'TOTP'
}

export type Authorization = Entity & Node & {
  __typename?: 'Authorization';
  clientID: Scalars['String']['output'];
  /** The creation time of entity */
  createdAt: Scalars['DateTime']['output'];
  /** The ID of an object */
  id: Scalars['ID']['output'];
  scopes: Array<Scalars['String']['output']>;
  /** The update time of entity */
  updatedAt: Scalars['DateTime']['output'];
};

/** A connection to a list of items. */
export type AuthorizationConnection = {
  __typename?: 'AuthorizationConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<AuthorizationEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']['output']>;
};

/** An edge in a connection */
export type AuthorizationEdge = {
  __typename?: 'AuthorizationEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of the edge */
  node?: Maybe<Authorization>;
};

export type Claim = {
  __typename?: 'Claim';
  name: Scalars['String']['output'];
  value: Scalars['String']['output'];
};

export type CreateAuthenticatorInput = {
  /** Definition of the new authenticator. */
  definition: AuthenticatorDefinition;
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type CreateAuthenticatorPayload = {
  __typename?: 'CreateAuthenticatorPayload';
  authenticator: Authenticator;
};

export type CreateGroupInput = {
  /** The optional description of the group. */
  description?: InputMaybe<Scalars['String']['input']>;
  /** The key of the group. */
  key: Scalars['String']['input'];
  /** The optional name of the group. */
  name?: InputMaybe<Scalars['String']['input']>;
};

export type CreateGroupPayload = {
  __typename?: 'CreateGroupPayload';
  group: Group;
};

export type CreateIdentityInput = {
  /** Definition of the new identity. */
  definition: IdentityDefinition;
  /** Password for the user if required. */
  password?: InputMaybe<Scalars['String']['input']>;
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type CreateIdentityPayload = {
  __typename?: 'CreateIdentityPayload';
  identity: Identity;
  user: User;
};

export type CreateRoleInput = {
  /** The optional description of the role. */
  description?: InputMaybe<Scalars['String']['input']>;
  /** The key of the role. */
  key: Scalars['String']['input'];
  /** The optional name of the role. */
  name?: InputMaybe<Scalars['String']['input']>;
};

export type CreateRolePayload = {
  __typename?: 'CreateRolePayload';
  role: Role;
};

export type CreateSessionInput = {
  /** Target client ID. */
  clientID: Scalars['String']['input'];
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type CreateSessionPayload = {
  __typename?: 'CreateSessionPayload';
  accessToken: Scalars['String']['output'];
  refreshToken: Scalars['String']['output'];
};

export type CreateUserInput = {
  /** Definition of the identity of new user. */
  definition: IdentityDefinition;
  /** Password for the user if required. */
  password?: InputMaybe<Scalars['String']['input']>;
  /** Indicate whether to send the new password to the user. */
  sendPassword?: InputMaybe<Scalars['Boolean']['input']>;
  /** Indicate whether the user is required to change password on next login. */
  setPasswordExpired?: InputMaybe<Scalars['Boolean']['input']>;
};

export type CreateUserPayload = {
  __typename?: 'CreateUserPayload';
  user: User;
};

export type DeleteAuthenticatorInput = {
  /** Target authenticator ID. */
  authenticatorID: Scalars['ID']['input'];
};

export type DeleteAuthenticatorPayload = {
  __typename?: 'DeleteAuthenticatorPayload';
  user: User;
};

export type DeleteAuthorizationInput = {
  /** Target authorization ID. */
  authorizationID: Scalars['ID']['input'];
};

export type DeleteAuthorizationPayload = {
  __typename?: 'DeleteAuthorizationPayload';
  user: User;
};

export type DeleteGroupInput = {
  /** The ID of the group. */
  id: Scalars['ID']['input'];
};

export type DeleteGroupPayload = {
  __typename?: 'DeleteGroupPayload';
  ok?: Maybe<Scalars['Boolean']['output']>;
};

export type DeleteIdentityInput = {
  /** Target identity ID. */
  identityID: Scalars['ID']['input'];
};

export type DeleteIdentityPayload = {
  __typename?: 'DeleteIdentityPayload';
  user: User;
};

export type DeleteRoleInput = {
  /** The ID of the role. */
  id: Scalars['ID']['input'];
};

export type DeleteRolePayload = {
  __typename?: 'DeleteRolePayload';
  ok?: Maybe<Scalars['Boolean']['output']>;
};

export type DeleteUserInput = {
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type DeleteUserPayload = {
  __typename?: 'DeleteUserPayload';
  deletedUserID: Scalars['ID']['output'];
};

export type Entity = {
  /** The creation time of entity */
  createdAt: Scalars['DateTime']['output'];
  /** The ID of entity */
  id: Scalars['ID']['output'];
  /** The update time of entity */
  updatedAt: Scalars['DateTime']['output'];
};

export type GenerateOobotpCodeInput = {
  /** Purpose of the generated OTP code. */
  purpose?: InputMaybe<OtpPurpose>;
  /** Target user's email or phone number. */
  target: Scalars['String']['input'];
};

export type GenerateOobotpCodePayload = {
  __typename?: 'GenerateOOBOTPCodePayload';
  code: Scalars['String']['output'];
};

/** Authgear group */
export type Group = Entity & Node & {
  __typename?: 'Group';
  /** The creation time of entity */
  createdAt: Scalars['DateTime']['output'];
  /** The optional description of the group. */
  description?: Maybe<Scalars['String']['output']>;
  /** The ID of an object */
  id: Scalars['ID']['output'];
  /** The key of the group. */
  key: Scalars['String']['output'];
  /** The optional name of the group. */
  name?: Maybe<Scalars['String']['output']>;
  /** The list of roles this group has. */
  roles?: Maybe<RoleConnection>;
  /** The update time of entity */
  updatedAt: Scalars['DateTime']['output'];
  /** The list of users in the group. */
  users?: Maybe<UserConnection>;
};


/** Authgear group */
export type GroupRolesArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


/** Authgear group */
export type GroupUsersArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};

/** A connection to a list of items. */
export type GroupConnection = {
  __typename?: 'GroupConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<GroupEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']['output']>;
};

/** An edge in a connection */
export type GroupEdge = {
  __typename?: 'GroupEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of the edge */
  node?: Maybe<Group>;
};

export type Identity = Entity & Node & {
  __typename?: 'Identity';
  claims: Scalars['IdentityClaims']['output'];
  /** The creation time of entity */
  createdAt: Scalars['DateTime']['output'];
  /** The ID of an object */
  id: Scalars['ID']['output'];
  type: IdentityType;
  /** The update time of entity */
  updatedAt: Scalars['DateTime']['output'];
};


export type IdentityClaimsArgs = {
  names?: InputMaybe<Array<Scalars['String']['input']>>;
};

/** A connection to a list of items. */
export type IdentityConnection = {
  __typename?: 'IdentityConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<IdentityEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']['output']>;
};

/** Definition of an identity. This is a union object, exactly one of the available fields must be present. */
export type IdentityDefinition = {
  /** Login ID identity definition. */
  loginID?: InputMaybe<IdentityDefinitionLoginId>;
};

export type IdentityDefinitionLoginId = {
  /** The login ID key. */
  key: Scalars['String']['input'];
  /** The login ID. */
  value: Scalars['String']['input'];
};

/** An edge in a connection */
export type IdentityEdge = {
  __typename?: 'IdentityEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of the edge */
  node?: Maybe<Identity>;
};

export enum IdentityType {
  Anonymous = 'ANONYMOUS',
  Biometric = 'BIOMETRIC',
  Ldap = 'LDAP',
  LoginId = 'LOGIN_ID',
  Oauth = 'OAUTH',
  Passkey = 'PASSKEY',
  Siwe = 'SIWE'
}

export type Mutation = {
  __typename?: 'Mutation';
  /** Add the group to the roles. */
  addGroupToRoles: AddGroupToRolesPayload;
  /** Add the group to the users. */
  addGroupToUsers: AddGroupToUsersPayload;
  /** Add the role to the groups. */
  addRoleToGroups: AddRoleToGroupsPayload;
  /** Add the role to the users. */
  addRoleToUsers: AddRoleToUsersPayload;
  /** Add the user to the groups. */
  addUserToGroups: AddUserToGroupsPayload;
  /** Add the user to the roles. */
  addUserToRoles: AddUserToRolesPayload;
  /** Anonymize specified user */
  anonymizeUser: AnonymizeUserPayload;
  /** Create authenticator of user */
  createAuthenticator: CreateAuthenticatorPayload;
  /** Create a new group. */
  createGroup: CreateGroupPayload;
  /** Create new identity for user */
  createIdentity: CreateIdentityPayload;
  /** Create a new role. */
  createRole: CreateRolePayload;
  /** Create a session of a given user */
  createSession: CreateSessionPayload;
  /** Create new user */
  createUser: CreateUserPayload;
  /** Delete authenticator of user */
  deleteAuthenticator: DeleteAuthenticatorPayload;
  /** Delete authorization */
  deleteAuthorization: DeleteAuthorizationPayload;
  /** Delete an existing group. The associations between the group with other roles and other users will also be deleted. */
  deleteGroup: DeleteGroupPayload;
  /** Delete identity of user */
  deleteIdentity: DeleteIdentityPayload;
  /** Delete an existing role. The associations between the role with other groups and other users will also be deleted. */
  deleteRole: DeleteRolePayload;
  /** Delete specified user */
  deleteUser: DeleteUserPayload;
  /** Generate OOB OTP code for user */
  generateOOBOTPCode: GenerateOobotpCodePayload;
  /** Remove the group from the roles. */
  removeGroupFromRoles: RemoveGroupFromRolesPayload;
  /** Remove the group to the users. */
  removeGroupFromUsers: RemoveGroupToUsersPayload;
  /** Revoke user grace period for MFA enrollment */
  removeMFAGracePeriod: RemoveMfaGracePeriodPayload;
  /** Remove the role from the groups. */
  removeRoleFromGroups: RemoveRoleFromGroupsPayload;
  /** Remove the role to the users. */
  removeRoleFromUsers: RemoveRoleFromUsersPayload;
  /** Remove the user from the groups. */
  removeUserFromGroups: RemoveUserFromGroupsPayload;
  /** Remove the user from the roles. */
  removeUserFromRoles: RemoveUserFromRolesPayload;
  /** Reset password of user */
  resetPassword: ResetPasswordPayload;
  /** Revoke all sessions of user */
  revokeAllSessions: RevokeAllSessionsPayload;
  /** Revoke session of user */
  revokeSession: RevokeSessionPayload;
  /** Schedule account anonymization */
  scheduleAccountAnonymization: ScheduleAccountAnonymizationPayload;
  /** Schedule account deletion */
  scheduleAccountDeletion: ScheduleAccountDeletionPayload;
  /** Send a reset password message to user */
  sendResetPasswordMessage?: Maybe<Scalars['Boolean']['output']>;
  /** Set disabled status of user */
  setDisabledStatus: SetDisabledStatusPayload;
  /** Grant user grace period for MFA enrollment */
  setMFAGracePeriod: SetMfaGracePeriodPayload;
  /** Force user to change password on next login */
  setPasswordExpired: SetPasswordExpiredPayload;
  /** Set verified status of a claim of user */
  setVerifiedStatus: SetVerifiedStatusPayload;
  /** Unschedule account anonymization */
  unscheduleAccountAnonymization: UnscheduleAccountAnonymizationPayload;
  /** Unschedule account deletion */
  unscheduleAccountDeletion: UnscheduleAccountDeletionPayload;
  /** Update an existing group. */
  updateGroup: UpdateGroupPayload;
  /** Update an existing identity of user */
  updateIdentity: UpdateIdentityPayload;
  /** Update an existing role. */
  updateRole: UpdateRolePayload;
  /** Update user */
  updateUser: UpdateUserPayload;
};


export type MutationAddGroupToRolesArgs = {
  input: AddGroupToRolesInput;
};


export type MutationAddGroupToUsersArgs = {
  input: AddGroupToUsersInput;
};


export type MutationAddRoleToGroupsArgs = {
  input: AddRoleToGroupsInput;
};


export type MutationAddRoleToUsersArgs = {
  input: AddRoleToUsersInput;
};


export type MutationAddUserToGroupsArgs = {
  input: AddUserToGroupsInput;
};


export type MutationAddUserToRolesArgs = {
  input: AddUserToRolesInput;
};


export type MutationAnonymizeUserArgs = {
  input: AnonymizeUserInput;
};


export type MutationCreateAuthenticatorArgs = {
  input: CreateAuthenticatorInput;
};


export type MutationCreateGroupArgs = {
  input: CreateGroupInput;
};


export type MutationCreateIdentityArgs = {
  input: CreateIdentityInput;
};


export type MutationCreateRoleArgs = {
  input: CreateRoleInput;
};


export type MutationCreateSessionArgs = {
  input: CreateSessionInput;
};


export type MutationCreateUserArgs = {
  input: CreateUserInput;
};


export type MutationDeleteAuthenticatorArgs = {
  input: DeleteAuthenticatorInput;
};


export type MutationDeleteAuthorizationArgs = {
  input: DeleteAuthorizationInput;
};


export type MutationDeleteGroupArgs = {
  input: DeleteGroupInput;
};


export type MutationDeleteIdentityArgs = {
  input: DeleteIdentityInput;
};


export type MutationDeleteRoleArgs = {
  input: DeleteRoleInput;
};


export type MutationDeleteUserArgs = {
  input: DeleteUserInput;
};


export type MutationGenerateOobotpCodeArgs = {
  input: GenerateOobotpCodeInput;
};


export type MutationRemoveGroupFromRolesArgs = {
  input: RemoveGroupFromRolesInput;
};


export type MutationRemoveGroupFromUsersArgs = {
  input: RemoveGroupFromUsersInput;
};


export type MutationRemoveMfaGracePeriodArgs = {
  input: RemoveMfaGracePeriodInput;
};


export type MutationRemoveRoleFromGroupsArgs = {
  input: RemoveRoleFromGroupsInput;
};


export type MutationRemoveRoleFromUsersArgs = {
  input: RemoveRoleFromUsersInput;
};


export type MutationRemoveUserFromGroupsArgs = {
  input: RemoveUserFromGroupsInput;
};


export type MutationRemoveUserFromRolesArgs = {
  input: RemoveUserFromRolesInput;
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


export type MutationScheduleAccountAnonymizationArgs = {
  input: ScheduleAccountAnonymizationInput;
};


export type MutationScheduleAccountDeletionArgs = {
  input: ScheduleAccountDeletionInput;
};


export type MutationSendResetPasswordMessageArgs = {
  input: SendResetPasswordMessageInput;
};


export type MutationSetDisabledStatusArgs = {
  input: SetDisabledStatusInput;
};


export type MutationSetMfaGracePeriodArgs = {
  input: SetMfaGracePeriodInput;
};


export type MutationSetPasswordExpiredArgs = {
  input: SetPasswordExpiredInput;
};


export type MutationSetVerifiedStatusArgs = {
  input: SetVerifiedStatusInput;
};


export type MutationUnscheduleAccountAnonymizationArgs = {
  input: UnscheduleAccountAnonymizationInput;
};


export type MutationUnscheduleAccountDeletionArgs = {
  input: UnscheduleAccountDeletionInput;
};


export type MutationUpdateGroupArgs = {
  input: UpdateGroupInput;
};


export type MutationUpdateIdentityArgs = {
  input: UpdateIdentityInput;
};


export type MutationUpdateRoleArgs = {
  input: UpdateRoleInput;
};


export type MutationUpdateUserArgs = {
  input: UpdateUserInput;
};

/** An object with an ID */
export type Node = {
  /** The id of the object */
  id: Scalars['ID']['output'];
};

export enum OtpPurpose {
  Login = 'LOGIN',
  Verification = 'VERIFICATION'
}

/** Information about pagination in a connection. */
export type PageInfo = {
  __typename?: 'PageInfo';
  /** When paginating forwards, the cursor to continue. */
  endCursor?: Maybe<Scalars['String']['output']>;
  /** When paginating forwards, are there more items? */
  hasNextPage: Scalars['Boolean']['output'];
  /** When paginating backwards, are there more items? */
  hasPreviousPage: Scalars['Boolean']['output'];
  /** When paginating backwards, the cursor to continue. */
  startCursor?: Maybe<Scalars['String']['output']>;
};

export type Query = {
  __typename?: 'Query';
  /** Audit logs */
  auditLogs?: Maybe<AuditLogConnection>;
  /** All groups */
  groups?: Maybe<GroupConnection>;
  /** Fetches an object given its ID */
  node?: Maybe<Node>;
  /** Lookup nodes by a list of IDs. */
  nodes: Array<Maybe<Node>>;
  /** All roles */
  roles?: Maybe<RoleConnection>;
  /** All users */
  users?: Maybe<UserConnection>;
};


export type QueryAuditLogsArgs = {
  activityTypes?: InputMaybe<Array<AuditLogActivityType>>;
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  rangeFrom?: InputMaybe<Scalars['DateTime']['input']>;
  rangeTo?: InputMaybe<Scalars['DateTime']['input']>;
  sortDirection?: InputMaybe<SortDirection>;
  userIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
};


export type QueryGroupsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  excludedIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  searchKeyword?: InputMaybe<Scalars['String']['input']>;
};


export type QueryNodeArgs = {
  id: Scalars['ID']['input'];
};


export type QueryNodesArgs = {
  ids: Array<Scalars['ID']['input']>;
};


export type QueryRolesArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  excludedIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
  searchKeyword?: InputMaybe<Scalars['String']['input']>;
};


export type QueryUsersArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  groupKeys?: InputMaybe<Array<Scalars['String']['input']>>;
  last?: InputMaybe<Scalars['Int']['input']>;
  roleKeys?: InputMaybe<Array<Scalars['String']['input']>>;
  searchKeyword?: InputMaybe<Scalars['String']['input']>;
  sortBy?: InputMaybe<UserSortBy>;
  sortDirection?: InputMaybe<SortDirection>;
};

export type RemoveGroupFromRolesInput = {
  /** The key of the group. */
  groupKey: Scalars['String']['input'];
  /** The list of role keys. */
  roleKeys: Array<Scalars['String']['input']>;
};

export type RemoveGroupFromRolesPayload = {
  __typename?: 'RemoveGroupFromRolesPayload';
  group: Group;
};

export type RemoveGroupFromUsersInput = {
  /** The key of the group. */
  groupKey: Scalars['String']['input'];
  /** The list of user ids. */
  userIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
};

export type RemoveGroupToUsersPayload = {
  __typename?: 'RemoveGroupToUsersPayload';
  group: Group;
};

export type RemoveRoleFromGroupsInput = {
  /** The list of group keys. */
  groupKeys: Array<Scalars['String']['input']>;
  /** The key of the role. */
  roleKey: Scalars['String']['input'];
};

export type RemoveRoleFromGroupsPayload = {
  __typename?: 'RemoveRoleFromGroupsPayload';
  role: Role;
};

export type RemoveRoleFromUsersInput = {
  /** The key of the role. */
  roleKey: Scalars['String']['input'];
  /** The list of user ids. */
  userIDs?: InputMaybe<Array<Scalars['ID']['input']>>;
};

export type RemoveRoleFromUsersPayload = {
  __typename?: 'RemoveRoleFromUsersPayload';
  role: Role;
};

export type RemoveUserFromGroupsInput = {
  /** The list of group keys. */
  groupKeys?: InputMaybe<Array<Scalars['String']['input']>>;
  /** The ID of the user. */
  userID: Scalars['ID']['input'];
};

export type RemoveUserFromGroupsPayload = {
  __typename?: 'RemoveUserFromGroupsPayload';
  user: User;
};

export type RemoveUserFromRolesInput = {
  /** The list of role keys. */
  roleKeys?: InputMaybe<Array<Scalars['String']['input']>>;
  /** The id of the user. */
  userID: Scalars['ID']['input'];
};

export type RemoveUserFromRolesPayload = {
  __typename?: 'RemoveUserFromRolesPayload';
  user: User;
};

export type ResetPasswordInput = {
  /** New password. */
  password?: InputMaybe<Scalars['String']['input']>;
  /** Indicate whether to send the new password to the user. */
  sendPassword?: InputMaybe<Scalars['Boolean']['input']>;
  /** Indicate whether the user is required to change password on next login. */
  setPasswordExpired?: InputMaybe<Scalars['Boolean']['input']>;
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type ResetPasswordPayload = {
  __typename?: 'ResetPasswordPayload';
  user: User;
};

export type RevokeAllSessionsInput = {
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type RevokeAllSessionsPayload = {
  __typename?: 'RevokeAllSessionsPayload';
  user: User;
};

export type RevokeSessionInput = {
  /** Target session ID. */
  sessionID: Scalars['ID']['input'];
};

export type RevokeSessionPayload = {
  __typename?: 'RevokeSessionPayload';
  user: User;
};

/** Authgear role */
export type Role = Entity & Node & {
  __typename?: 'Role';
  /** The creation time of entity */
  createdAt: Scalars['DateTime']['output'];
  /** The optional description of the role. */
  description?: Maybe<Scalars['String']['output']>;
  /** The list of groups this role is in. */
  groups?: Maybe<GroupConnection>;
  /** The ID of an object */
  id: Scalars['ID']['output'];
  /** The key of the role. */
  key: Scalars['String']['output'];
  /** The optional name of the role. */
  name?: Maybe<Scalars['String']['output']>;
  /** The update time of entity */
  updatedAt: Scalars['DateTime']['output'];
  /** The list of users who has this role. */
  users?: Maybe<UserConnection>;
};


/** Authgear role */
export type RoleGroupsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


/** Authgear role */
export type RoleUsersArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};

/** A connection to a list of items. */
export type RoleConnection = {
  __typename?: 'RoleConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<RoleEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']['output']>;
};

/** An edge in a connection */
export type RoleEdge = {
  __typename?: 'RoleEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of the edge */
  node?: Maybe<Role>;
};

export type ScheduleAccountAnonymizationInput = {
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type ScheduleAccountAnonymizationPayload = {
  __typename?: 'ScheduleAccountAnonymizationPayload';
  user: User;
};

export type ScheduleAccountDeletionInput = {
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type ScheduleAccountDeletionPayload = {
  __typename?: 'ScheduleAccountDeletionPayload';
  user: User;
};

export type SendResetPasswordMessageInput = {
  /** Target login ID. */
  loginID: Scalars['ID']['input'];
};

export type Session = Entity & Node & {
  __typename?: 'Session';
  acr: Scalars['String']['output'];
  amr: Array<Scalars['String']['output']>;
  clientID?: Maybe<Scalars['String']['output']>;
  /** The creation time of entity */
  createdAt: Scalars['DateTime']['output'];
  createdByIP: Scalars['String']['output'];
  displayName: Scalars['String']['output'];
  /** The ID of an object */
  id: Scalars['ID']['output'];
  lastAccessedAt: Scalars['DateTime']['output'];
  lastAccessedByIP: Scalars['String']['output'];
  type: SessionType;
  /** The update time of entity */
  updatedAt: Scalars['DateTime']['output'];
  userAgent?: Maybe<Scalars['String']['output']>;
};

/** A connection to a list of items. */
export type SessionConnection = {
  __typename?: 'SessionConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<SessionEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']['output']>;
};

/** An edge in a connection */
export type SessionEdge = {
  __typename?: 'SessionEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of the edge */
  node?: Maybe<Session>;
};

export enum SessionType {
  Idp = 'IDP',
  OfflineGrant = 'OFFLINE_GRANT'
}

export type SetDisabledStatusInput = {
  /** Indicate whether the target user is disabled. */
  isDisabled: Scalars['Boolean']['input'];
  /** Indicate the disable reason; If not provided, the user will be disabled with no reason. */
  reason?: InputMaybe<Scalars['String']['input']>;
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type SetDisabledStatusPayload = {
  __typename?: 'SetDisabledStatusPayload';
  user: User;
};

export type SetMfaGracePeriodInput = {
  /** Indicate when will user's MFA grace period end */
  endAt: Scalars['DateTime']['input'];
  /** Target user ID */
  userID: Scalars['ID']['input'];
};

export type SetMfaGracePeriodPayload = {
  __typename?: 'SetMFAGracePeriodPayload';
  user: User;
};

export type SetPasswordExpiredInput = {
  /** Indicate whether the user's password is expired. */
  expired: Scalars['Boolean']['input'];
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type SetPasswordExpiredPayload = {
  __typename?: 'SetPasswordExpiredPayload';
  user: User;
};

export type SetVerifiedStatusInput = {
  /** Name of the claim to set verified status. */
  claimName: Scalars['String']['input'];
  /** Value of the claim. */
  claimValue: Scalars['String']['input'];
  /** Indicate whether the target claim is verified. */
  isVerified: Scalars['Boolean']['input'];
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type SetVerifiedStatusPayload = {
  __typename?: 'SetVerifiedStatusPayload';
  user: User;
};

export enum SortDirection {
  Asc = 'ASC',
  Desc = 'DESC'
}

export type UnscheduleAccountAnonymizationInput = {
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type UnscheduleAccountAnonymizationPayload = {
  __typename?: 'UnscheduleAccountAnonymizationPayload';
  user: User;
};

export type UnscheduleAccountDeletionInput = {
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type UnscheduleAccountDeletionPayload = {
  __typename?: 'UnscheduleAccountDeletionPayload';
  user: User;
};

export type UpdateGroupInput = {
  /** The new description of the group. Pass null if you do not need to update the description. Pass an empty string to remove the description. */
  description?: InputMaybe<Scalars['String']['input']>;
  /** The ID of the group. */
  id: Scalars['ID']['input'];
  /** The new key of the group. Pass null if you do not need to update the key. */
  key?: InputMaybe<Scalars['String']['input']>;
  /** The new name of the group. Pass null if you do not need to update the name. Pass an empty string to remove the name. */
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateGroupPayload = {
  __typename?: 'UpdateGroupPayload';
  group: Group;
};

export type UpdateIdentityInput = {
  /** New definition of the identity. */
  definition: IdentityDefinition;
  /** Target identity ID. */
  identityID: Scalars['ID']['input'];
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type UpdateIdentityPayload = {
  __typename?: 'UpdateIdentityPayload';
  identity: Identity;
  user: User;
};

export type UpdateRoleInput = {
  /** The new description of the role. Pass null if you do not need to update the description. Pass an empty string to remove the description. */
  description?: InputMaybe<Scalars['String']['input']>;
  /** The ID of the role. */
  id: Scalars['ID']['input'];
  /** The new key of the role. Pass null if you do not need to update the key. */
  key?: InputMaybe<Scalars['String']['input']>;
  /** The new name of the role. Pass null if you do not need to update the name. Pass an empty string to remove the name. */
  name?: InputMaybe<Scalars['String']['input']>;
};

export type UpdateRolePayload = {
  __typename?: 'UpdateRolePayload';
  role: Role;
};

export type UpdateUserInput = {
  /** Whole custom attributes to be set on the user. */
  customAttributes?: InputMaybe<Scalars['UserCustomAttributes']['input']>;
  /** Whole standard attributes to be set on the user. */
  standardAttributes?: InputMaybe<Scalars['UserStandardAttributes']['input']>;
  /** Target user ID. */
  userID: Scalars['ID']['input'];
};

export type UpdateUserPayload = {
  __typename?: 'UpdateUserPayload';
  user: User;
};

/** Authgear user */
export type User = Entity & Node & {
  __typename?: 'User';
  /** The scheduled anonymization time of the user */
  anonymizeAt?: Maybe<Scalars['DateTime']['output']>;
  /** The list of authenticators */
  authenticators?: Maybe<AuthenticatorConnection>;
  /** The list of third party app authorizations */
  authorizations?: Maybe<AuthorizationConnection>;
  /** The list of biometric registrations */
  biometricRegistrations: Array<Identity>;
  /** The creation time of entity */
  createdAt: Scalars['DateTime']['output'];
  /** The user's custom attributes */
  customAttributes: Scalars['UserCustomAttributes']['output'];
  /** The scheduled deletion time of the user */
  deleteAt?: Maybe<Scalars['DateTime']['output']>;
  /** The reason of disabled */
  disableReason?: Maybe<Scalars['String']['output']>;
  /** The list of computed roles this user has. */
  effectiveRoles?: Maybe<RoleConnection>;
  /** The end user account id constructed based on user's personal data. (e.g. email, phone...etc) */
  endUserAccountID?: Maybe<Scalars['String']['output']>;
  /** The user's formatted name */
  formattedName?: Maybe<Scalars['String']['output']>;
  /** The list of groups this user has. */
  groups?: Maybe<GroupConnection>;
  /** The ID of an object */
  id: Scalars['ID']['output'];
  /** The list of identities */
  identities?: Maybe<IdentityConnection>;
  /** Indicates if the user is anonymized */
  isAnonymized: Scalars['Boolean']['output'];
  /** Indicates if the user is anonymous */
  isAnonymous: Scalars['Boolean']['output'];
  /** Indicates if the user is deactivated */
  isDeactivated: Scalars['Boolean']['output'];
  /** Indicates if the user is disabled */
  isDisabled: Scalars['Boolean']['output'];
  /** The last login time of user */
  lastLoginAt?: Maybe<Scalars['DateTime']['output']>;
  /** The list of login ids */
  loginIDs: Array<Identity>;
  /** Indicate when will user's MFA grace period will end */
  mfaGracePeriodEndAt?: Maybe<Scalars['DateTime']['output']>;
  /** The list of oauth connections */
  oauthConnections: Array<Identity>;
  /** The list of passkeys */
  passkeys: Array<Identity>;
  /** The primary passwordless via email authenticator */
  primaryOOBOTPEmailAuthenticator?: Maybe<Authenticator>;
  /** The primary passwordless via phone authenticator */
  primaryOOBOTPSMSAuthenticator?: Maybe<Authenticator>;
  /** The primary password authenticator */
  primaryPassword?: Maybe<Authenticator>;
  /** The list of roles this user has. */
  roles?: Maybe<RoleConnection>;
  /** The list of secondary passwordless via email authenticators */
  secondaryOOBOTPEmailAuthenticators: Array<Authenticator>;
  /** The list of secondary passwordless via phone authenticators */
  secondaryOOBOTPSMSAuthenticators: Array<Authenticator>;
  /** The secondary password authenticator */
  secondaryPassword?: Maybe<Authenticator>;
  /** The list of secondary TOTP authenticators */
  secondaryTOTPAuthenticators: Array<Authenticator>;
  /** The list of first party app sessions */
  sessions?: Maybe<SessionConnection>;
  /** The user's standard attributes */
  standardAttributes: Scalars['UserStandardAttributes']['output'];
  /** The update time of entity */
  updatedAt: Scalars['DateTime']['output'];
  /** The list of user's verified claims */
  verifiedClaims: Array<Claim>;
  /** The web3 claims */
  web3: Scalars['Web3Claims']['output'];
};


/** Authgear user */
export type UserAuthenticatorsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  authenticatorKind?: InputMaybe<AuthenticatorKind>;
  authenticatorType?: InputMaybe<AuthenticatorType>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


/** Authgear user */
export type UserAuthorizationsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


/** Authgear user */
export type UserEffectiveRolesArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


/** Authgear user */
export type UserGroupsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


/** Authgear user */
export type UserIdentitiesArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  identityType?: InputMaybe<IdentityType>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


/** Authgear user */
export type UserRolesArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};


/** Authgear user */
export type UserSessionsArgs = {
  after?: InputMaybe<Scalars['String']['input']>;
  before?: InputMaybe<Scalars['String']['input']>;
  first?: InputMaybe<Scalars['Int']['input']>;
  last?: InputMaybe<Scalars['Int']['input']>;
};

/** A connection to a list of items. */
export type UserConnection = {
  __typename?: 'UserConnection';
  /** Information to aid in pagination. */
  edges?: Maybe<Array<Maybe<UserEdge>>>;
  /** Information to aid in pagination. */
  pageInfo: PageInfo;
  /** Total number of nodes in the connection. */
  totalCount?: Maybe<Scalars['Int']['output']>;
};

/** An edge in a connection */
export type UserEdge = {
  __typename?: 'UserEdge';
  /**  cursor for use in pagination */
  cursor: Scalars['String']['output'];
  /** The item at the end of the edge */
  node?: Maybe<User>;
};

export enum UserSortBy {
  CreatedAt = 'CREATED_AT',
  LastLoginAt = 'LAST_LOGIN_AT'
}

export type RemoveMfaGracePeriodInput = {
  /** Target user ID */
  userID: Scalars['ID']['input'];
};

export type RemoveMfaGracePeriodPayload = {
  __typename?: 'removeMFAGracePeriodPayload';
  user: User;
};
