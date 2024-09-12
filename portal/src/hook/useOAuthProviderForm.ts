import { produce } from "immer";
import {
  createOAuthSSOProviderItemKey,
  OAuthSSOProviderClientSecret,
  OAuthSSOProviderConfig,
  OAuthSSOProviderItemKey,
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

  const initialProvidersKey = providerList.map((x) =>
    createOAuthSSOProviderItemKey(x.type, x.app_type)
  );

  return { providers, initialProvidersKey };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secretConfig: PortalAPISecretConfig,
  _initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  // eslint-disable-next-line complexity
  return produce([config, secretConfig], ([config, secretConfig]) => {
    const configs: OAuthSSOProviderConfig[] = [];
    const clientSecrets: OAuthSSOProviderClientSecret[] = [];
    for (const p of currentState.providers) {
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

    const identities = (
      effectiveConfig.authentication?.identities ?? []
    ).slice();
    const oauthIdentityIndex = identities.indexOf("oauth");

    if (currentState.providers.length > 0 && oauthIdentityIndex === -1) {
      identities.push("oauth");
    }
    if (currentState.providers.length === 0) {
      identities.splice(oauthIdentityIndex, 1);
    }

    config.authentication ??= {};
    config.authentication.identities = identities;

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
