import { NodeType } from "../graphql/adminapi/node";

export function extractRawID(id: string): string {
  const decoded = atob(id);
  const parts = decoded.split(":");
  if (parts.length !== 2) {
    throw new Error("invalid graphql ID: " + decoded);
  }
  return parts[1];
}

export function toTypedID(typename: NodeType | "App", rawID: string): string {
  return btoa(`${typename}:${rawID}`)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}
