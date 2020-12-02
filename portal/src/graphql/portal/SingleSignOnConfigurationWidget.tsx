import React, { useCallback, useEffect, useMemo, useState } from "react";
import cn from "classnames";
import { Label, Toggle } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import FormTextField from "../../FormTextField";
import {
  OAuthClientCredentialItem,
  OAuthSSOProviderConfig,
  OAuthSSOProviderType,
} from "../../types";

import styles from "./SingleSignOnConfigurationWidget.module.scss";

interface WidgetHeaderLabelProps {
  icon: React.ReactNode;
  messageID: string;
}

interface WidgetHeaderProps extends WidgetHeaderLabelProps {
  isEnabled: boolean;
  setIsEnabled: (value: boolean) => void;
}

interface SingleSignOnConfigurationWidgetProps {
  className?: string;

  jsonPointer: string;
  clientSecretParentJsonPointer: string | RegExp;

  isEnabled: boolean;
  onIsEnabledChange: (value: boolean) => void;

  config: OAuthSSOProviderConfig;
  secret: OAuthClientCredentialItem;
  onChange: (
    config: OAuthSSOProviderConfig,
    secret: OAuthClientCredentialItem
  ) => void;
}

type WidgetTextFieldKey =
  | keyof Omit<OAuthSSOProviderConfig, "type">
  | "client_secret";

interface OAuthProviderInfo {
  providerType: OAuthSSOProviderType;
  iconNode: React.ReactNode;
  fields: Set<WidgetTextFieldKey>;
  isSecretFieldTextArea: boolean;
}

const TEXT_FIELD_STYLE = { errorMessage: { whiteSpace: "pre" } };
const MULTILINE_TEXT_FIELD_STYLE = {
  errorMessage: { whiteSpace: "pre" },
  field: { minHeight: "160px" },
};

const oauthProviders: Record<OAuthSSOProviderType, OAuthProviderInfo> = {
  apple: {
    providerType: "apple",
    iconNode: <i className={cn("fab", "fa-apple", styles.widgetLabelIcon)} />,
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
      "key_id",
      "team_id",
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
    ]),
    isSecretFieldTextArea: false,
  },
};

const WidgetHeaderLabel: React.FC<WidgetHeaderLabelProps> = function WidgetHeaderLabel(
  props: WidgetHeaderLabelProps
) {
  const { icon, messageID } = props;
  return (
    <div className={styles.widgetLabel}>
      {icon}
      <Label className={styles.widgetLabelText}>
        <FormattedMessage id={messageID} />
      </Label>
    </div>
  );
};

const WidgetHeader: React.FC<WidgetHeaderProps> = function WidgetHeader(
  props: WidgetHeaderProps
) {
  const { icon, messageID, isEnabled, setIsEnabled } = props;

  const onChange = useCallback(
    (_, value?: boolean) => {
      setIsEnabled(value ?? false);
    },
    [setIsEnabled]
  );

  return (
    <Toggle
      checked={isEnabled}
      onChange={onChange}
      inlineLabel={true}
      label={<WidgetHeaderLabel icon={icon} messageID={messageID} />}
    ></Toggle>
  );
};

// eslint-disable-next-line complexity
const SingleSignOnConfigurationWidget: React.FC<SingleSignOnConfigurationWidgetProps> = function SingleSignOnConfigurationWidget(
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
  } = props;

  const {
    providerType,
    isSecretFieldTextArea,
    iconNode,
    fields: visibleFields,
  } = oauthProviders[config.type];

  const messageID = "OAuthBranding." + providerType;

  const [extended, setExtended] = useState(isEnabled);

  const clientSecretJSONPointer = useMemo(() => {
    if (typeof clientSecretParentJsonPointer === "string") {
      return clientSecretParentJsonPointer
        ? `${clientSecretParentJsonPointer}/client_secret`
        : "";
    }
    return new RegExp(
      `${clientSecretParentJsonPointer.source.replace("$", "")}/client_secret$`
    );
  }, [clientSecretParentJsonPointer]);

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
    (_, value?: string) => onChange({ ...config, tenant: value ?? "" }, secret),
    [onChange, config, secret]
  );
  const onKeyIDChange = useCallback(
    (_, value?: string) => onChange({ ...config, key_id: value ?? "" }, secret),
    [onChange, config, secret]
  );
  const onTeamIDChange = useCallback(
    (_, value?: string) =>
      onChange({ ...config, team_id: value ?? "" }, secret),
    [onChange, config, secret]
  );

  const onClientSecretChange = useCallback(
    (_, value?: string) =>
      onChange(config, { ...secret, client_secret: value ?? "" }),
    [onChange, config, secret]
  );

  return (
    <ExtendableWidget
      className={className}
      extendButtonAriaLabelId={messageID}
      extendButtonDisabled={isEnabled}
      extended={extended}
      onExtendClicked={onExtendClicked}
      readOnly={!isEnabled}
      HeaderComponent={
        <WidgetHeader
          icon={iconNode}
          isEnabled={isEnabled}
          setIsEnabled={onIsEnabledChange}
          messageID={messageID}
        />
      }
    >
      {visibleFields.has("alias") && (
        <FormTextField
          jsonPointer={`${jsonPointer}/alias`}
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
          jsonPointer={`${jsonPointer}/client_id`}
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
          jsonPointer={clientSecretJSONPointer}
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
          value={secret.client_secret}
          onChange={onClientSecretChange}
        />
      )}
      {visibleFields.has("tenant") && (
        <FormTextField
          jsonPointer={`${jsonPointer}/tenant`}
          parentJSONPointer={jsonPointer}
          fieldName="tenant"
          fieldNameMessageID="SingleSignOnConfigurationScreen.widget.tenant"
          className={styles.textField}
          styles={TEXT_FIELD_STYLE}
          value={config.tenant ?? ""}
          onChange={onTenantChange}
        />
      )}
      {visibleFields.has("key_id") && (
        <FormTextField
          jsonPointer={`${jsonPointer}/key_id`}
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
          jsonPointer={`${jsonPointer}/team_id`}
          parentJSONPointer={jsonPointer}
          fieldName="team_id"
          fieldNameMessageID="SingleSignOnConfigurationScreen.widget.team-id"
          className={styles.textField}
          styles={TEXT_FIELD_STYLE}
          value={config.team_id ?? ""}
          onChange={onTeamIDChange}
        />
      )}
    </ExtendableWidget>
  );
};

export default SingleSignOnConfigurationWidget;
