import React, { useContext, useEffect, useState } from "react";
import cn from "classnames";
import {
  Toggle,
  Label,
  TextField,
  ITextFieldProps,
  IToggleProps,
} from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import { OAuthSSOProviderConfig, OAuthSSOProviderType } from "../../types";
import {
  errorFormatter,
  makeMissingFieldSelector,
  Violation,
  violationSelector,
} from "../../util/validation";

import styles from "./SingleSignOnConfigurationWidget.module.scss";

interface WidgetHeaderLabelProps {
  icon: React.ReactNode;
  serviceMessageId: string;
}

interface WidgetHeaderProps extends WidgetHeaderLabelProps {
  enabled: boolean;
  setEnabled: IToggleProps["onChange"];
}

interface SingleSignOnConfigurationWidgetProps {
  className?: string;

  errorLocation?: string;

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
  violations?: Violation[];
}

type WidgetTextFieldKey =
  | keyof Omit<OAuthSSOProviderConfig, "type">
  | "client_secret";

type WidgetErrorState = Partial<Record<WidgetTextFieldKey, string>>;

interface OAuthProviderInfo {
  messageId: string;
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
    messageId: "apple",
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
    messageId: "google",
    iconNode: <i className={cn("fab", "fa-google", styles.widgetLabelIcon)} />,
    fields: new Set<WidgetTextFieldKey>([
      "alias",
      "client_id",
      "client_secret",
    ]),
    isSecretFieldTextArea: false,
  },
  facebook: {
    messageId: "facebook",
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
    messageId: "linkedin",
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
    messageId: "azureadv2",
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
  const { icon, serviceMessageId } = props;
  const { renderToString } = useContext(Context);
  return (
    <div className={styles.widgetLabel}>
      {icon}
      <Label className={styles.widgetLabelText}>
        <FormattedMessage
          id="SingleSignOnConfigurationScreen.widget.header"
          values={{ serviveName: renderToString(serviceMessageId) }}
        />
      </Label>
    </div>
  );
};

const WidgetHeader: React.FC<WidgetHeaderProps> = function WidgetHeader(
  props: WidgetHeaderProps
) {
  const { icon, serviceMessageId, enabled, setEnabled } = props;
  return (
    <Toggle
      checked={enabled}
      onChange={setEnabled}
      inlineLabel={true}
      label={
        <WidgetHeaderLabel icon={icon} serviceMessageId={serviceMessageId} />
      }
    ></Toggle>
  );
};

const SingleSignOnConfigurationWidget: React.FC<SingleSignOnConfigurationWidgetProps> = function SingleSignOnConfigurationWidget(
  props: SingleSignOnConfigurationWidgetProps
) {
  const {
    className,
    errorLocation,
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
    violations,
  } = props;
  const { renderToString } = useContext(Context);

  const {
    messageId: serviceMessageId,
    isSecretFieldTextArea,
    iconNode,
    fields: visibleFields,
  } = oauthProviders[serviceProviderType];

  const [errorMap, setErrorMap] = useState<WidgetErrorState>({});

  useEffect(() => {
    if (
      errorLocation == null ||
      violations == null ||
      violations.length === 0
    ) {
      setErrorMap({});
      return;
    }

    const violationMap = violationSelector(violations, {
      alias: makeMissingFieldSelector(errorLocation, "alias"),
      key_id: makeMissingFieldSelector(errorLocation, "key_id"),
      client_id: makeMissingFieldSelector(errorLocation, "client_id"),
      tenant: makeMissingFieldSelector(errorLocation, "tenant"),
      team_id: makeMissingFieldSelector(errorLocation, "team_id"),
    });

    setErrorMap({
      alias: errorFormatter(
        "SingleSignOnConfigurationScreen.widget.alias",
        violationMap["alias"],
        renderToString
      ),
      key_id: errorFormatter(
        "SingleSignOnConfigurationScreen.widget.key-id",
        violationMap["key_id"],
        renderToString
      ),
      client_id: errorFormatter(
        "SingleSignOnConfigurationScreen.widget.client-id",
        violationMap["client_id"],
        renderToString
      ),
      tenant: errorFormatter(
        "SingleSignOnConfigurationScreen.widget.tenant",
        violationMap["tenant"],
        renderToString
      ),
      team_id: errorFormatter(
        "SingleSignOnConfigurationScreen.widget.team-id",
        violationMap["team_id"],
        renderToString
      ),
    });
  }, [renderToString, violations, errorLocation]);

  return (
    <ExtendableWidget
      className={className}
      extendButtonAriaLabelId={serviceMessageId}
      extendable={true}
      readOnly={!enabled}
      initiallyExtended={enabled}
      HeaderComponent={
        <WidgetHeader
          icon={iconNode}
          enabled={enabled}
          setEnabled={setEnabled}
          serviceMessageId={serviceMessageId}
        />
      }
    >
      {visibleFields.has("alias") && (
        <TextField
          className={styles.textField}
          styles={TEXT_FIELD_STYLE}
          label={renderToString("SingleSignOnConfigurationScreen.widget.alias")}
          value={alias}
          onChange={onAliasChange}
          errorMessage={errorMap["alias"]}
        />
      )}
      {visibleFields.has("client_id") && (
        <TextField
          className={styles.textField}
          styles={TEXT_FIELD_STYLE}
          label={renderToString(
            "SingleSignOnConfigurationScreen.widget.client-id"
          )}
          value={clientID}
          onChange={onClientIDChange}
          errorMessage={errorMap["client_id"]}
        />
      )}
      {visibleFields.has("client_secret") && (
        <TextField
          className={styles.textField}
          styles={
            isSecretFieldTextArea
              ? MULTILINE_TEXT_FIELD_STYLE
              : TEXT_FIELD_STYLE
          }
          multiline={isSecretFieldTextArea}
          label={renderToString(
            "SingleSignOnConfigurationScreen.widget.client-secret"
          )}
          value={clientSecret}
          onChange={onClientSecretChange}
        />
      )}
      {visibleFields.has("tenant") && (
        <TextField
          className={styles.textField}
          styles={TEXT_FIELD_STYLE}
          label={renderToString(
            "SingleSignOnConfigurationScreen.widget.tenant"
          )}
          value={tenant}
          onChange={onTenantChange}
          errorMessage={errorMap["tenant"]}
        />
      )}
      {visibleFields.has("key_id") && (
        <TextField
          className={styles.textField}
          styles={TEXT_FIELD_STYLE}
          label={renderToString(
            "SingleSignOnConfigurationScreen.widget.key-id"
          )}
          value={keyID}
          onChange={onKeyIDChange}
          errorMessage={errorMap["key_id"]}
        />
      )}
      {visibleFields.has("team_id") && (
        <TextField
          className={styles.textField}
          styles={TEXT_FIELD_STYLE}
          label={renderToString(
            "SingleSignOnConfigurationScreen.widget.team-id"
          )}
          value={teamID}
          onChange={onTeamIDChange}
          errorMessage={errorMap["team_id"]}
        />
      )}
    </ExtendableWidget>
  );
};

export default SingleSignOnConfigurationWidget;
