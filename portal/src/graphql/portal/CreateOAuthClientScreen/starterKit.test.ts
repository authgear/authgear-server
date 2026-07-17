import { describe, it, expect } from "@jest/globals";
import { buildEnvFileContent, appendRedirectURI } from "./starterKit";
import type { StarterKit } from "./frameworks";

const KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-react",
  downloadUrl:
    "https://github.com/authgear/authgear-example-react/archive/refs/heads/main.zip",
  redirectURI: "http://localhost:4000/auth-redirect",
  homepageUrl: "http://localhost:4000",
  env: [
    { key: "VITE_AUTHGEAR_CLIENT_ID", token: "clientID" },
    { key: "VITE_AUTHGEAR_ENDPOINT", token: "endpoint" },
    { key: "VITE_AUTHGEAR_REDIRECT_URL", token: "redirectURI" },
  ],
  installCmd: "npm i",
  startCmd: "npm start",
  guideUrl: "https://docs.authgear.com/tutorials/spa/react",
};

describe("buildEnvFileContent", () => {
  it("substitutes live values in catalog order", () => {
    const result = buildEnvFileContent(KIT, {
      clientID: "abc123",
      endpoint: "https://demo.authgear.cloud",
    });
    expect(result).toBe(
      [
        "VITE_AUTHGEAR_CLIENT_ID=abc123",
        "VITE_AUTHGEAR_ENDPOINT=https://demo.authgear.cloud",
        "VITE_AUTHGEAR_REDIRECT_URL=http://localhost:4000/auth-redirect",
      ].join("\n")
    );
  });
});

describe("appendRedirectURI", () => {
  it("appends when absent", () => {
    expect(
      appendRedirectURI(["http://localhost/after-authentication"], KIT.redirectURI)
    ).toEqual([
      "http://localhost/after-authentication",
      "http://localhost:4000/auth-redirect",
    ]);
  });

  it("is a no-op when already present", () => {
    const uris = [KIT.redirectURI];
    expect(appendRedirectURI(uris, KIT.redirectURI)).toEqual([KIT.redirectURI]);
  });

  it("preserves existing URIs", () => {
    const uris = ["a", "b"];
    expect(appendRedirectURI(uris, "c")).toEqual(["a", "b", "c"]);
  });
});
