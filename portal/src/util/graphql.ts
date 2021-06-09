export function extractRawID(id: string): string {
  const decoded = atob(id);
  const parts = decoded.split(":");
  if (parts.length !== 2) {
    throw new Error("invalid graphql ID: " + decoded);
  }
  return parts[1];
}
