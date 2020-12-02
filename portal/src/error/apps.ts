export interface APIDuplicatedAppIDError {
  errorName: "AlreadyExists";
  reason: "DuplicatedAppID";
}

export interface APIReservedAppIDError {
  errorName: "Forbidden";
  reason: "AppIDReserved";
}

export interface APIInvalidAppIDError {
  errorName: "Invalid";
  reason: "InvalidAppID";
}
