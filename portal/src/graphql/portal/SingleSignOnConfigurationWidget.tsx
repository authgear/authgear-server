import { Checkbox, DirectionalHint, Label, Text, Icon } from "@fluentui/react";
import { Context, FormattedMessage } from "../../intl";
import cn from "classnames";
import { produce } from "immer";
import React, { useCallback, useContext, useMemo, useState } from "react";
import FormTextField from "../../FormTextField";
import ChoiceButton from "../../ChoiceButton";
import {
  createOAuthSSOProviderItemKey,
  isOAuthSSOProvider,
  OAuthSSOFeatureConfig,
  OAuthSSOProviderConfig,
  OAuthSSOProviderFeatureConfig,
  OAuthSSOProviderItemKey,
  OAuthSSOProviderType,
  OAuthSSOWeChatAppType,
  parseOAuthSSOProviderItemKey,
  SSOProviderFormSecretViewModel,
} from "../../types";
import Widget from "../../Widget";
import ExternalLink from "../../ExternalLink";

import FormTextFieldList from "../../FormTextFieldList";
import LabelWithTooltip from "../../LabelWithTooltip";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import styles from "./SingleSignOnConfigurationWidget.module.css";
import ActionButton from "../../ActionButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import {
  OAuthProviderFormModel,
  SSOProviderFormState,
} from "../../hook/useOAuthProviderForm";
import { Badge } from "../../components/v2/Badge/Badge";
import { Callout } from "../../components/v2/Callout/Callout";
import { isOAuthProviderMissingCredential } from "../../model/oauthProviders";
import { EffectiveSecretConfig } from "./globalTypes.generated";

const MASKED_SECRET = "***************";

type CredentialStatus = "active" | "missing_credential" | "demo";

interface SingleSignOnConfigurationWidgetProps {
  className?: string;

  jsonPointer: string;
  clientSecretParentJsonPointer: RegExp;

  isDemoCredentialAvailable: boolean;
  config: OAuthSSOProviderConfig;
  secret: SSOProviderFormSecretViewModel;
  onChange: (
    config: OAuthSSOProviderConfig,
    secret: SSOProviderFormSecretViewModel
  ) => void;

  disabled: boolean;
}

type WidgetTextFieldKey =
  | keyof Omit<OAuthSSOProviderConfig, "type" | "claims">
  | "client_secret"
  | "email_required";

interface OAuthProviderInfo {
  providerType: OAuthSSOProviderType;
  iconClassName: string;
  fields: Set<WidgetTextFieldKey>;
  isSecretFieldTextArea: boolean;
  appType?: OAuthSSOWeChatAppType;
  titleId: string;
  subtitleId?: string;
  descriptionId: string;
  inactiveMessageId: string;
}

const TEXT_FIELD_STYLE = { errorMessage: { whiteSpace: "pre" } };
const MULTILINE_TEXT_FIELD_STYLE = {
  errorMessage: { whiteSpace: "pre" },
  field: { minHeight: "160px" },
};

