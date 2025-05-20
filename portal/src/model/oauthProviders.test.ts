import { isOAuthProviderMissingCredential } from "./oauthProviders";
import {
  OAuthSSOProviderConfig,
  SSOProviderFormSecretViewModel,
} from "../types";

describe("isOAuthProviderMissingCredential", () => {
  const createConfig = (
    type: OAuthSSOProviderConfig["type"],
    overrides: Partial<OAuthSSOProviderConfig> = {}
  ): OAuthSSOProviderConfig => ({
    type,
    alias: "test",
    ...overrides,
  });

  const createSecret = (
    overrides: Partial<SSOProviderFormSecretViewModel> = {}
  ): SSOProviderFormSecretViewModel => ({
    originalAlias: null,
    newAlias: "test",
    newClientSecret: "",
    ...overrides,
  });

  describe("Apple provider", () => {
    it("should be true when missing required fields", () => {
      const config = createConfig("apple");
      const secret = createSecret();
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(true);
    });

    it("should be false when all required fields are present", () => {
      const config = createConfig("apple", {
        client_id: "client-id",
        key_id: "key-id",
        team_id: "team-id",
      });
      const secret = createSecret({ newClientSecret: "secret" });
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(false);
    });
  });

  describe("Google provider", () => {
    it("should be true when missing required fields", () => {
      const config = createConfig("google");
      const secret = createSecret();
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(true);
    });

    it("should be false when all required fields are present", () => {
      const config = createConfig("google", {
        client_id: "client-id",
      });
      const secret = createSecret({ newClientSecret: "secret" });
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(false);
    });
  });

  describe("Facebook provider", () => {
    it("should be true when missing required fields", () => {
      const config = createConfig("facebook");
      const secret = createSecret();
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(true);
    });

    it("should be false when all required fields are present", () => {
      const config = createConfig("facebook", {
        client_id: "client-id",
      });
      const secret = createSecret({ newClientSecret: "secret" });
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(false);
    });
  });

  describe("GitHub provider", () => {
    it("should be true when missing required fields", () => {
      const config = createConfig("github");
      const secret = createSecret();
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(true);
    });

    it("should be false when all required fields are present", () => {
      const config = createConfig("github", {
        client_id: "client-id",
      });
      const secret = createSecret({ newClientSecret: "secret" });
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(false);
    });
  });

  describe("LinkedIn provider", () => {
    it("should be true when missing required fields", () => {
      const config = createConfig("linkedin");
      const secret = createSecret();
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(true);
    });

    it("should be false when all required fields are present", () => {
      const config = createConfig("linkedin", {
        client_id: "client-id",
      });
      const secret = createSecret({ newClientSecret: "secret" });
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(false);
    });
  });

  describe("Azure AD v2 provider", () => {
    it("should be true when missing required fields", () => {
      const config = createConfig("azureadv2");
      const secret = createSecret();
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(true);
    });

    it("should be false when all required fields are present", () => {
      const config = createConfig("azureadv2", {
        client_id: "client-id",
        tenant: "tenant-id",
      });
      const secret = createSecret({ newClientSecret: "secret" });
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(false);
    });
  });

  describe("Azure AD B2C provider", () => {
    it("should be true when missing required fields", () => {
      const config = createConfig("azureadb2c");
      const secret = createSecret();
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(true);
    });

    it("should be false when all required fields are present", () => {
      const config = createConfig("azureadb2c", {
        client_id: "client-id",
        tenant: "tenant-id",
        policy: "policy-id",
      });
      const secret = createSecret({ newClientSecret: "secret" });
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(false);
    });
  });

  describe("ADFS provider", () => {
    it("should be true when missing required fields", () => {
      const config = createConfig("adfs");
      const secret = createSecret();
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(true);
    });

    it("should be false when all required fields are present", () => {
      const config = createConfig("adfs", {
        client_id: "client-id",
        discovery_document_endpoint: "https://example.com",
      });
      const secret = createSecret({ newClientSecret: "secret" });
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(false);
    });
  });

  describe("WeChat provider", () => {
    it("should be true when missing required fields", () => {
      const config = createConfig("wechat");
      const secret = createSecret();
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(true);
    });

    it("should be false when all required fields are present", () => {
      const config = createConfig("wechat", {
        client_id: "client-id",
        account_id: "account-id",
      });
      const secret = createSecret({ newClientSecret: "secret" });
      expect(isOAuthProviderMissingCredential(config, secret)).toBe(false);
    });
  });
});
