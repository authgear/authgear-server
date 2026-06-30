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
  // btoa only accepts Latin1. Convert the string to its UTF-8 byte
  // representation first so non-ASCII input (e.g. CJK characters typed into a
  // search box) does not throw InvalidCharacterError. This matches the
  // backend's base64(UTF-8) encoding and leaves ASCII IDs unchanged.
  const utf8Bytes = encodeURIComponent(`${typename}:${rawID}`).replace(
    /%([0-9A-F]{2})/g,
    (_, hex: string) => String.fromCharCode(parseInt(hex, 16))
  );
  return btoa(utf8Bytes)
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=/g, "");
}
