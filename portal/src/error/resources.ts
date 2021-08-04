export interface APIResourceNotFoundError {
  errorName: "ResourceNotFound";
  reason: "ResourceNotFound";
}

export interface APIResourceTooLargeError {
  errorName: string;
  reason: "ResourceTooLarge";
  info: {
    size: number;
    max_size: number;
    path: string;
  };
}
