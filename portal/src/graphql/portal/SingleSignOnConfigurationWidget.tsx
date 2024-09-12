import { Checkbox, DirectionalHint, Label, Text } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import cn from "classnames";
import { produce } from "immer";
import React, { useCallback, useContext, useMemo } from "react";
import FormTextField from "../../FormTextField";
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

const MASKED_SECRET = "***************";

interface SingleSignOnConfigurationWidgetProps {
  className?: string;

  jsonPointer: string;
  clientSecretParentJsonPointer: RegExp;

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
      "create_disabled",
      "delete_disabled",
      "email_required",
    ]),
    isSecretFieldTextArea: false,
    titleId: "AddSingleSignOnConfigurationScreen.card.azureadb2c.title",
    descriptionId:
      "AddSingleSignOnConfigurationScreen.card.azureadb2c.description",
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

function defaultAlias(
  providerType: OAuthSSOProviderType,
  appType?: OAuthSSOWeChatAppType
) {
  return appType ? [providerType, appType].join("_") : providerType;
}

export function useSingleSignOnConfigurationWidget(
  providerItemKey: OAuthSSOProviderItemKey,
  form: OAuthProviderFormModel,
  oauthSSOFeatureConfig?: OAuthSSOFeatureConfig
): SingleSignOnConfigurationWidgetProps {
  const {
    state: { providers },
    setState,
  } = form;

  const [providerType, appType] = parseOAuthSSOProviderItemKey(providerItemKey);

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

  const index = providers.findIndex((p) =>
    isOAuthSSOProvider(p.config, providerType, appType)
  );
  const jsonPointer = useMemo(() => {
    return index >= 0 ? `/identity/oauth/providers/${index}` : "";
  }, [index]);
  const clientSecretParentJsonPointer =
    index >= 0
      ? new RegExp(`/secrets/\\d+/data/items/${index}`)
      : /placeholder/;

  const onChange = useCallback(
    (config: OAuthSSOProviderConfig, secret: SSOProviderFormSecretViewModel) =>
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

  return {
    jsonPointer: jsonPointer,
    clientSecretParentJsonPointer: clientSecretParentJsonPointer,
    config: provider.config,
    secret: provider.secret,
    onChange: onChange,
    disabled: disabled,
  };
}

const SingleSignOnConfigurationWidget: React.VFC<SingleSignOnConfigurationWidgetProps> =
  // eslint-disable-next-line complexity
  function SingleSignOnConfigurationWidget(
    props: SingleSignOnConfigurationWidgetProps
  ) {
    const {
      className,
      jsonPointer,
      clientSecretParentJsonPointer,
      config,
      secret,
      onChange,
      disabled: featureDisabled,
    } = props;

    const { renderToString } = useContext(Context);

    const providerItemKey = createOAuthSSOProviderItemKey(
      config.type,
      config.app_type
    );

    const { isSecretFieldTextArea, fields: visibleFields } =
      oauthProviders[providerItemKey];

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
        onChange({ ...config, client_id: value ?? "" }, secret),
      [onChange, config, secret]
    );
    const onTenantChange = useCallback(
      (_, value?: string) =>
        onChange({ ...config, tenant: value ?? "" }, secret),
      [onChange, config, secret]
    );
    const onPolicyChange = useCallback(
      (_, value?: string) =>
        onChange({ ...config, policy: value ?? "" }, secret),
      [onChange, config, secret]
    );
    const onDiscoveryDocumentEndpointChange = useCallback(
      (_, value?: string) =>
        onChange(
          { ...config, discovery_document_endpoint: value ?? "" },
          secret
        ),
      [onChange, config, secret]
    );
    const onKeyIDChange = useCallback(
      (_, value?: string) =>
        onChange({ ...config, key_id: value ?? "" }, secret),
      [onChange, config, secret]
    );
    const onTeamIDChange = useCallback(
      (_, value?: string) =>
        onChange({ ...config, team_id: value ?? "" }, secret),
      [onChange, config, secret]
    );

    const onClientSecretChange = useCallback(
      (_, value?: string) =>
        onChange(config, { ...secret, newClientSecret: value ?? "" }),
      [onChange, config, secret]
    );
    const onAccountIDChange = useCallback(
      (_, value?: string) =>
        onChange({ ...config, account_id: value ?? "" }, secret),
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
      </Widget>
    );
  };

interface OAuthClientCardProps {
  className?: string;
  providerItemKey: OAuthSSOProviderItemKey;
  isAdded?: boolean;
  onAddClick?: (k: OAuthSSOProviderItemKey) => void;
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
          {isAdded ? (
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
  providerItemKey: OAuthSSOProviderItemKey;
  onEditClick?: (k: OAuthSSOProviderItemKey) => void;
  onDeleteClick?: (k: OAuthSSOProviderItemKey) => void;
}

export const OAuthClientRow: React.VFC<OAuthClientRowProps> =
  function OAuthClientRow(props) {
    const { className, providerItemKey, onEditClick, onDeleteClick } = props;
    const { renderToString } = useContext(Context);
    const { themes } = useSystemConfig();

    const { titleId, subtitleId, descriptionId } =
      oauthProviders[providerItemKey];

    const handleEditClick = useCallback(() => {
      onEditClick?.(providerItemKey);
    }, [onEditClick, providerItemKey]);

    const handleDeleteClick = useCallback(() => {
      onDeleteClick?.(providerItemKey);
    }, [onDeleteClick, providerItemKey]);

    return (
      <div className={cn(styles.rowContainer, className)}>
        <div className={styles.rowIcon}>
          <OAuthClientIcon providerItemKey={providerItemKey} />
        </div>
        <div className={styles.rowContent}>
          <div className={styles.rowName}>
            <Text variant="medium" className={styles.rowTitle}>
              {`${renderToString(titleId)}${
                subtitleId != null ? ` (${renderToString(subtitleId)})` : ""
              }`}
            </Text>
          </div>
          <div className={styles.rowDescription}>
            <Text variant="small" className={styles.rowDescription}>
              <FormattedMessage id={descriptionId} />
            </Text>
          </div>
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

export default SingleSignOnConfigurationWidget;
