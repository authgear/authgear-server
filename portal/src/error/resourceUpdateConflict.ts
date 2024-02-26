export interface APIResourceUpdateConflictError {
  errorName: string;
  reason: "ResourceUpdateConflict";
  info: {
    path: string;
  };
}
