import type { StarterKit } from "./frameworks";

export interface StarterKitEnvValues {
  clientID: string;
  endpoint: string;
}

/** Render the .env file body for a starter kit, substituting live values. */
export function buildEnvFileContent(
  starterKit: StarterKit,
  values: StarterKitEnvValues
): string {
  return starterKit.env
    .map((v) => {
      let value: string;
      switch (v.token) {
        case "clientID":
          value = values.clientID;
          break;
        case "endpoint":
          value = values.endpoint;
          break;
        case "redirectURI":
          value = starterKit.redirectURI;
          break;
      }
      return `${v.key}=${value}`;
    })
    .join("\n");
}

/** Append `uri` to `uris` if not already present (non-destructive dedup). */
export function appendRedirectURI(uris: string[], uri: string): string[] {
  if (uris.includes(uri)) {
    return uris;
  }
  return [...uris, uri];
}
