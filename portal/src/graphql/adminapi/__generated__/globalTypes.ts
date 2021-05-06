/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

//==============================================================
// START Enums and Input Objects
//==============================================================

export enum AuthenticatorKind {
  PRIMARY = "PRIMARY",
  SECONDARY = "SECONDARY",
}

export enum AuthenticatorType {
  OOB_OTP_EMAIL = "OOB_OTP_EMAIL",
  OOB_OTP_SMS = "OOB_OTP_SMS",
  PASSWORD = "PASSWORD",
  TOTP = "TOTP",
}

export enum IdentityType {
  ANONYMOUS = "ANONYMOUS",
  BIOMETRIC = "BIOMETRIC",
  LOGIN_ID = "LOGIN_ID",
  OAUTH = "OAUTH",
}

export enum SearchUsersSortBy {
  CREATED_AT = "CREATED_AT",
  LAST_LOGIN_AT = "LAST_LOGIN_AT",
}

export enum SessionType {
  IDP = "IDP",
  OFFLINE_GRANT = "OFFLINE_GRANT",
}

export enum SortDirection {
  ASC = "ASC",
  DESC = "DESC",
}

/**
 * Definition of an identity. This is a union object, exactly one of the available fields must be present.
 */
export interface IdentityDefinition {
  loginID?: IdentityDefinitionLoginID | null;
}

export interface IdentityDefinitionLoginID {
  key: string;
  value: string;
}

//==============================================================
// END Enums and Input Objects
//==============================================================
