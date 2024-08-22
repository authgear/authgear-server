import React, { useCallback, useContext, useMemo, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { produce } from "immer";
import { MessageBar, MessageBarType, Text } from "@fluentui/react";
import {
  Context as IntlContext,
  FormattedMessage,
} from "@oursky/react-messageformat";
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
  SSOProviderFormSecretViewModel,
} from "../../types";
import styles from "./SingleSignOnConfigurationScreen.module.css";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import { AppSecretKey } from "./globalTypes.generated";
import ActionButton from "../../ActionButton";
import { startReauthentication } from "./Authenticated";

interface LocationState {
  isRevealSecrets: boolean;
}
function isLocationState(raw: unknown): raw is LocationState {
  return (
    raw != null &&
    typeof raw === "object" &&
    (raw as Partial<LocationState>).isRevealSecrets != null
  );
}

interface SSOProviderFormState {
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
          newClientSecret: "",
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
  isEditable: boolean;
}

const OAuthClientItem: React.VFC<OAuthClientItemProps> =
  function OAuthClientItem(props) {
    const {
      providerItemKey,
      form,
      oauthSSOFeatureConfig,
      limitReached,
      isEditable,
    } = props;
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
    }, [oauthSSOFeatureConfig?.providers, providerType]);

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
            originalAlias: null,
            newAlias: defaultAlias(providerType, appType),
            newClientSecret: "",
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
      (
        config: OAuthSSOProviderConfig,
        secret: SSOProviderFormSecretViewModel
      ) =>
        setState((state) =>
          produce(state, (state) => {
            const existingIdx = state.providers.findIndex((p) =>
              isOAuthSSOProvider(p.config, providerType, appType)
            );
            if (existingIdx === -1) {
              state.providers.push({
                config,
                secret: {
                  originalAlias: null,
                  newAlias: secret.newAlias,
                  newClientSecret: secret.newClientSecret,
                },
              });
            } else {
              state.providers[existingIdx] = {
                config,
                secret: {
                  originalAlias: secret.originalAlias,
                  newAlias: secret.newAlias,
                  newClientSecret: secret.newClientSecret,
                },
              };
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
        isEditable={isEditable}
      />
    );
  };

interface SingleSignOnConfigurationContentProps {
  form: AppSecretConfigFormModel<FormState>;
  oauthClientsMaximum: number;
  oauthSSOFeatureConfig?: OAuthSSOFeatureConfig;
}

const SingleSignOnConfigurationContent: React.VFC<SingleSignOnConfigurationContentProps> =
  function SingleSignOnConfigurationContent(props) {
    const { oauthSSOFeatureConfig, oauthClientsMaximum } = props;
    const { state } = props.form;

    const { renderToString } = useContext(IntlContext);

    const limitReached =
      Object.values(state.isEnabled).filter(Boolean).length >=
      oauthClientsMaximum;

    const isEditing = useMemo(() => {
      const isAnySecretPresent =
        state.providers.filter(
          (p) =>
            p.secret.originalAlias != null && p.secret.newClientSecret != null
        ).length !== 0;
      const isNoExistingSecret =
        state.providers.filter((p) => p.secret.originalAlias != null).length ===
        0;
      return isAnySecretPresent || isNoExistingSecret;
    }, [state.providers]);

    const navigate = useNavigate();

    const onRevealSecrets = useCallback(() => {
      const locationState: LocationState = {
        isRevealSecrets: true,
      };

      startReauthentication(navigate, locationState).catch((e) => {
        // Normally there should not be any error.
        console.error(e);
      });
    }, [navigate]);

    return (
      <ScreenContent className={styles.screenContent}>
        <div className={styles.widget}>
          <ActionButton
            iconProps={{ iconName: "Edit" }}
            text={renderToString("SingleSignOnConfigurationScreen.edit")}
            onClick={onRevealSecrets}
            disabled={isEditing}
          />
        </div>
        <ShowOnlyIfSIWEIsDisabled className={styles.widget}>
          {oauthSSOProviderItemKeys.map((providerItemKey) => (
            <OAuthClientItem
              key={providerItemKey}
              providerItemKey={providerItemKey}
              form={props.form}
              oauthSSOFeatureConfig={oauthSSOFeatureConfig}
              limitReached={limitReached}
              isEditable={isEditing}
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

const SingleSignOnConfigurationHeaderContent: React.VFC<{
  children?: React.ReactNode;
}> = function SingleSignOnConfigurationHeaderContent(props) {
  const { children } = props;
  const { appID } = useParams() as { appID: string };
  const featureConfig = useAppFeatureConfigQuery(appID);
  const oauthClientsMaximum = useMemo(
    () =>
      featureConfig.effectiveFeatureConfig?.identity?.oauth
        ?.maximum_providers ?? 99,
    [featureConfig.effectiveFeatureConfig?.identity?.oauth?.maximum_providers]
  );
  return (
    <>
      <div className={styles.headerContent}>
        <ScreenTitle>
          <FormattedMessage id="SingleSignOnConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription>
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
      </div>
      {children}
    </>
  );
};

const SingleSignOnConfigurationScreen1: React.VFC<{
  appID: string;
  secretVisitToken: string | null;
}> = function SingleSignOnConfigurationScreen1({ appID, secretVisitToken }) {
  const config = useAppSecretConfigForm({
    appID,
    secretVisitToken,
    constructFormState,
    constructConfig,
    constructSecretUpdateInstruction,
  });
  const featureConfig = useAppFeatureConfigQuery(appID);

  const form: AppSecretConfigFormModel<FormState> = config;

  const oauthClientsMaximum = useMemo(
    () =>
      featureConfig.effectiveFeatureConfig?.identity?.oauth
        ?.maximum_providers ?? 99,
    [featureConfig.effectiveFeatureConfig?.identity?.oauth?.maximum_providers]
  );

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
    <FormContainer
      form={form}
      HeaderComponent={SingleSignOnConfigurationHeaderContent}
    >
      <SingleSignOnConfigurationContent
        form={form}
        oauthSSOFeatureConfig={
          featureConfig.effectiveFeatureConfig?.identity?.oauth
        }
        oauthClientsMaximum={oauthClientsMaximum}
      />
    </FormContainer>
  );
};

const SECRETS = [AppSecretKey.OauthSsoProviderClientSecrets];

const SingleSignOnConfigurationScreen: React.VFC = () => {
  const { appID } = useParams() as { appID: string };
  const state = useLocationEffect(() => {
    // Pop the state
  });
  const [shouldRefreshToken] = useState<boolean>(() => {
    if (isLocationState(state) && state.isRevealSecrets) {
      return true;
    }
    return false;
  });

  const { token, error, loading, retry } = useAppSecretVisitToken(
    appID,
    SECRETS,
    shouldRefreshToken
  );

  if (error) {
    return <ShowError error={error} onRetry={retry} />;
  }

  if (token === undefined || loading) {
    return <ShowLoading />;
  }

  return (
    <SingleSignOnConfigurationScreen1 appID={appID} secretVisitToken={token} />
  );
};

export default SingleSignOnConfigurationScreen;
