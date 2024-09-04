import { produce } from "immer";
import {
  createOAuthSSOProviderItemKey,
  OAuthSSOProviderClientSecret,
  OAuthSSOProviderConfig,
  OAuthSSOProviderItemKey,
  oauthSSOProviderItemKeys,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
  PortalAPISecretConfigUpdateInstruction,
  SSOProviderFormSecretViewModel,
} from "../types";
import { clearEmptyObject } from "../util/misc";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "./useAppSecretConfigForm";

export interface SSOProviderFormState {
  config: OAuthSSOProviderConfig;
  secret: SSOProviderFormSecretViewModel;
}

interface FormState {
  providers: SSOProviderFormState[];
  isEnabled: Record<OAuthSSOProviderItemKey, boolean>;
  initialProvidersKey: OAuthSSOProviderItemKey[];
}

function constructFormState(
  appConfig: PortalAPIAppConfig,
  secretConfig: PortalAPISecretConfig
): FormState {
  const providerList = appConfig.identity?.oauth?.providers ?? [];
  const secretMap = new Map<string, OAuthSSOProviderClientSecret>();
  for (const item of secretConfig.oauthSSOProviderClientSecrets ?? []) {
    secretMap.set(item.alias, item);
  }

  const providers: SSOProviderFormState[] = [];
  for (const config of providerList) {
    const existingSecret = secretMap.get(config.alias);
    const secretViewModel: SSOProviderFormSecretViewModel = existingSecret
      ? {
          originalAlias: existingSecret.alias,
          newAlias: existingSecret.alias,
          newClientSecret: existingSecret.clientSecret ?? null,
        }
      : {
          originalAlias: config.alias,
          newAlias: config.alias,
          newClientSecret: null,
        };
    providers.push({
      config,
      secret: secretViewModel,
    });
  }

  const isEnabled = {} as Record<OAuthSSOProviderItemKey, boolean>;
  const isOAuthEnabled =
    appConfig.authentication?.identities?.includes("oauth") ?? true;
  for (const itemKey of oauthSSOProviderItemKeys) {
    isEnabled[itemKey] = isOAuthEnabled;
  }

  const initialProvidersKey = providerList.map((x) =>
    createOAuthSSOProviderItemKey(x.type, x.app_type)
  );

  return { providers, isEnabled, initialProvidersKey };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secretConfig: PortalAPISecretConfig,
  initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  // eslint-disable-next-line complexity
  return produce([config, secretConfig], ([config, secretConfig]) => {
    const providers = currentState.providers.filter(
      (p) =>
        currentState.isEnabled[
          createOAuthSSOProviderItemKey(p.config.type, p.config.app_type)
        ]
    );

    const configs: OAuthSSOProviderConfig[] = [];
    const clientSecrets: OAuthSSOProviderClientSecret[] = [];
    for (const p of providers) {
      configs.push(p.config);
      clientSecrets.push({
        alias: p.secret.newAlias,
        clientSecret: p.secret.newClientSecret,
      });
    }

    config.identity ??= {};
    config.identity.oauth ??= {};
    config.identity.oauth.providers = configs;

    secretConfig.oauthSSOProviderClientSecrets = clientSecrets;

    function hasOAuthProviders(s: FormState) {
      return Object.values(s.isEnabled).some(Boolean);
    }
    if (hasOAuthProviders(initialState) !== hasOAuthProviders(currentState)) {
      const identities = (
        effectiveConfig.authentication?.identities ?? []
      ).slice();
      const index = identities.indexOf("oauth");
      const isEnabled = hasOAuthProviders(currentState);

      if (isEnabled && index === -1) {
        identities.push("oauth");
      } else if (!isEnabled && index >= 0) {
        identities.splice(index, 1);
      }
      config.authentication ??= {};
      config.authentication.identities = identities;
    }

    clearEmptyObject(config);
  });
}

function constructSecretUpdateInstruction(
  _config: PortalAPIAppConfig,
  _secretConfig: PortalAPISecretConfig,
  currentState: FormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  return {
    oauthSSOProviderClientSecrets: {
      action: "set",
      data: currentState.providers.map((p) => p.secret),
    },
  };
}

export type OAuthProviderFormModel = AppSecretConfigFormModel<FormState>;

export function useOAuthProviderForm(
  appID: string,
  secretVisitToken: string | null
): OAuthProviderFormModel {
  return useAppSecretConfigForm({
    appID,
    secretVisitToken,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });
}
