import React, { useCallback, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { MessageBar, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import SingleSignOnConfigurationWidget from "./SingleSignOnConfigurationWidget";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import { clearEmptyObject } from "../../util/misc";
import FormContainer from "../../FormContainer";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import {
  createOAuthSSOProviderItemKey,
  isOAuthSSOProvider,
  OAuthClientSecret,
  OAuthSSOFeatureConfig,
  OAuthSSOProviderConfig,
  OAuthSSOProviderFeatureConfig,
  OAuthSSOProviderItemKey,
  oauthSSOProviderItemKeys,
  OAuthSSOProviderType,
  OAuthSSOWeChatAppType,
  parseOAuthSSOProviderItemKey,
  PortalAPIAppConfig,
  PortalAPISecretConfig,
} from "../../types";
import styles from "./SingleSignOnConfigurationScreen.module.scss";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";

interface SSOProviderFormState {
  config: OAuthSSOProviderConfig;
  secret: OAuthClientSecret;
}

interface FormState {
  providers: SSOProviderFormState[];
  isEnabled: Record<OAuthSSOProviderItemKey, boolean>;
}

function constructFormState(
  appConfig: PortalAPIAppConfig,
  secretConfig: PortalAPISecretConfig
): FormState {
  const providerList = appConfig.identity?.oauth?.providers ?? [];
  const secretMap = new Map<string, OAuthClientSecret>();
  for (const item of secretConfig.oauthClientSecrets ?? []) {
    secretMap.set(item.alias, item);
  }

  const providers: SSOProviderFormState[] = [];
  for (const config of providerList) {
    providers.push({
      config,
      secret: secretMap.get(config.alias) ?? {
        alias: config.alias,
        clientSecret: "",
      },
    });
  }

  const isEnabled = {} as Record<OAuthSSOProviderItemKey, boolean>;
  const isOAuthEnabled =
    appConfig.authentication?.identities?.includes("oauth") ?? true;
  for (const itemKey of oauthSSOProviderItemKeys) {
    isEnabled[itemKey] =
      isOAuthEnabled &&
      providers.some(
        (p) =>
          createOAuthSSOProviderItemKey(p.config.type, p.config.app_type) ===
          itemKey
      );
  }

  return { providers, isEnabled };
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
    const clientSecrets: OAuthClientSecret[] = [];
    for (const p of providers) {
      configs.push(p.config);
      clientSecrets.push(p.secret);
    }

    config.identity ??= {};
    config.identity.oauth ??= {};
    config.identity.oauth.providers = configs;

    secretConfig.oauthClientSecrets = clientSecrets;

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

function defaultAlias(
  providerType: OAuthSSOProviderType,
  appType?: OAuthSSOWeChatAppType
) {
  return appType ? [providerType, appType].join("_") : providerType;
}

interface OAuthClientItemProps {
  providerItemKey: OAuthSSOProviderItemKey;
  form: AppSecretConfigFormModel<FormState>;
  oauthSSOFeatureConfig?: OAuthSSOFeatureConfig;
  limitReached: boolean;
}

const OAuthClientItem: React.FC<OAuthClientItemProps> =
  function OAuthClientItem(props) {
    const { providerItemKey, form, oauthSSOFeatureConfig, limitReached } =
      props;
    const {
      state: { providers, isEnabled },
      setState,
    } = form;

    const [providerType, appType] =
      parseOAuthSSOProviderItemKey(providerItemKey);

    const disabled = useMemo(() => {
      const providersConfig = oauthSSOFeatureConfig?.providers ?? {};
      const providerConfig = providersConfig[
        providerType
      ] as OAuthSSOProviderFeatureConfig | null;
      return providerConfig?.disabled ?? false;
    }, [oauthSSOFeatureConfig, providerType]);

    const provider = useMemo<SSOProviderFormState>(
      () =>
        providers.find((p) =>
          isOAuthSSOProvider(p.config, providerType, appType)
        ) ?? {
          config: {
            type: providerType,
            alias: defaultAlias(providerType, appType),
            ...(appType && { app_type: appType }),
          },
          secret: {
            alias: defaultAlias(providerType, appType),
            clientSecret: "",
          },
        },
      [providers, providerType, appType]
    );

    const enabledProviders = providers.filter(
      (p) =>
        isEnabled[
          createOAuthSSOProviderItemKey(p.config.type, p.config.app_type)
        ]
    );
    const index = enabledProviders.findIndex((p) =>
      isOAuthSSOProvider(p.config, providerType, appType)
    );
    const jsonPointer = index >= 0 ? `/identity/oauth/providers/${index}` : "";
    const clientSecretParentJsonPointer =
      index >= 0
        ? new RegExp(`/secrets/\\d+/data/items/${index}`)
        : /placeholder/;

    const onIsEnabledChange = useCallback(
      (isEnabled: boolean) => {
        setState((state) =>
          produce(state, (state) => {
            state.isEnabled[
              createOAuthSSOProviderItemKey(providerType, appType)
            ] = isEnabled;
            const hasProvider = state.providers.some((p) =>
              isOAuthSSOProvider(p.config, providerType, appType)
            );
            if (isEnabled && !hasProvider) {
              state.providers.push(provider);
            }
          })
        );
      },
      [setState, providerType, appType, provider]
    );

    const onChange = useCallback(
      (config: OAuthSSOProviderConfig, secret: OAuthClientSecret) =>
        setState((state) =>
          produce(state, (state) => {
            const index = state.providers.findIndex((p) =>
              isOAuthSSOProvider(p.config, providerType, appType)
            );
            if (index === -1) {
              state.providers.push({ config, secret });
            } else if (index >= 0) {
              state.providers[index] = { config, secret };
            }
          })
        ),
      [setState, providerType, appType]
    );

    return (
      <SingleSignOnConfigurationWidget
        className={styles.widget}
        jsonPointer={jsonPointer}
        clientSecretParentJsonPointer={clientSecretParentJsonPointer}
        isEnabled={
          isEnabled[createOAuthSSOProviderItemKey(providerType, appType)]
        }
        onIsEnabledChange={onIsEnabledChange}
        config={provider.config}
        secret={provider.secret}
        onChange={onChange}
        disabled={disabled}
        limitReached={limitReached}
      />
    );
  };

interface SingleSignOnConfigurationContentProps {
  form: AppSecretConfigFormModel<FormState>;
  oauthSSOFeatureConfig?: OAuthSSOFeatureConfig;
}

const SingleSignOnConfigurationContent: React.FC<SingleSignOnConfigurationContentProps> =
  function SingleSignOnConfigurationContent(props) {
    const { oauthSSOFeatureConfig } = props;
    const { state } = props.form;

    const oauthClientsMaximum = useMemo(
      () => oauthSSOFeatureConfig?.maximum_providers ?? 99,
      [oauthSSOFeatureConfig?.maximum_providers]
    );

    const limitReached =
      Object.values(state.isEnabled).filter(Boolean).length >=
      oauthClientsMaximum;

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="SingleSignOnConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <Text className={styles.description} block={true}>
            <FormattedMessage
              id="SingleSignOnConfigurationScreen.description"
              values={{
                HREF: "https://docs.authgear.com/strategies/how-to-setup-sso-integrations",
              }}
            />
          </Text>
          {oauthClientsMaximum < 99 && (
            <MessageBar>
              <FormattedMessage
                id="FeatureConfig.sso.maximum"
                values={{
                  planPagePath: "../../billing",
                  maximum: oauthClientsMaximum,
                }}
              />
            </MessageBar>
          )}
        </ScreenDescription>
        {oauthSSOProviderItemKeys.map((providerItemKey) => (
          <OAuthClientItem
            key={providerItemKey}
            providerItemKey={providerItemKey}
            form={props.form}
            oauthSSOFeatureConfig={props.oauthSSOFeatureConfig}
            limitReached={limitReached}
          />
        ))}
      </ScreenContent>
    );
  };

const SingleSignOnConfigurationScreen: React.FC =
  function SingleSignOnConfigurationScreen() {
    const { appID } = useParams();

    const form = useAppSecretConfigForm(
      appID,
      constructFormState,
      constructConfig
    );

    const featureConfig = useAppFeatureConfigQuery(appID);

    if (form.isLoading || featureConfig.loading) {
      return <ShowLoading />;
    }

    if (form.loadError ?? featureConfig.error) {
      return (
        <ShowError
          error={form.loadError ?? featureConfig.error}
          onRetry={() => {
            form.reload();
            featureConfig.refetch().finally(() => {});
          }}
        />
      );
    }

    return (
      <FormContainer form={form}>
        <SingleSignOnConfigurationContent
          form={form}
          oauthSSOFeatureConfig={
            featureConfig.effectiveFeatureConfig?.identity?.oauth
          }
        />
      </FormContainer>
    );
  };

export default SingleSignOnConfigurationScreen;
