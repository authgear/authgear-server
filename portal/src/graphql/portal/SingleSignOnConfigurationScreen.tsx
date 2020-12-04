import React, { useCallback, useContext, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { Link } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import SingleSignOnConfigurationWidget from "./SingleSignOnConfigurationWidget";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import { clearEmptyObject } from "../../util/misc";
import {
  OAuthClientCredentialItem,
  OAuthSecretItem,
  OAuthSSOProviderConfig,
  OAuthSSOProviderType,
  oauthSSOProviderTypes,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
} from "../../types";

import styles from "./SingleSignOnConfigurationScreen.module.scss";
import NavBreadcrumb, { BreadcrumbItem } from "../../NavBreadcrumb";
import FormContainer from "../../FormContainer";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";

interface SSOProviderFormState {
  config: OAuthSSOProviderConfig;
  secret: OAuthClientCredentialItem;
}

interface FormState {
  providers: SSOProviderFormState[];
  isEnabled: Record<OAuthSSOProviderType, boolean>;
}

function constructFormState(
  appConfig: PortalAPIAppConfig,
  secretConfig: PortalAPISecretConfig
): FormState {
  const providerList = appConfig.identity?.oauth?.providers ?? [];
  const secretMap = new Map<string, OAuthClientCredentialItem>();
  for (const item of secretConfig.secrets) {
    if (item.key === "sso.oauth.client") {
      for (const clientSecret of item.data.items) {
        secretMap.set(clientSecret.alias, clientSecret);
      }
      break;
    }
  }

  const providers: SSOProviderFormState[] = [];
  for (const config of providerList) {
    providers.push({
      config,
      secret: secretMap.get(config.alias) ?? {
        alias: config.alias,
        client_secret: "",
      },
    });
  }

  const isEnabled = {} as Record<OAuthSSOProviderType, boolean>;
  const isOAuthEnabled =
    appConfig.authentication?.identities?.includes("oauth") ?? true;
  for (const type of oauthSSOProviderTypes) {
    isEnabled[type] =
      isOAuthEnabled && providers.some((p) => p.config.type === type);
  }

  return { providers, isEnabled };
}

function constructConfig(
  config: PortalAPIAppConfig,
  secrets: PortalAPISecretConfig,
  initialState: FormState,
  currentState: FormState,
  effectiveConfig: PortalAPIAppConfig
): [PortalAPIAppConfig, PortalAPISecretConfig] {
  // eslint-disable-next-line complexity
  return produce([config, secrets], ([config, { secrets }]) => {
    const providers = currentState.providers.filter(
      (p) => currentState.isEnabled[p.config.type]
    );

    const configs: OAuthSSOProviderConfig[] = [];
    const clientSecrets: OAuthClientCredentialItem[] = [];
    for (const p of providers) {
      configs.push(p.config);
      clientSecrets.push(p.secret);
    }

    config.identity ??= {};
    config.identity.oauth ??= {};
    config.identity.oauth.providers = configs;

    const secretItem: OAuthSecretItem = {
      key: "sso.oauth.client",
      data: { items: clientSecrets },
    };

    const secretIndex = secrets.findIndex((s) => s.key === "sso.oauth.client");
    if (clientSecrets.length === 0) {
      if (secretIndex >= 0) {
        secrets.splice(secretIndex, 1);
      }
    } else if (secretIndex >= 0) {
      secrets[secretIndex] = secretItem;
    } else {
      secrets.push(secretItem);
    }

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

interface OAuthClientItemProps {
  providerType: OAuthSSOProviderType;
  form: AppSecretConfigFormModel<FormState>;
}

const OAuthClientItem: React.FC<OAuthClientItemProps> = function OAuthClientItem(
  props
) {
  const { providerType, form } = props;
  const {
    state: { providers, isEnabled },
    setState,
  } = form;

  const provider = useMemo(
    () =>
      providers.find((p) => p.config.type === providerType) ?? {
        config: {
          type: providerType,
          alias: providerType,
        },
        secret: { alias: providerType, client_secret: "" },
      },
    [providers, providerType]
  );

  const enabledProviders = providers.filter((p) => isEnabled[p.config.type]);
  const index = enabledProviders.findIndex(
    (p) => p.config.type === providerType
  );
  const jsonPointer = index >= 0 ? `/identity/oauth/providers/${index}` : "";
  const clientSecretParentJsonPointer =
    index >= 0 ? RegExp(`^/secrets/[0-9]+/data/items/${index}$`) : "";

  const onIsEnabledChange = useCallback(
    (isEnabled: boolean) => {
      setState((state) =>
        produce(state, (state) => {
          state.isEnabled[providerType] = isEnabled;
          const hasProvider = state.providers.some(
            (p) => p.config.type === providerType
          );
          if (isEnabled && !hasProvider) {
            state.providers.push(provider);
          }
        })
      );
    },
    [setState, providerType, provider]
  );

  const onChange = useCallback(
    (config: OAuthSSOProviderConfig, secret: OAuthClientCredentialItem) =>
      setState((state) =>
        produce(state, (state) => {
          const index = state.providers.findIndex(
            (p) => p.config.type === providerType
          );
          if (index === -1) {
            state.providers.push({ config, secret });
          } else if (index >= 0) {
            state.providers[index] = { config, secret };
          }
        })
      ),
    [setState, providerType]
  );

  return (
    <SingleSignOnConfigurationWidget
      className={styles.widget}
      jsonPointer={jsonPointer}
      clientSecretParentJsonPointer={clientSecretParentJsonPointer}
      isEnabled={isEnabled[providerType]}
      onIsEnabledChange={onIsEnabledChange}
      config={provider.config}
      secret={provider.secret}
      onChange={onChange}
    />
  );
};

interface SingleSignOnConfigurationContentProps {
  form: AppSecretConfigFormModel<FormState>;
}

const SingleSignOnConfigurationContent: React.FC<SingleSignOnConfigurationContentProps> = function SingleSignOnConfigurationContent(
  props
) {
  const { renderToString } = useContext(Context);

  const navBreadcrumbItems: BreadcrumbItem[] = useMemo(() => {
    return [
      {
        to: ".",
        label: <FormattedMessage id="SingleSignOnConfigurationScreen.title" />,
      },
    ];
  }, []);

  return (
    <div className={styles.root}>
      <NavBreadcrumb items={navBreadcrumbItems} />
      <Link
        href={renderToString("SingleSignOnConfigurationScreen.help-link-href")}
        target="_blank"
        className={styles.helpLink}
      >
        <FormattedMessage id="SingleSignOnConfigurationScreen.help-link-label" />
      </Link>

      {oauthSSOProviderTypes.map((providerType) => (
        <OAuthClientItem
          key={providerType}
          providerType={providerType}
          form={props.form}
        />
      ))}
    </div>
  );
};

const SingleSignOnConfigurationScreen: React.FC = function SingleSignOnConfigurationScreen() {
  const { appID } = useParams();

  const form = useAppSecretConfigForm(
    appID,
    constructFormState,
    constructConfig
  );

  if (form.isLoading) {
    return <ShowLoading />;
  }

  if (form.loadError) {
    return <ShowError error={form.loadError} onRetry={form.reload} />;
  }

  return (
    <FormContainer form={form}>
      <SingleSignOnConfigurationContent form={form} />
    </FormContainer>
  );
};

export default SingleSignOnConfigurationScreen;
