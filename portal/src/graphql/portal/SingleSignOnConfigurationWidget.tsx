import React, { useCallback, useContext, useMemo, useState } from "react";
import cn from "classnames";
import { produce } from "immer";
import { Checkbox, DirectionalHint } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import Widget from "../../Widget";
import FormTextField from "../../FormTextField";
import {
  createOAuthSSOProviderItemKey,
  OAuthSSOProviderConfig,
  OAuthSSOProviderItemKey,
  OAuthSSOProviderType,
  OAuthSSOWeChatAppType,
  SSOProviderFormSecretViewModel,
} from "../../types";

import styles from "./SingleSignOnConfigurationWidget.module.css";
import FormTextFieldList from "../../FormTextFieldList";
import LabelWithTooltip from "../../LabelWithTooltip";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import Toggle from "../../Toggle";

interface WidgetHeaderLabelProps {
  icon: React.ReactNode;
  messageID: string;
}

interface WidgetHeaderProps extends WidgetHeaderLabelProps {
  checked: boolean;
  onChange: (value: boolean) => void;
  disabled: boolean;
}

interface SingleSignOnConfigurationWidgetProps {
  className?: string;

  jsonPointer: string;
  clientSecretParentJsonPointer: RegExp;

  isEnabled: boolean;
  onIsEnabledChange: (value: boolean) => void;

  config: OAuthSSOProviderConfig;
  secret: SSOProviderFormSecretViewModel;
  onChange: (
    config: OAuthSSOProviderConfig,
    secret: SSOProviderFormSecretViewModel
  ) => void;

  disabled: boolean;
  limitReached: boolean;
  isEditable: boolean;
}

type WidgetTextFieldKey =
  | keyof Omit<OAuthSSOProviderConfig, "type" | "claims">
  | "client_secret"
  | "email_required";

interface OAuthProviderInfo {
  providerType: OAuthSSOProviderType;
  iconNode: React.ReactNode;
  fields: Set<WidgetTextFieldKey>;
  isSecretFieldTextArea: boolean;
  appType?: OAuthSSOWeChatAppType;
}

const TEXT_FIELD_STYLE = { errorMessage: { whiteSpace: "pre" } };
const MULTILINE_TEXT_FIELD_STYLE = {
  errorMessage: { whiteSpace: "pre" },
  field: { minHeight: "160px" },
};

const oauthProviders: Record<OAuthSSOProviderItemKey, OAuthProviderInfo> = {
  apple: {
    providerType: "apple",
    iconNode: <i className={cn("fab", "fa-apple", styles.widgetLabelIcon)} />,
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
  },
  google: {
    providerType: "google",
    iconNode: <i className={cn("fab", "fa-google", styles.widgetLabelIcon)} />,
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
  },
  facebook: {
    providerType: "facebook",
    iconNode: (
      <i className={cn("fab", "fa-facebook", styles.widgetLabelIcon)} />
    ),
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
  },
  github: {
    providerType: "github",
    iconNode: <i className={cn("fab", "fa-github", styles.widgetLabelIcon)} />,
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
      "email_required",
    ]),
    isSecretFieldTextArea: false,
  },
  linkedin: {
    providerType: "linkedin",
    iconNode: (
      <i className={cn("fab", "fa-linkedin", styles.widgetLabelIcon)} />
    ),
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "create_disabled",
      "delete_disabled",
    ]),
    isSecretFieldTextArea: false,
  },
  azureadv2: {
    providerType: "azureadv2",
    iconNode: (
      <i className={cn("fab", "fa-microsoft", styles.widgetLabelIcon)} />
    ),
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
  },
  azureadb2c: {
    providerType: "azureadb2c",
    iconNode: (
      <i className={cn("fab", "fa-microsoft", styles.widgetLabelIcon)} />
    ),
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
  },
  adfs: {
    providerType: "adfs",
    iconNode: (
      <i className={cn("fab", "fa-microsoft", styles.widgetLabelIcon)} />
    ),
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
  },
  "wechat.web": {
    providerType: "wechat",
    appType: "web",
    iconNode: <i className={cn("fab", "fa-weixin", styles.widgetLabelIcon)} />,
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
  },
  "wechat.mobile": {
    providerType: "wechat",
    appType: "web",
    iconNode: <i className={cn("fab", "fa-weixin", styles.widgetLabelIcon)} />,
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
  },
};

const WidgetHeader: React.VFC<WidgetHeaderProps> = function WidgetHeader(
  props: WidgetHeaderProps
) {
  const { icon, messageID, checked, onChange: onChangeProp, disabled } = props;

  const onChange = useCallback(
    (_, value?: boolean) => {
      onChangeProp(value ?? false);
    },
    [onChangeProp]
  );

  return (
    <Toggle
      checked={checked}
      onChange={onChange}
      inlineLabel={true}
      label={
        <>
          {icon}
          <FormattedMessage id={messageID} />
        </>
      }
      disabled={disabled}
    />
  );
};

const SingleSignOnConfigurationWidget: React.VFC<SingleSignOnConfigurationWidgetProps> =
  // eslint-disable-next-line complexity
  function SingleSignOnConfigurationWidget(
    props: SingleSignOnConfigurationWidgetProps
  ) {
    const {
      className,
      jsonPointer,
      clientSecretParentJsonPointer,
      isEnabled,
      onIsEnabledChange,
      config,
      secret,
      onChange,
      disabled: featureDisabled,
      limitReached,
      isEditable,
    } = props;

    const { renderToString } = useContext(Context);

    const [extended, setExtended] = useState(isEnabled);

    const onToggleButtonClick = useCallback(() => {
      setExtended((prev) => {
        return !prev;
      });
    }, []);

    const providerItemKey = createOAuthSSOProviderItemKey(
      config.type,
      config.app_type
    );

    const {
      isSecretFieldTextArea,
      iconNode,
      fields: visibleFields,
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

    const disabledByLimitReached = useMemo(() => {
      return !isEnabled && limitReached;
    }, [limitReached, isEnabled]);

    const noneditable =
      !isEnabled || featureDisabled || disabledByLimitReached || !isEditable;

    const masking = isEnabled ? "***************" : "";

    const canToggle = useMemo(() => {
      if (!isEditable) {
        // Not in edit mode, no toggle possible.
        return false;
      }
      // Can only turn off if limit reached or feature disabled
      if (featureDisabled || disabledByLimitReached) {
        return isEnabled;
      }
      return true;
    }, [featureDisabled, disabledByLimitReached, isEditable, isEnabled]);

    return (
      <Widget
        className={className}
        extended={isEnabled || extended}
        showToggleButton={true}
        toggleButtonDisabled={isEnabled}
        onToggleButtonClick={onToggleButtonClick}
      >
        <WidgetHeader
          icon={iconNode}
          checked={isEnabled}
          onChange={onIsEnabledChange}
          messageID={messageID}
          disabled={!canToggle}
        />
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
                ? masking
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

export default SingleSignOnConfigurationWidget;
