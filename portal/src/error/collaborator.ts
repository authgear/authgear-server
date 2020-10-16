export interface APIDuplicatedCollaboratorInvitationError {
  errorName: "AlreadyExists";
  reason: "CollaboratorInvitationDuplicate";
}

export interface APICollaboratorSelfDeletionError {
  errorName: "Forbidden";
  reason: "CollaboratorSelfDeletion";
}

export interface APICollaboratorInvitationInvalidCodeError {
  errorName: "Invalid";
  reason: "CollaboratorInvitationInvalidCode";
}

export interface APICollaboratorDuplicateError {
  errorName: "AlreadyExists";
  reason: "CollaboratorDuplicate";
}
