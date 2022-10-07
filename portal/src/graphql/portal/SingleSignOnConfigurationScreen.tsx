import React, { useCallback, useMemo } from "react";
import { useParams } from "react-router-dom";
import { produce } from "immer";
import { MessageBar, MessageBarType, Text } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";
import SingleSignOnConfigurationWidget from "./SingleSignOnConfigurationWidget";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import ShowOnlyIfSIWEIsDisabled from "./ShowOnlyIfSIWEIsDisabled";
import { clearEmptyObject } from "../../util/misc";
import FormContainer from "../../FormContainer";
import {
  AppSecretConfigFormModel,
  useAppSecretConfigForm,
} from "../../hook/useAppSecretConfigForm";
import {
  createOAuthSSOProviderItemKey,
  isOAuthSSOProvider,
  OAuthSSOProviderClientSecret,
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
  PortalAPISecretConfigUpdateInstruction,
} from "../../types";
import styles from "./SingleSignOnConfigurationScreen.module.css";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import {
  AuthgearGTMEvent,
  AuthgearGTMEventType,
  useAuthgearGTMEventBase,
  useGTMDispatch,
} from "../../GTMProvider";

interface SSOProviderFormState {
  config: OAuthSSOProviderConfig;
  secret: OAuthSSOProviderClientSecret;
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
      clientSecrets.push(p.secret);
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
  secretConfig: PortalAPISecretConfig,
  _currentState: FormState
): PortalAPISecretConfigUpdateInstruction | undefined {
  return {
    oauthSSOProviderClientSecrets: {
      action: "set",
      data:
        secretConfig.oauthSSOProviderClientSecrets?.map((s) => {
          return {
            alias: s.alias,
            clientSecret: s.clientSecret,
          };
        }) ?? [],
    },
  };
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

const OAuthClientItem: React.VFC<OAuthClientItemProps> =
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
    const jsonPointer = useMemo(() => {
      return index >= 0 ? `/identity/oauth/providers/${index}` : "";
    }, [index]);
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
      (config: OAuthSSOProviderConfig, secret: OAuthSSOProviderClientSecret) =>
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

const SingleSignOnConfigurationContent: React.VFC<SingleSignOnConfigurationContentProps> =
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
        <ShowOnlyIfSIWEIsDisabled className={styles.widget}>
          <ScreenDescription className={styles.widget}>
            <Text className={styles.description} block={true}>
              <FormattedMessage id="SingleSignOnConfigurationScreen.description" />
            </Text>
            {oauthClientsMaximum < 99 ? (
              <FeatureDisabledMessageBar
                messageID="FeatureConfig.sso.maximum"
                messageValues={{
                  maximum: oauthClientsMaximum,
                }}
              />
            ) : null}
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
          <MessageBar
            messageBarType={MessageBarType.info}
            className={styles.widget}
          >
            <FormattedMessage id="SingleSignOnConfigurationScreen.whatsapp-otp-doc.message" />
          </MessageBar>
        </ShowOnlyIfSIWEIsDisabled>
      </ScreenContent>
    );
  };

const SingleSignOnConfigurationScreen: React.VFC =
  function SingleSignOnConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const config = useAppSecretConfigForm({
      appID,
      constructFormState,
      constructConfig,
      constructSecretUpdateInstruction,
    });

    const featureConfig = useAppFeatureConfigQuery(appID);

    const sendDataToGTM = useGTMDispatch();
    const gtmEventBase = useAuthgearGTMEventBase();
    const save = useCallback(async () => {
      // compare if there is any newly added providers
      // then send the gtm event
      const initialProvidersKey = config.state.initialProvidersKey;
      const currentProvidersKey = config.state.providers
        .map((p) =>
          createOAuthSSOProviderItemKey(p.config.type, p.config.app_type)
        )
        .filter((key) => config.state.isEnabled[key]);
      const addedProviders = currentProvidersKey.filter(
        (t) => !initialProvidersKey.includes(t)
      );

      await config.save();
      if (addedProviders.length > 0) {
        const event: AuthgearGTMEvent = {
          ...gtmEventBase,
          event: AuthgearGTMEventType.AddedSSOProviders,
          event_data: {
            providers: addedProviders,
          },
        };
        sendDataToGTM(event);
      }
    }, [config, gtmEventBase, sendDataToGTM]);

    const form: AppSecretConfigFormModel<FormState> = {
      ...config,
      save,
    };

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