const oauthProviders: Record<OAuthSSOProviderItemKey, OAuthProviderInfo> = {
  apple: {
    providerType: "apple",
    iconClassName: "fa-apple",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "key_id",
      "team_id",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: true,
    titleId: "AddSingleSignOnConfigurationScreen.card.apple.title",
    descriptionId: "AddSingleSignOnConfigurationScreen.card.apple.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.apple.inactiveMessage",
  },
  google: {
    providerType: "google",
    iconClassName: "fa-google",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.google.title",
    descriptionId: "AddSingleSignOnConfigurationScreen.card.google.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.google.inactiveMessage",
  },
  facebook: {
    providerType: "facebook",
    iconClassName: "fa-facebook",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.facebook.title",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.facebook.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.facebook.inactiveMessage",
  },
  github: {
    providerType: "github",
    iconClassName: "fa-github",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
      "email_required",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.github.title",
    descriptionId: "AddSingleSignOnConfigurationScreen.card.github.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.github.inactiveMessage",
  },
  linkedin: {
    providerType: "linkedin",
    iconClassName: "fa-linkedin",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.linkedin.title",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.linkedin.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.linkedin.inactiveMessage",
  },
  azureadv2: {
    providerType: "azureadv2",
    iconClassName: "fa-microsoft",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "tenant",
      "create_disabled",
      "delete_disabled",
      "email_required",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.azureadv2.title",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.azureadv2.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.azureadv2.inactiveMessage",
  },
  azureadb2c: {
    providerType: "azureadb2c",
    iconClassName: "fa-microsoft",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "tenant",
      "policy",
      "domain_hint",
      "create_disabled",
      "delete_disabled",
      "email_required",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.azureadb2c.title",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.azureadb2c.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.azureadb2c.inactiveMessage",
  },
  adfs: {
    providerType: "adfs",
    iconClassName: "fa-microsoft",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "discovery_document_endpoint",
      "create_disabled",
      "delete_disabled",
      "email_required",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.adfs.title",
    descriptionId: "AddSingleSignOnConfigurationScreen.card.adfs.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.adfs.inactiveMessage",
  },
  "wechat.web": {
    providerType: "wechat",
    appType: "web",
    iconClassName: "fa-weixin",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "account_id",
      "is_sandbox_account",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.wechat.web.title",
    subtitleId: "AddSingleSignOnConfigurationScreen.card.wechat.web.subtitle",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.wechat.web.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.wechat.web.inactiveMessage",
  },
  "wechat.mobile": {
    providerType: "wechat",
    appType: "mobile",
    iconClassName: "fa-weixin",
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "account_id",
      "wechat_redirect_uris",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.wechat.mobile.title",
    subtitleId:
      "AddSingleSignOnConfigurationScreen.card.wechat.mobile.subtitle",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.wechat.mobile.description",
    inactiveMessageId:
      "SingleSignOnConfigurationWidget.providers.wechat.mobile.inactiveMessage",
  },
};

interface OAuthClientIconProps {
  className?: string;
  providerItemKey: OAuthSSOProviderItemKey;
}

const OAuthClientIcon: React.VFC<OAuthClientIconProps> =
  function OAuthClientIcon(props) {
    const { providerItemKey } = props;
    const { iconClassName } = oauthProviders[providerItemKey];
    return <i className={cn("fab", iconClassName, styles.widgetLabelIcon)} />;
  };

function ProviderStatus({
  providerConfig,
  providersWithDemoCredentials,
}: {
  providerConfig: OAuthSSOProviderConfig;
  providersWithDemoCredentials: Set<string>;
}) {
  if (providerConfig.credentials_behavior === "use_demo_credentials") {
    if (providersWithDemoCredentials.has(providerConfig.type)) {
      return (
        <Badge
          size="1"
          variant="warning"
          text={
            <FormattedMessage id="SingleSignOnConfigurationScreen.providerStatus.demo" />
          }
        />
      );
    }
    return (
      <Badge
        size="1"
        variant="error"
        text={
          <FormattedMessage id="SingleSignOnConfigurationScreen.providerStatus.inactive" />
        }
      />
    );
  }
  return (
    <Badge
      size="1"
      variant="success"
      text={
        <FormattedMessage id="SingleSignOnConfigurationScreen.providerStatus.active" />
      }
    />
  );
}

