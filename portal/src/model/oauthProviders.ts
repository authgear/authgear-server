import {
  OAuthSSOProviderConfig,
  SSOProviderFormSecretViewModel,
} from "../types";

export function deriveOAuthProviderDisabled(
  config: OAuthSSOProviderConfig,
  secret: SSOProviderFormSecretViewModel
): boolean {
  const hasClientSecret =
    secret.newClientSecret != null && secret.newClientSecret !== "";

  switch (config.type) {
    case "apple":
      // Apple requires client_id, key_id, team_id and client_secret
      return !(
        config.client_id &&
        config.key_id &&
        config.team_id &&
        hasClientSecret
      );

    case "google":
    case "facebook":
    case "github":
    case "linkedin":
      // These providers require client_id and client_secret
      return !(config.client_id && hasClientSecret);

    case "azureadv2":
      // Azure AD v2 requires client_id, tenant and client_secret
      return !(config.client_id && config.tenant && hasClientSecret);

    case "azureadb2c":
      // Azure AD B2C requires client_id, tenant, policy and client_secret
      return !(
        config.client_id &&
        config.tenant &&
        config.policy &&
        hasClientSecret
      );

    case "adfs":
      // ADFS requires client_id, discovery_document_endpoint and client_secret
      return !(
        config.client_id &&
        config.discovery_document_endpoint &&
        hasClientSecret
      );

    case "wechat":
      // WeChat requires client_id, client_secret and account_id
      return !(config.client_id && config.account_id && hasClientSecret);
  }
}
