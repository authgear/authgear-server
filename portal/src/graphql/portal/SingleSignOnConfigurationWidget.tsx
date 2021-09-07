import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import cn from "classnames";
import { Checkbox, DirectionalHint, Toggle, MessageBar } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import FormTextField from "../../FormTextField";
import {
  createOAuthSSOProviderItemKey,
  OAuthClientSecret,
  OAuthSSOProviderConfig,
  OAuthSSOProviderItemKey,
  OAuthSSOProviderType,
  OAuthSSOWeChatAppType,
} from "../../types";

import styles from "./SingleSignOnConfigurationWidget.module.scss";
import FormTextFieldList from "../../FormTextFieldList";
import LabelWithTooltip from "../../LabelWithTooltip";

interface WidgetHeaderLabelProps {
  icon: React.ReactNode;
  messageID: string;
}

interface WidgetHeaderProps extends WidgetHeaderLabelProps {
  isEnabled: boolean;
  setIsEnabled: (value: boolean) => void;
  disabled: boolean;
  disabledByLimitReached: boolean;
}

interface SingleSignOnConfigurationWidgetProps {
  className?: string;

  jsonPointer: string;
  clientSecretParentJsonPointer: string;

  isEnabled: boolean;
  onIsEnabledChange: (value: boolean) => void;

  config: OAuthSSOProviderConfig;
  secret: OAuthClientSecret;
  onChange: (config: OAuthSSOProviderConfig, secret: OAuthClientSecret) => void;

  disabled: boolean;
  limitReached: boolean;
}

type WidgetTextFieldKey =
  | keyof Omit<OAuthSSOProviderConfig, "type">
  | "client_secret";

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
      "modify_disabled",
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
      "modify_disabled",
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
      "modify_disabled",
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
      "modify_disabled",
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
      "modify_disabled",
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
      "modify_disabled",
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
      "modify_disabled",
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
      "modify_disabled",
    ]),
    isSecretFieldTextArea: false,
  },
};

const WidgetHeader: React.FC<WidgetHeaderProps> = function WidgetHeader(
  props: WidgetHeaderProps
) {
  const {
    icon,
    messageID,
    isEnabled,
    setIsEnabled,
    disabled,
    disabledByLimitReached,
  } = props;

  const onChange = useCallback(
    (_, value?: boolean) => {
      setIsEnabled(value ?? false);
    },
    [setIsEnabled]
  );

  let messageBar;
  if (disabled) {
    messageBar = (
      <MessageBar>
        <FormattedMessage
          id="FeatureConfig.disabled"
          values={{
            planPagePath: "../../billing",
          }}
        />
      </MessageBar>
    );
  }

  return (
    <div>
      <Toggle
        checked={isEnabled}
        onChange={onChange}
        inlineLabel={true}
        label={
          <>
            {icon}
            <FormattedMessage id={messageID} />
          </>
        }
        disabled={!isEnabled && (disabled || disabledByLimitReached)}
      ></Toggle>
      {messageBar}
    </div>
  );
};

