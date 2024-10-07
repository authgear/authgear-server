export function extractRawID(id: string): string {
  // eslint-disable-next-line @typescript-eslint/no-deprecated
  const decoded = atob(id);
  const parts = decoded.split(":");
  if (parts.length !== 2) {
    throw new Error("invalid graphql ID: " + decoded);
  }
  return parts[1];
}

export function toTypedID(typename: string, rawID: string): string {
  // eslint-disable-next-line @typescript-eslint/no-deprecated
  return btoa(`${typename}:${rawID}`)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}
