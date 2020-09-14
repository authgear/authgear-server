import React, { useCallback, useMemo, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Link, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

import SingleSignOnConfigurationWidget from "./SingleSignOnConfigurationWidget";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ButtonWithLoading from "../../ButtonWithLoading";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import { useUpdateAppAndSecretConfigMutation } from "./mutations/updateAppAndSecretMutation";
import { clearEmptyObject, ensureNonEmptyString } from "../../util/misc";
import { parseError } from "../../util/error";
import { Violation } from "../../util/validation";
import {
  OAuthClientCredentialItem,
  OAuthSSOProviderConfig,
  OAuthSSOProviderType,
  oauthSSOProviderTypes,
  PortalAPIApp,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
} from "../../types";

import styles from "./SingleSignOnConfigurationScreen.module.scss";

interface SingleSignOnConfigurationProps {
  rawAppConfig: PortalAPIAppConfig | null;
  effectiveAppConfig: PortalAPIAppConfig | null;
  secretConfig: PortalAPISecretConfig | null;
  updatingAppConfig: boolean;
  updateAppConfig: (
    appConfig: PortalAPIAppConfig,
    secretConfig: PortalAPISecretConfig
  ) => Promise<PortalAPIApp | null>;
  updateAppConfigError: unknown;
}

export interface OAuthSSOProviderConfigState extends OAuthSSOProviderConfig {
  enabled: boolean;
  clientSecret?: string;
}

export type SingleSignOnScreenState = {
  [key in OAuthSSOProviderType]: OAuthSSOProviderConfigState;
};

function getProviderConfigStateFromProviders(
  providerType: OAuthSSOProviderType,
  providerByType: Partial<Record<OAuthSSOProviderType, OAuthSSOProviderConfig>>,
  clientSecretByAlias: Partial<Record<string, string>>
): OAuthSSOProviderConfigState {
  const providerConfig = providerByType[providerType] ?? {
    type: providerType,
  };
  // FIXME: make alias required
  const alias = providerConfig.alias ?? providerType;
  return {
    ...providerConfig,
    enabled: !!providerByType[providerType],
    clientSecret: clientSecretByAlias[alias],
  };
}

function extractOAuthSecret(secretConfig: PortalAPISecretConfig) {
  for (const secret of secretConfig.secrets) {
    if (secret.key === "sso.oauth.client") {
      return secret;
    }
  }
  return null;
}

function constructScreenState(
  providers: OAuthSSOProviderConfig[],
  oauthCredentials: OAuthClientCredentialItem[]
) {
  const providerByType = providers.reduce(
    (
      map: Partial<Record<OAuthSSOProviderType, OAuthSSOProviderConfig>>,
      provider
    ) => {
      map[provider.type] = provider;
      return map;
    },
    {}
  );

  const clientSecretByAlias = oauthCredentials.reduce(
    (
      map: Partial<Record<string, string>>,
      credential: OAuthClientCredentialItem
    ) => {
      if (credential.alias !== "") {
        map[credential.alias] = credential.client_secret;
      }
      return map;
    },
    {}
  );

  const screenState = oauthSSOProviderTypes.reduce(
    (
      map: Partial<
        { [key in OAuthSSOProviderType]: OAuthSSOProviderConfigState }
      >,
      providerType
    ) => {
      map[providerType] = getProviderConfigStateFromProviders(
        providerType,
        providerByType,
        clientSecretByAlias
      );
      return map;
    },
    {}
  );

  return screenState as SingleSignOnScreenState;
}

function constructProvidersFromState(state: SingleSignOnScreenState) {
  const providers: OAuthSSOProviderConfig[] = [];
  for (const providerType of oauthSSOProviderTypes) {
    const providerConfigState = state[providerType];
    if (!providerConfigState.enabled) {
      continue;
    }
    providers.push({
      type: providerConfigState.type,
      alias: ensureNonEmptyString(providerConfigState.alias),
      client_id: ensureNonEmptyString(providerConfigState.client_id),
      tenant: ensureNonEmptyString(providerConfigState.tenant),
      key_id: ensureNonEmptyString(providerConfigState.key_id),
      team_id: ensureNonEmptyString(providerConfigState.team_id),
    });
  }

  return providers;
}

function constructOAuthCredentialsFromState(state: SingleSignOnScreenState) {
  const credentials: OAuthClientCredentialItem[] = [];

  oauthSSOProviderTypes.forEach((provider) => {
    const { enabled, clientSecret, alias } = state[provider];
    if (enabled && clientSecret != null && clientSecret.trim() !== "") {
      credentials.push({
        // FIXME: make alias required
        alias: alias ?? provider,
        client_secret: clientSecret,
      });
    }
  });

  return credentials;
}

function constructViolationMap(
  violations: Violation[],
  providers: OAuthSSOProviderType[]
) {
  const map: Partial<Record<OAuthSSOProviderType, Violation[]>> = {};
  const unhandledViolation: Violation[] = [];
  for (const violation of violations) {
    // general violation has no location -> not handled
    const locationPrefix = "/identity/oauth/providers";
    if (!violation.location.startsWith(locationPrefix)) {
      unhandledViolation.push(violation);
      continue;
    }
    // if the error lies in identity.oauth.providers
    // expect last segment to be integer
    const indexStr = violation.location.split("/").pop();
    const index = parseInt(indexStr ?? "", 10);
    if (isNaN(index) || index < 0 || index >= providers.length) {
      // not recognized or out of range
      unhandledViolation.push(violation);
      continue;
    }
    const targetProvider = providers[index];
    map[targetProvider] = map[targetProvider] ?? [];
    map[targetProvider]?.push(violation);
  }
  return { map, unhandledViolation };
}

function determineIsAllProviderDisabled(state: SingleSignOnScreenState) {
  for (const providerType of oauthSSOProviderTypes) {
    if (state[providerType].enabled) {
      return false;
    }
  }
  return true;
}

const SingleSignOnConfiguration: React.FC<SingleSignOnConfigurationProps> = function SingleSignOnConfiguration(
  props: SingleSignOnConfigurationProps
) {
  const {
    rawAppConfig,
    effectiveAppConfig,
    secretConfig,
    updateAppConfig,
    updatingAppConfig,
    updateAppConfigError,
  } = props;

  const providers = useMemo(() => {
    return effectiveAppConfig?.identity?.oauth?.providers ?? [];
  }, [effectiveAppConfig]);

  const oauthCredentials = useMemo(() => {
    if (secretConfig == null) {
      return [];
    }
    const oauthSecret = extractOAuthSecret(secretConfig);
    return oauthSecret?.data.items ?? [];
  }, [secretConfig]);

  const initialState: SingleSignOnScreenState = useMemo(() => {
    return constructScreenState(providers, oauthCredentials);
  }, [providers, oauthCredentials]);

  const [state, setState] = useState(initialState);
  const [unhandledViolation, setUnhandleViolation] = useState<Violation[]>([]);

  const enabledProviderTypesRef = useRef<OAuthSSOProviderType[]>([]);

  const onSaveClick = useCallback(() => {
    if (rawAppConfig == null || secretConfig == null) {
      return;
    }

    const newProviders = constructProvidersFromState(state);
    enabledProviderTypesRef.current = newProviders.map(
      (provider) => provider.type
    );
    const newOAuthCredentials = constructOAuthCredentialsFromState(state);

    const newAppConfig = produce(rawAppConfig, (draftConfig) => {
      draftConfig.identity = draftConfig.identity ?? {};
      draftConfig.identity.oauth = draftConfig.identity.oauth ?? {};
      const { oauth } = draftConfig.identity;

      const isAllProviderDisabled = determineIsAllProviderDisabled(state);
      if (isAllProviderDisabled) {
        delete oauth.providers;
        clearEmptyObject(draftConfig);
        return;
      }

      oauth.providers = newProviders;

      clearEmptyObject(draftConfig);
    });

    const newSecretConfig = produce(secretConfig, (draftConfig) => {
      const oauthSecret = extractOAuthSecret(draftConfig);

      if (oauthSecret == null) {
        if (newOAuthCredentials.length > 0) {
          draftConfig.secrets.push({
            key: "sso.oauth.client",
            data: {
              items: newOAuthCredentials,
            },
          });
        }
      } else {
        if (newOAuthCredentials.length > 0) {
          oauthSecret.data.items = newOAuthCredentials;
        } else {
          const index = draftConfig.secrets.findIndex(
            (secret) => secret.key === "sso.oauth.client"
          );
          if (index >= 0) {
            draftConfig.secrets.splice(index, 1);
          }
        }
      }
    });

    updateAppConfig(newAppConfig, newSecretConfig).catch(() => {});
  }, [state, rawAppConfig, secretConfig, updateAppConfig]);

  const violationMap = useMemo(() => {
    if (updateAppConfigError == null) {
      setUnhandleViolation([]);
      return {};
    }
    const providers = enabledProviderTypesRef.current;
    const violations = parseError(updateAppConfigError);
    const {
      map,
      unhandledViolation: _unhandledViolation,
    } = constructViolationMap(violations, providers);
    setUnhandleViolation(_unhandledViolation);
    return map;
  }, [updateAppConfigError]);

  return (
    <section className={styles.screenContent}>
      {unhandledViolation.length > 0 && (
        <div className={styles.error}>
          <ShowError error={updateAppConfigError} />
        </div>
      )}
      <SingleSignOnConfigurationWidget
        className={styles.widget}
        serviceProviderType={"apple"}
        screenState={state}
        setScreenState={setState}
        violations={violationMap["apple"]}
      />
      <ButtonWithLoading
        className={styles.saveButton}
        loading={updatingAppConfig}
        labelId="save"
        loadingLabelId="saving"
        onClick={onSaveClick}
      />
    </section>
  );
};

const SingleSignOnConfigurationScreen: React.FC = function SingleSignOnConfigurationScreen() {
  const { appID } = useParams();
  const { data, loading, error, refetch } = useAppAndSecretConfigQuery(appID);
  const {
    updateAppAndSecretConfig,
    loading: updatingAppAndSecretConfig,
    error: updateAppAndSecretConfigError,
  } = useUpdateAppAndSecretConfigMutation(appID);

  if (loading) {
    return <ShowLoading />;
  }

  if (error != null) {
    return <ShowError error={error} onRetry={refetch} />;
  }

  const rawAppConfig =
    data?.node?.__typename === "App" ? data.node.rawAppConfig : null;
  const effectiveAppConfig =
    data?.node?.__typename === "App" ? data.node.effectiveAppConfig : null;
  const secretConfig =
    data?.node?.__typename === "App" ? data.node.rawSecretConfig : null;

  return (
    <main className={styles.root} role="main">
      <Text as="h1" className={styles.header}>
        <FormattedMessage id="SingleSignOnConfigurationScreen.title" />
      </Text>
      <Link href="#" className={styles.helpLink}>
        <FormattedMessage id="SingleSignOnConfigurationScreen.help-link" />
      </Link>
      <SingleSignOnConfiguration
        rawAppConfig={rawAppConfig}
        effectiveAppConfig={effectiveAppConfig}
        secretConfig={secretConfig}
        updatingAppConfig={updatingAppAndSecretConfig}
        updateAppConfig={updateAppAndSecretConfig}
        updateAppConfigError={updateAppAndSecretConfigError}
      />
    </main>
  );
};

export default SingleSignOnConfigurationScreen;
