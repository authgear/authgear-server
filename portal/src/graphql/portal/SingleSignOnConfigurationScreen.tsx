import React, { useMemo } from "react";
import { useParams } from "react-router-dom";
import { Link, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

import SingleSignOnConfigurationWidget from "./SingleSignOnConfigurationWidget";
import {
  OAuthClientCredentialItem,
  OAuthSSOProviderConfig,
  OAuthSSOProviderType,
  oauthSSOProviderTypes,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
} from "../../types";

import styles from "./SingleSignOnConfigurationScreen.module.scss";
import { useAppAndSecretConfigQuery } from "./query/appAndSecretConfigQuery";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";

interface SingleSignOnConfigurationProps {
  rawAppConfig: PortalAPIAppConfig | null;
  effectiveAppConfig: PortalAPIAppConfig | null;
  secretConfig: PortalAPISecretConfig | null;
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
    client_id: "",
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

const SingleSignOnConfiguration: React.FC<SingleSignOnConfigurationProps> = function SingleSignOnConfiguration(
  props: SingleSignOnConfigurationProps
) {
  const { effectiveAppConfig, secretConfig } = props;

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

  const [state, setState] = React.useState(initialState);

  return (
    <section className={styles.screenContent}>
      <SingleSignOnConfigurationWidget
        className={styles.widget}
        serviceProviderType="apple"
        screenState={state}
        setScreenState={setState}
      />
    </section>
  );
};

const SingleSignOnConfigurationScreen: React.FC = function SingleSignOnConfigurationScreen() {
  const { appID } = useParams();
  const { data, loading, error, refetch } = useAppAndSecretConfigQuery(appID);

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
      />
    </main>
  );
};

export default SingleSignOnConfigurationScreen;
