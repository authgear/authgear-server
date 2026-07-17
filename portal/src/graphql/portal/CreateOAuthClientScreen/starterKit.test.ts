import { describe, it, expect } from "@jest/globals";
import { buildConfigContent, appendRedirectURI } from "./starterKit";
import type { StarterKit } from "./frameworks";

const KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-react",
  downloadUrl: "https://github.com/authgear/authgear-example-react/archive/HEAD.zip",
  redirectURI: "http://localhost:4000/auth-redirect",
  homepageUrl: "http://localhost:4000",
  config: {
    format: "dotenv",
    fileName: ".env",
    vars: [
      { key: "VITE_AUTHGEAR_CLIENT_ID", token: "clientID" },
      { key: "VITE_AUTHGEAR_ENDPOINT", token: "endpoint" },
      { key: "VITE_AUTHGEAR_REDIRECT_URL", token: "redirectURI" },
    ],
  },
  installCmd: "npm i",
  startCmd: "npm start",
  guideUrl: "https://docs.authgear.com/tutorials/spa/react",
};

const JS_KIT: StarterKit = {
  repoUrl: "https://github.com/authgear/authgear-example-spa-js",
  downloadUrl:
    "https://github.com/authgear/authgear-example-spa-js/archive/HEAD.zip",
  redirectURI: "http://localhost:3000/",
  homepageUrl: "http://localhost:3000",
  config: {
    format: "js",
    fileName: "public/app.js",
    vars: [
      { key: "AUTHGEAR_CLIENT_ID", token: "clientID" },
      { key: "AUTHGEAR_ENDPOINT", token: "endpoint" },
    ],
  },
  installCmd: "npm install",
  startCmd: "npm run dev",
  guideUrl: "https://docs.authgear.com/get-started/single-page-app/website",
};

describe("buildConfigContent", () => {
  it("renders dotenv format with live values in catalog order", () => {
    const result = buildConfigContent(KIT, {
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

  it("renders js format as const assignments", () => {
    const result = buildConfigContent(JS_KIT, {
      clientID: "abc123",
      endpoint: "https://demo.authgear.cloud",
    });
    expect(result).toBe(
      [
        'const AUTHGEAR_CLIENT_ID = "abc123";',
        'const AUTHGEAR_ENDPOINT = "https://demo.authgear.cloud";',
      ].join("\n")
    );
  });

  it("renders literal tokens as their static value", () => {
    const kit: StarterKit = {
      ...KIT,
      config: {
        format: "dotenv",
        fileName: ".env.local",
        vars: [
          { key: "AUTHGEAR_CLIENT_ID", token: "clientID" },
          {
            key: "SESSION_SECRET",
            token: "literal",
            literalValue: "a-random-string",
          },
        ],
      },
    };
    const result = buildConfigContent(kit, {
      clientID: "abc123",
      endpoint: "https://demo.authgear.cloud",
    });
    expect(result).toBe(
      ["AUTHGEAR_CLIENT_ID=abc123", "SESSION_SECRET=a-random-string"].join("\n")
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
