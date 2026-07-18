import type { StarterKit } from "./frameworks";

export interface StarterKitConfigValues {
  clientID: string;
  endpoint: string;
}

/** Render the config block for a starter kit, substituting live values. */
export function buildConfigContent(
  starterKit: StarterKit,
  values: StarterKitConfigValues
): string {
  const { config } = starterKit;
  return config.vars
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
          value = starterKit.redirectURIs[0] ?? "";
          break;
        case "literal":
          value = v.literalValue ?? "";
          break;
      }
      if (config.format === "js") {
        return `const ${v.key} = "${value}";`;
      }
      if (config.format === "swift") {
        return `static let ${v.key} = "${value}"`;
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
