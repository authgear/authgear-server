export interface APIDuplicatedCollaboratorInvitationError {
  errorName: "AlreadyExists";
  reason: "CollaboratorInvitationDuplicate";
}

export interface APICollaboratorSelfDeletionError {
  errorName: "Forbidden";
  reason: "CollaboratorSelfDeletion";
}
