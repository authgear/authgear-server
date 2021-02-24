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
  LOGIN_ID = "LOGIN_ID",
  OAUTH = "OAUTH",
}

export enum SessionType {
  IDP = "IDP",
  OFFLINE_GRANT = "OFFLINE_GRANT",
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