const SingleSignOnConfigurationWidget: React.FC<SingleSignOnConfigurationWidgetProps> =
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
      disabled,
      limitReached,
    } = props;

    const { renderToString } = useContext(Context);

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

    const [extended, setExtended] = useState(isEnabled);

    // Always extended when enabled
    // Collapse on disabled
    // make sure text field mounted for showing error
    useEffect(() => {
      if (isEnabled) {
        setExtended(true);
      }
    }, [isEnabled]);

    const onExtendClicked = useCallback(() => {
      setExtended(!extended);
    }, [extended]);

    const onAliasChange = useCallback(
      (_, value?: string) =>
        onChange(
          { ...config, alias: value ?? "" },
          { ...secret, alias: value ?? "" }
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
        onChange(config, { ...secret, clientSecret: value ?? "" }),
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
    const onModifyDisabledChange = useCallback(
      (_, value?: boolean) =>
        onChange({ ...config, modify_disabled: value ?? false }, secret),
      [onChange, config, secret]
    );

    const disabledByLimitReached = useMemo(() => {
      return !isEnabled && limitReached;
    }, [limitReached, isEnabled]);

    return (
      <ExtendableWidget
        className={className}
        extendButtonAriaLabelId={messageID}
        extendButtonDisabled={isEnabled}
        extended={extended}
        onExtendClicked={onExtendClicked}
        readOnly={!isEnabled || disabled || disabledByLimitReached}
        HeaderComponent={
          <WidgetHeader
            icon={iconNode}
            isEnabled={isEnabled}
            setIsEnabled={onIsEnabledChange}
            messageID={messageID}
            disabled={disabled}
            disabledByLimitReached={disabledByLimitReached}
          />
        }
      >
        {visibleFields.has("alias") && (
          <FormTextField
            parentJSONPointer={jsonPointer}
            fieldName="alias"
            fieldNameMessageID="SingleSignOnConfigurationScreen.widget.alias"
            className={styles.textField}
            styles={TEXT_FIELD_STYLE}
            value={config.alias}
            onChange={onAliasChange}
          />
        )}
        {visibleFields.has("client_id") && (
          <FormTextField
            parentJSONPointer={jsonPointer}
            fieldName="client_id"
            fieldNameMessageID="SingleSignOnConfigurationScreen.widget.client-id"
            className={styles.textField}
            styles={TEXT_FIELD_STYLE}
            value={config.client_id ?? ""}
            onChange={onClientIDChange}
          />
        )}
        {visibleFields.has("client_secret") && (
          <FormTextField
            parentJSONPointer={clientSecretParentJsonPointer}
            fieldName="client_secret"
            fieldNameMessageID="SingleSignOnConfigurationScreen.widget.client-secret"
            className={styles.textField}
            styles={
              isSecretFieldTextArea
                ? MULTILINE_TEXT_FIELD_STYLE
                : TEXT_FIELD_STYLE
            }
            multiline={isSecretFieldTextArea}
            value={secret.clientSecret}
            onChange={onClientSecretChange}
          />
        )}
        {visibleFields.has("tenant") && (
          <FormTextField
            parentJSONPointer={jsonPointer}
            fieldName="tenant"
            fieldNameMessageID="SingleSignOnConfigurationScreen.widget.tenant"
            className={styles.textField}
            styles={TEXT_FIELD_STYLE}
            value={config.tenant ?? ""}
            onChange={onTenantChange}
          />
        )}
        {visibleFields.has("discovery_document_endpoint") && (
          <FormTextField
            parentJSONPointer={jsonPointer}
            fieldName="discovery_document_endpoint"
            fieldNameMessageID="SingleSignOnConfigurationScreen.widget.discovery-document-endpoint"
            className={styles.textField}
            styles={TEXT_FIELD_STYLE}
            value={config.discovery_document_endpoint ?? ""}
            onChange={onDiscoveryDocumentEndpointChange}
            placeholder="http://example.com/.well-known/openid-configuration"
          />
        )}
        {visibleFields.has("key_id") && (
          <FormTextField
            parentJSONPointer={jsonPointer}
            fieldName="key_id"
            fieldNameMessageID="SingleSignOnConfigurationScreen.widget.key-id"
            className={styles.textField}
            styles={TEXT_FIELD_STYLE}
            value={config.key_id ?? ""}
            onChange={onKeyIDChange}
          />
        )}
        {visibleFields.has("team_id") && (
          <FormTextField
            parentJSONPointer={jsonPointer}
            fieldName="team_id"
            fieldNameMessageID="SingleSignOnConfigurationScreen.widget.team-id"
            className={styles.textField}
            styles={TEXT_FIELD_STYLE}
            value={config.team_id ?? ""}
            onChange={onTeamIDChange}
          />
        )}
        {visibleFields.has("account_id") && (
          <FormTextField
            parentJSONPointer={jsonPointer}
            fieldName="account_id"
            fieldNameMessageID="SingleSignOnConfigurationScreen.widget.account-id"
            className={styles.textField}
            styles={TEXT_FIELD_STYLE}
            value={config.account_id ?? ""}
            onChange={onAccountIDChange}
          />
        )}
        {visibleFields.has("is_sandbox_account") && (
          <Checkbox
            label={renderToString(
              "SingleSignOnConfigurationScreen.widget.is-sandbox-account"
            )}
            className={styles.checkbox}
            checked={config.is_sandbox_account ?? false}
            onChange={onIsSandBoxAccountChange}
          />
        )}
        {visibleFields.has("wechat_redirect_uris") && (
          <FormTextFieldList
            parentJSONPointer={jsonPointer}
            fieldName="wechat_redirect_uris"
            list={config.wechat_redirect_uris ?? []}
            onListChange={onWeChatRedirectUrisChange}
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
          />
        )}
        {visibleFields.has("modify_disabled") && (
          <Checkbox
            label={renderToString(
              "SingleSignOnConfigurationScreen.widget.modify-disabled"
            )}
            className={styles.checkbox}
            checked={config.modify_disabled ?? false}
            onChange={onModifyDisabledChange}
          />
        )}
      </ExtendableWidget>
    );
  };

export default SingleSignOnConfigurationWidget;
