import React, { useCallback, useEffect, useState } from "react";
import cn from "classnames";
import { Toggle, Label, ITextFieldProps, IToggleProps } from "@fluentui/react";
import { FormattedMessage } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import FormTextField from "../../FormTextField";
import { OAuthSSOProviderConfig, OAuthSSOProviderType } from "../../types";

import styles from "./SingleSignOnConfigurationWidget.module.scss";

interface WidgetHeaderLabelProps {
  icon: React.ReactNode;
  messageID: string;
}

interface WidgetHeaderProps extends WidgetHeaderLabelProps {
  enabled: boolean;
  setEnabled: IToggleProps["onChange"];
}

interface SingleSignOnConfigurationWidgetProps {
  className?: string;

  jsonPointer: string;
  clientSecretJsonPointer: string;

  enabled: boolean;
  alias: string;
  serviceProviderType: OAuthSSOProviderType;
  clientID: string;
  clientSecret: string;
  tenant?: string;
  keyID?: string;
  teamID?: string;

  setEnabled: IToggleProps["onChange"];
  onAliasChange: ITextFieldProps["onChange"];
  onClientIDChange: ITextFieldProps["onChange"];
  onClientSecretChange: ITextFieldProps["onChange"];
  onTenantChange: ITextFieldProps["onChange"];
  onKeyIDChange: ITextFieldProps["onChange"];
  onTeamIDChange: ITextFieldProps["onChange"];
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
  const { icon, messageID, enabled, setEnabled } = props;
  return (
    <Toggle
      checked={enabled}
      onChange={setEnabled}
      inlineLabel={true}
      label={<WidgetHeaderLabel icon={icon} messageID={messageID} />}
    ></Toggle>
  );
};

const SingleSignOnConfigurationWidget: React.FC<SingleSignOnConfigurationWidgetProps> = function SingleSignOnConfigurationWidget(
  props: SingleSignOnConfigurationWidgetProps
) {
  const {
    className,
    jsonPointer,
    clientSecretJsonPointer,
    enabled,
    alias,
    clientID,
    clientSecret,
    tenant,
    keyID,
    teamID,
    setEnabled,
    onAliasChange,
    onClientIDChange,
    onClientSecretChange,
    onTenantChange,
    onKeyIDChange,
    onTeamIDChange,
    serviceProviderType,
  } = props;

  const {
    providerType,
    isSecretFieldTextArea,
    iconNode,
    fields: visibleFields,
  } = oauthProviders[serviceProviderType];

  const messageID = "OAuthBranding." + providerType;

  const [extended, setExtended] = useState(enabled);

  // Always extended when enabled
  // Collapse on disabled
  // make sure text field mounted for showing error
  useEffect(() => {
    setExtended(enabled);
  }, [enabled]);

  const onExtendClicked = useCallback(() => {
    setExtended(!extended);
  }, [extended]);

  return (
    <ExtendableWidget
      className={className}
      extendButtonAriaLabelId={messageID}
      extendButtonDisabled={enabled}
      extended={extended}
      onExtendClicked={onExtendClicked}
      readOnly={!enabled}
      HeaderComponent={
        <WidgetHeader
          icon={iconNode}
          enabled={enabled}
          setEnabled={setEnabled}
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
          value={alias}
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
          value={clientID}
          onChange={onClientIDChange}
        />
      )}
      {visibleFields.has("client_secret") && (
        <FormTextField
          jsonPointer={`${clientSecretJsonPointer}/client_secret`}
          parentJSONPointer={clientSecretJsonPointer}
          fieldName="client_secret"
          fieldNameMessageID="SingleSignOnConfigurationScreen.widget.client-secret"
          className={styles.textField}
          styles={
            isSecretFieldTextArea
              ? MULTILINE_TEXT_FIELD_STYLE
              : TEXT_FIELD_STYLE
          }
          multiline={isSecretFieldTextArea}
          value={clientSecret}
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
          value={tenant}
          onChange={onTenantChange}
        />
      )}
      {visibleFields.has("key_id") && (
        <FormTextField
          jsonPointer={`${jsonPointer}/key_id`}
          parentJSONPointer={jsonPointer}
          fieldName="key-id"
          fieldNameMessageID="SingleSignOnConfigurationScreen.widget.key-id"
          className={styles.textField}
          styles={TEXT_FIELD_STYLE}
          value={keyID}
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
          value={teamID}
          onChange={onTeamIDChange}
        />
      )}
    </ExtendableWidget>
  );
};

export default SingleSignOnConfigurationWidget;