export function useSingleSignOnConfigurationWidget(
  initialAlias: string,
  providerItemKey: OAuthSSOProviderItemKey,
  form: OAuthProviderFormModel,
  effectiveSecretConfig: EffectiveSecretConfig | undefined,
  oauthSSOFeatureConfig?: OAuthSSOFeatureConfig
): SingleSignOnConfigurationWidgetProps {
  const {
    state: { providers },
    setState,
  } = form;

  const [providerType, appType] = parseOAuthSSOProviderItemKey(providerItemKey);

  const [providerIndex] = useState<number>(() => {
    const existingIndex = providers.findIndex((p) =>
      isOAuthSSOProvider(p.config, providerType, initialAlias, appType)
    );
    if (existingIndex !== -1) {
      return existingIndex;
    }
    // Insert at the end if it does not exist
    return providers.length;
  });

  const disabled = useMemo(() => {
    const providersConfig = oauthSSOFeatureConfig?.providers ?? {};
    const providerConfig = providersConfig[
      providerType
    ] as OAuthSSOProviderFeatureConfig | null;
    return providerConfig?.disabled ?? false;
  }, [oauthSSOFeatureConfig?.providers, providerType]);

  const provider = useMemo<SSOProviderFormState>(() => {
    const newConfig = {
      config: {
        type: providerType,
        alias: initialAlias,
        ...(appType && { app_type: appType }),
      },
      secret: {
        originalAlias: null,
        newAlias: initialAlias,
        newClientSecret: "",
      },
    } satisfies SSOProviderFormState;
    return providers.length > providerIndex
      ? providers[providerIndex]
      : newConfig;
  }, [providerType, initialAlias, appType, providers, providerIndex]);

  const jsonPointer = useMemo(() => {
    return `/identity/oauth/providers/${providerIndex}`;
  }, [providerIndex]);
  const clientSecretParentJsonPointer = new RegExp(
    `/secrets/\\d+/data/items/${providerIndex}`
  );

  const providersWithDemoCredentials = useMemo<Set<string>>(() => {
    return new Set(
      effectiveSecretConfig?.oauthSSOProviderDemoSecrets?.map((it) => it.type)
    );
  }, [effectiveSecretConfig]);
  const isDemoCredentialAvailable = providersWithDemoCredentials.has(
    provider.config.type
  );

  const onChange = useCallback(
    (
      newConfig: OAuthSSOProviderConfig,
      secret: SSOProviderFormSecretViewModel
    ) =>
      setState((state) =>
        produce(state, (state) => {
          const config = produce(newConfig, (config) => {
            if (isDemoCredentialAvailable) {
              // If demo credential is avaiable, the user have to choose between demo credential and custom credential
              return;
            }
            // Else, set it automatically
            config.credentials_behavior = isOAuthProviderMissingCredential(
              config,
              secret
            )
              ? "use_demo_credentials"
              : "use_project_credentials";
          });

          if (providerIndex === -1) {
            state.providers.push({
              config,
              secret: {
                originalAlias: null,
                newAlias: secret.newAlias,
                newClientSecret: secret.newClientSecret,
              },
            });
          } else {
            state.providers[providerIndex] = {
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
    [setState, providerIndex, isDemoCredentialAvailable]
  );

  return {
    jsonPointer: jsonPointer,
    clientSecretParentJsonPointer: clientSecretParentJsonPointer,
    isDemoCredentialAvailable: isDemoCredentialAvailable,
    config: provider.config,
    secret: provider.secret,
    onChange: onChange,
    disabled: disabled,
  };
}

// If we do not do this, then some optional config, like domain_hint, when being clear,
// is domain_hint="".
// The JSON schema rejects empty string.
// So when it is an empty string, it should be set to undefined instead.
function emptyStringToUndefined(value: string | undefined): string | undefined {
  if (value == null || value === "") {
    return undefined;
  }
  return value;
}

const SingleSignOnConfigurationWidget: React.VFC<SingleSignOnConfigurationWidgetProps> =
  function SingleSignOnConfigurationWidget(
    props: SingleSignOnConfigurationWidgetProps
  ) {
    const {
      className,
      jsonPointer,
      clientSecretParentJsonPointer,
      isDemoCredentialAvailable,
      config,
      secret,
      onChange,
      disabled: featureDisabled,
    } = props;
    const isMissingCredential =
      config.credentials_behavior === "use_demo_credentials";

    const { renderToString } = useContext(Context);

    const providerItemKey = createOAuthSSOProviderItemKey(
      config.type,
      config.app_type
    );

    const {
      isSecretFieldTextArea,
      fields: visibleFields,
      inactiveMessageId,
    } = oauthProviders[providerItemKey];

    const messageID = "OAuthBranding." + providerItemKey;

    const onAliasChange = useCallback(
      (_, value?: string) =>
        onChange(
          { ...config, alias: value ?? "" },
          { ...secret, newAlias: value ?? "" }
        ),
      [onChange, config, secret]
    );

    const onClientIDChange = useCallback(
      (_, value?: string) =>
        onChange(
          { ...config, client_id: emptyStringToUndefined(value) },
          secret
        ),
      [onChange, config, secret]
    );
    const onTenantChange = useCallback(
      (_, value?: string) =>
        onChange({ ...config, tenant: emptyStringToUndefined(value) }, secret),
      [onChange, config, secret]
    );
    const onPolicyChange = useCallback(
      (_, value?: string) =>
        onChange({ ...config, policy: emptyStringToUndefined(value) }, secret),
      [onChange, config, secret]
    );
    const onDomainHintChange = useCallback(
      (_, value?: string) =>
        onChange(
          { ...config, domain_hint: emptyStringToUndefined(value) },
          secret
        ),
      [onChange, config, secret]
    );
    const onDiscoveryDocumentEndpointChange = useCallback(
      (_, value?: string) =>
        onChange(
          {
            ...config,
            discovery_document_endpoint: emptyStringToUndefined(value),
          },
          secret
        ),
      [onChange, config, secret]
    );
    const onKeyIDChange = useCallback(
      (_, value?: string) =>
        onChange({ ...config, key_id: emptyStringToUndefined(value) }, secret),
      [onChange, config, secret]
    );
    const onTeamIDChange = useCallback(
      (_, value?: string) =>
        onChange({ ...config, team_id: emptyStringToUndefined(value) }, secret),
      [onChange, config, secret]
    );

    const onClientSecretChange = useCallback(
      (_, value?: string) =>
        onChange(config, { ...secret, newClientSecret: value ?? "" }),
      [onChange, config, secret]
    );
    const onAccountIDChange = useCallback(
      (_, value?: string) =>
        onChange(
          { ...config, account_id: emptyStringToUndefined(value) },
          secret
        ),
      [onChange, config, secret]
    );
    const onIsSandBoxAccountChange = useCallback(
      (_, value?: boolean) =>
        onChange({ ...config, is_sandbox_account: value ?? false }, secret),
      [onChange, config, secret]
    );
    const onWeChatRedirectUrisChange = useCallback(
      (list: string[]) =>
        onChange(
          { ...config, wechat_redirect_uris: list.length > 0 ? list : [] },
          secret
        ),
      [onChange, config, secret]
    );
    const onCreateDisabledChange = useCallback(
      (_, value?: boolean) =>
        onChange({ ...config, create_disabled: value ?? false }, secret),
      [onChange, config, secret]
    );
    const onDeleteDisabledChange = useCallback(
      (_, value?: boolean) =>
        onChange({ ...config, delete_disabled: value ?? false }, secret),
      [onChange, config, secret]
    );
    const onEmailRequiredChange = useCallback(
      (_, value?: boolean) => {
        const newConfig = produce(config, (config) => {
          if (value != null) {
            config.claims ??= {};
            config.claims.email ??= {};
            if (!value) {
              config.claims.email.required = false;
            } else {
              delete config.claims.email.required;
            }
          }
        });
        onChange(newConfig, secret);
      },
      [onChange, config, secret]
    );

    const noneditable = featureDisabled;

    const credentialStatus = useMemo<CredentialStatus>(() => {
      if (isMissingCredential && !isDemoCredentialAvailable) {
        return "missing_credential";
      }
      if (isMissingCredential && isDemoCredentialAvailable) {
        return "demo";
      }
      return "active";
    }, [isDemoCredentialAvailable, isMissingCredential]);

    const isDemoCredentialSelected = credentialStatus === "demo";

    const handleDemoCredentialSelectedChange = useCallback(
      (value: boolean) => {
        const newConfig = produce(config, (config) => {
          config.credentials_behavior = value
            ? "use_demo_credentials"
            : "use_project_credentials";
        });
        onChange(newConfig, secret);
      },
      [config, onChange, secret]
    );

    return (
      <Widget className={className}>
        <div className={styles.widgetHeader}>
          <div className={styles.widgetHeaderIcon}>
            <OAuthClientIcon providerItemKey={providerItemKey} />
          </div>
          <Label>{renderToString(messageID)}</Label>
        </div>
        {featureDisabled ? (
          <FeatureDisabledMessageBar messageID="FeatureConfig.disabled" />
        ) : null}
        {isDemoCredentialAvailable ? (
          <div className="grid grid-cols-2 gap-4 grid-flow-col p-px">
            <DemoCredentialStatusButton
              targetValue={false}
              value={isDemoCredentialSelected}
              onClick={handleDemoCredentialSelectedChange}
            />
            <DemoCredentialStatusButton
              targetValue={true}
              value={isDemoCredentialSelected}
              onClick={handleDemoCredentialSelectedChange}
            />
          </div>
        ) : null}
        {credentialStatus === "missing_credential" ? (
          <Callout
            className="w-full"
            type="error"
            text={<FormattedMessage id={inactiveMessageId} />}
            showCloseButton={false}
          />
        ) : credentialStatus === "demo" ? (
          <Callout
            className="w-full"
            type="warning"
            text={
              <FormattedMessage id="SingleSignOnConfigurationWidget.hint.usingDemoCredential" />
            }
            showCloseButton={false}
          />
        ) : null}
        {credentialStatus !== "demo" ? (
          <>
            {visibleFields.has("alias") ? (
              <FormTextField
                parentJSONPointer={jsonPointer}
                fieldName="alias"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.alias"
                )}
                className={styles.textField}
                styles={TEXT_FIELD_STYLE}
                value={config.alias}
                onChange={onAliasChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("client_id") ? (
              <FormTextField
                parentJSONPointer={jsonPointer}
                fieldName="client_id"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.client-id"
                )}
                className={styles.textField}
                styles={TEXT_FIELD_STYLE}
                value={config.client_id ?? ""}
                onChange={onClientIDChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("client_secret") ? (
              <FormTextField
                parentJSONPointer={clientSecretParentJsonPointer}
                fieldName="client_secret"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.client-secret"
                )}
                className={styles.textField}
                styles={
                  isSecretFieldTextArea
                    ? MULTILINE_TEXT_FIELD_STYLE
                    : TEXT_FIELD_STYLE
                }
                multiline={isSecretFieldTextArea}
                value={
                  noneditable || secret.newClientSecret == null
                    ? MASKED_SECRET
                    : secret.newClientSecret
                }
                onChange={onClientSecretChange}
                disabled={noneditable || secret.newClientSecret == null}
              />
            ) : null}
            {visibleFields.has("tenant") ? (
              <FormTextField
                parentJSONPointer={jsonPointer}
                fieldName="tenant"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.tenant"
                )}
                className={styles.textField}
                styles={TEXT_FIELD_STYLE}
                value={config.tenant ?? ""}
                onChange={onTenantChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("policy") ? (
              <FormTextField
                parentJSONPointer={jsonPointer}
                fieldName="policy"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.policy"
                )}
                className={styles.textField}
                styles={TEXT_FIELD_STYLE}
                value={config.policy ?? ""}
                placeholder={renderToString(
                  "SingleSignOnConfigurationScreen.widget.policy.placeholder"
                )}
                onChange={onPolicyChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("domain_hint") ? (
              <FormTextField
                parentJSONPointer={jsonPointer}
                fieldName="domain_hint"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.domain-hint"
                )}
                placeholder={renderToString(
                  "SingleSignOnConfigurationScreen.widget.domain-hint.placeholder"
                )}
                // @ts-expect-error
                description={
                  <FormattedMessage
                    id="SingleSignOnConfigurationScreen.widget.domain-hint.description"
                    values={{
                      externalLink: (chunks: React.ReactNode) => (
                        <ExternalLink
                          href="https://docs.microsoft.com/en-us/azure/active-directory-b2c/direct-signin?pivots=b2c-user-flow#redirect-sign-in-to-a-social-provider"
                          target="_blank"
                          rel="noreferrer"
                        >
                          {chunks}
                        </ExternalLink>
                      ),
                    }}
                  />
                }
                className={styles.textField}
                styles={TEXT_FIELD_STYLE}
                value={config.domain_hint ?? ""}
                onChange={onDomainHintChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("discovery_document_endpoint") ? (
              <FormTextField
                parentJSONPointer={jsonPointer}
                fieldName="discovery_document_endpoint"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.discovery-document-endpoint"
                )}
                className={styles.textField}
                styles={TEXT_FIELD_STYLE}
                value={config.discovery_document_endpoint ?? ""}
                onChange={onDiscoveryDocumentEndpointChange}
                placeholder="http://example.com/.well-known/openid-configuration"
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("key_id") ? (
              <FormTextField
                parentJSONPointer={jsonPointer}
                fieldName="key_id"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.key-id"
                )}
                className={styles.textField}
                styles={TEXT_FIELD_STYLE}
                value={config.key_id ?? ""}
                onChange={onKeyIDChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("team_id") ? (
              <FormTextField
                parentJSONPointer={jsonPointer}
                fieldName="team_id"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.team-id"
                )}
                className={styles.textField}
                styles={TEXT_FIELD_STYLE}
                value={config.team_id ?? ""}
                onChange={onTeamIDChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("account_id") ? (
              <FormTextField
                parentJSONPointer={jsonPointer}
                fieldName="account_id"
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.account-id"
                )}
                className={styles.textField}
                styles={TEXT_FIELD_STYLE}
                value={config.account_id ?? ""}
                onChange={onAccountIDChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("is_sandbox_account") ? (
              <Checkbox
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.is-sandbox-account"
                )}
                className={styles.checkbox}
                checked={config.is_sandbox_account ?? false}
                onChange={onIsSandBoxAccountChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("wechat_redirect_uris") ? (
              <FormTextFieldList
                parentJSONPointer={jsonPointer}
                fieldName="wechat_redirect_uris"
                list={config.wechat_redirect_uris ?? []}
                onListItemChange={onWeChatRedirectUrisChange}
                onListItemAdd={onWeChatRedirectUrisChange}
                onListItemDelete={onWeChatRedirectUrisChange}
                addButtonLabelMessageID="SingleSignOnConfigurationScreen.widget.add-uri"
                className={styles.fieldList}
                label={
                  <LabelWithTooltip
                    labelId="SingleSignOnConfigurationScreen.widget.wechat-redirect-uris-label"
                    tooltipHeaderId="SingleSignOnConfigurationScreen.widget.wechat-redirect-uris-label"
                    tooltipMessageId="SingleSignOnConfigurationScreen.widget.wechat-redirect-uris-tooltip-message"
                    directionalHint={DirectionalHint.bottomLeftEdge}
                  />
                }
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("email_required") ? (
              <Checkbox
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.email-required"
                )}
                className={styles.checkbox}
                checked={config.claims?.email?.required ?? true}
                onChange={onEmailRequiredChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("create_disabled") ? (
              <Checkbox
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.create-disabled"
                )}
                className={styles.checkbox}
                checked={config.create_disabled ?? false}
                onChange={onCreateDisabledChange}
                disabled={noneditable}
              />
            ) : null}
            {visibleFields.has("delete_disabled") ? (
              <Checkbox
                label={renderToString(
                  "SingleSignOnConfigurationScreen.widget.delete-disabled"
                )}
                className={styles.checkbox}
                checked={config.delete_disabled ?? false}
                onChange={onDeleteDisabledChange}
                disabled={noneditable}
              />
            ) : null}
          </>
        ) : null}
      </Widget>
    );
  };

interface OAuthClientCardProps {
  className?: string;
  providerItemKey: OAuthSSOProviderItemKey;
  isAdded?: boolean;
  onAddClick?: (k: OAuthSSOProviderItemKey) => void;
}

function canAddMultiple(provider: OAuthSSOProviderItemKey): boolean {
  switch (provider) {
    case "azureadb2c":
    case "azureadv2":
    case "adfs":
      return true;
    default:
      return false;
  }
}

export const OAuthClientCard: React.VFC<OAuthClientCardProps> =
  function OAuthClientCard(props) {
    const { className, providerItemKey, isAdded, onAddClick } = props;

    const {
      titleId: cardTitleId,
      subtitleId: cardSubtitleId,
      descriptionId: cardDescriptionId,
    } = oauthProviders[providerItemKey];

    const handleAddClick = useCallback(() => {
      onAddClick?.(providerItemKey);
    }, [onAddClick, providerItemKey]);

    return (
      <div className={cn(styles.cardContainer, className)}>
        <div className={styles.cardHeader}>
          <div className={styles.cardTitleRow}>
            <div className={styles.cardIcon}>
              <OAuthClientIcon providerItemKey={providerItemKey} />
            </div>
            <div className={styles.cardName}>
              <Text variant="medium" className={styles.cardTitle}>
                <FormattedMessage id={cardTitleId} />
              </Text>
              {cardSubtitleId != null ? (
                <Text variant="small" className={styles.cardSubtitle}>
                  <FormattedMessage id={cardSubtitleId} />
                </Text>
              ) : null}
            </div>
          </div>
          {isAdded && !canAddMultiple(providerItemKey) ? (
            <div className={styles.cardAddedBadge}>
              <Text variant="small" styles={{ root: { color: "#898989" } }}>
                <FormattedMessage id="AddSingleSignOnConfigurationScreen.card.button.added" />
              </Text>
            </div>
          ) : (
            <ActionButton
              iconProps={{ iconName: "Add" }}
              onClick={handleAddClick}
            />
          )}
        </div>
        <div className={styles.cardBody}>
          <Text variant="small" className={styles.cardDescription}>
            <FormattedMessage id={cardDescriptionId} />
          </Text>
        </div>
      </div>
    );
  };

interface OAuthClientRowProps {
  className?: string;
  providerConfig: OAuthSSOProviderConfig;
  showAlias: boolean;
  providersWithDemoCredentials: Set<string>;
  onEditClick?: (provider: OAuthSSOProviderConfig) => void;
  onDeleteClick?: (provider: OAuthSSOProviderConfig) => void;
}

export const OAuthClientRow: React.VFC<OAuthClientRowProps> =
  function OAuthClientRow(props) {
    const {
      className,
      providerConfig,
      showAlias,
      providersWithDemoCredentials,
      onEditClick,
      onDeleteClick,
    } = props;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();

    const providerItemKey = useMemo(
      () =>
        createOAuthSSOProviderItemKey(
          providerConfig.type,
          providerConfig.app_type
        ),
      [providerConfig]
    );

    const { titleId, subtitleId, descriptionId } =
      oauthProviders[providerItemKey];

    const handleEditClick = useCallback(() => {
      onEditClick?.(providerConfig);
    }, [onEditClick, providerConfig]);

    const handleDeleteClick = useCallback(() => {
      onDeleteClick?.(providerConfig);
    }, [onDeleteClick, providerConfig]);

    return (
      <div className={cn(styles.rowContainer, className)}>
        <div className={styles.rowColumn}>
          <div className={styles.rowIcon}>
            <OAuthClientIcon providerItemKey={providerItemKey} />
          </div>
          <div className={styles.rowContent}>
            <div className={styles.rowName}>
              <Text variant="medium" className={styles.rowTitle} block={true}>
                {`${renderToString(titleId)}${subtitleId != null ? ` (${renderToString(subtitleId)})` : ""
                  }`}
                {showAlias ? ` - ${providerConfig.alias}` : null}
              </Text>
            </div>
            <div className={styles.rowDescription}>
              <Text
                variant="small"
                className={styles.rowDescription}
                block={true}
              >
                <FormattedMessage id={descriptionId} />
              </Text>
            </div>
          </div>
        </div>
        <div className={styles.rowColumn}>
          <ProviderStatus
            providerConfig={providerConfig}
            providersWithDemoCredentials={providersWithDemoCredentials}
          />
        </div>
        <div className={styles.rowActions}>
          <ActionButton
            text={renderToString("SingleSignOnConfigurationScreen.edit")}
            styles={{ label: { fontWeight: 600 } }}
            theme={themes.actionButton}
            onClick={handleEditClick}
          />
          <ActionButton
            text={renderToString("SingleSignOnConfigurationScreen.delete")}
            styles={{ label: { fontWeight: 600 } }}
            theme={themes.destructive}
            onClick={handleDeleteClick}
          />
        </div>
      </div>
    );
  };

export const OAuthClientRowHeader: React.VFC<{ className?: string }> = ({
  className,
}) => {
  return (
    <div className={cn(styles.rowContainer, className)}>
      <div className={styles.rowColumn}>
        <Text variant="medium" className={styles.rowHeader} block={true}>
          <FormattedMessage id="SingleSignOnConfigurationScreen.header.provider" />
        </Text>
      </div>
      <div className={styles.rowColumn}>
        <Text variant="medium" className={styles.rowHeader} block={true}>
          <FormattedMessage id="SingleSignOnConfigurationScreen.header.configuration" />
        </Text>
      </div>
      <div className={styles.rowActions}></div>
    </div>
  );
};

export default SingleSignOnConfigurationWidget;

interface DemoCredentialStatusButtonProps {
  value: boolean;
  targetValue: boolean;
  disabled?: boolean;
  onClick?: (value: boolean) => void;
}

function DemoCredentialStatusButton(props: DemoCredentialStatusButtonProps) {
  const { targetValue, value, disabled, onClick: onClickProp } = props;
  const checked = targetValue === value;

  const { renderToString } = useContext(Context);

  const onClick = useCallback(
    (e) => {
      e.preventDefault();
      e.stopPropagation();
      onClickProp?.(targetValue);
    },
    [onClickProp, targetValue]
  );

  const textID = useMemo(() => {
    switch (targetValue) {
      case false:
        return "SingleSignOnConfigurationWidget.credentialStatusButton.custom.text";
      case true:
        return "SingleSignOnConfigurationWidget.credentialStatusButton.demo.text";
    }
  }, [targetValue]);

  const secondaryTextID = useMemo(() => {
    switch (targetValue) {
      case false:
        return "SingleSignOnConfigurationWidget.credentialStatusButton.custom.secondaryText";
      case true:
        return "SingleSignOnConfigurationWidget.credentialStatusButton.demo.secondaryText";
    }
  }, [targetValue]);

  const IconComponent = useMemo(() => {
    return function IconComponent() {
      const { themes } = useSystemConfig();
      let iconName: string;
      switch (targetValue) {
        case false:
          iconName = "BulletedList";
          break;
        case true:
          iconName = "FavoriteList";
          break;
      }
      return (
        <Icon
          className="mr-4"
          iconName={iconName}
          styles={{
            root: {
              color: themes.main.palette.themePrimary,
              fontSize: "24px",
              lineHeight: "1",
            },
          }}
        />
      );
    };
  }, [targetValue]);

  return (
    <ChoiceButton
      disabled={disabled}
      checked={checked}
      styles={{ root: { paddingLeft: 26 } }}
      text={renderToString(textID)}
      secondaryText={renderToString(secondaryTextID)}
      IconComponent={IconComponent}
      onClick={onClick}
    />
  );
}
