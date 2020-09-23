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
  OOB_OTP = "OOB_OTP",
  PASSWORD = "PASSWORD",
  TOTP = "TOTP",
}

export enum IdentityType {
  ANONYMOUS = "ANONYMOUS",
  LOGIN_ID = "LOGIN_ID",
  OAUTH = "OAUTH",
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
