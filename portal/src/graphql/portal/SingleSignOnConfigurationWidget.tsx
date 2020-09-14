import React, {
  Dispatch,
  SetStateAction,
  useCallback,
  useContext,
} from "react";
import { Toggle, Label, TextField } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import {
  OAuthSSOProviderConfigState,
  SingleSignOnScreenState,
} from "./SingleSignOnConfigurationScreen";
import { OAuthSSOProviderConfig, OAuthSSOProviderType } from "../../types";

import styles from "./SingleSignOnConfigurationWidget.module.scss";

interface WidgetHeaderLabelProps {
  icon: React.ReactNode;
  serviceMessageId: string;
}

interface WidgetHeaderProps extends WidgetHeaderLabelProps {
  enabled: boolean;
  setEnabled: (event: any, enabled?: boolean) => void;
}

interface SingleSignOnConfigurationWidgetProps {
  className?: string;
  serviceProviderType: OAuthSSOProviderType;
  serviceProviderConfig?: OAuthSSOProviderConfig;
  screenState: SingleSignOnScreenState;
  setScreenState: Dispatch<SetStateAction<SingleSignOnScreenState>>;
}

type WidgetTextFieldKey = keyof Omit<
  Omit<OAuthSSOProviderConfigState, "enabled">,
  "type"
>;

const serviceProviderMessageId: { [key in OAuthSSOProviderType]: string } = {
  apple: "apple",
  google: "google",
  facebook: "facebook",
  linkedin: "linkedin",
  azureadv2: "azureadv2",
};

const serviceProviderIcon: {
  [key in OAuthSSOProviderType]: React.ReactNode;
} = {
  apple: <div className={styles.widgetLabelIcon} />,
  google: <div className={styles.widgetLabelIcon} />,
  facebook: <div className={styles.widgetLabelIcon} />,
  linkedin: <div className={styles.widgetLabelIcon} />,
  azureadv2: <div className={styles.widgetLabelIcon} />,
};

const visibleFieldsMap: {
  [key in OAuthSSOProviderType]: Set<WidgetTextFieldKey>;
} = {
  apple: new Set<WidgetTextFieldKey>([
    "alias",
    "client_id",
    "clientSecret",
    "tenant",
    "key_id",
    "team_id",
  ]),
  google: new Set<WidgetTextFieldKey>([
    "alias",
    "client_id",
    "clientSecret",
    "tenant",
    "key_id",
    "team_id",
  ]),
  facebook: new Set<WidgetTextFieldKey>([
    "alias",
    "client_id",
    "clientSecret",
    "tenant",
    "key_id",
    "team_id",
  ]),
  linkedin: new Set<WidgetTextFieldKey>([
    "alias",
    "client_id",
    "clientSecret",
    "tenant",
    "key_id",
    "team_id",
  ]),
  azureadv2: new Set<WidgetTextFieldKey>([
    "alias",
    "client_id",
    "clientSecret",
    "tenant",
    "key_id",
    "team_id",
  ]),
};

const isSecretFieldTextArea: {
  [key in OAuthSSOProviderType]: boolean;
} = {
  apple: true,
  google: false,
  facebook: false,
  linkedin: false,
  azureadv2: false,
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
  const { className, screenState, serviceProviderType } = props;
  const { renderToString } = useContext(Context);

  const icon = serviceProviderIcon[serviceProviderType];
  const serviceMessageId = serviceProviderMessageId[serviceProviderType];
  const visibleFields = visibleFieldsMap[serviceProviderType];
  const widgetState = screenState[serviceProviderType];

  return (
    <ExtendableWidget
      className={className}
      extendButtonAriaLabelId={serviceMessageId}
      extendable={true}
      initiallyExtended={true}
      readOnly={false}
      HeaderComponent={
        <WidgetHeader
          icon={icon}
          enabled={widgetState.enabled}
          setEnabled={() => {}}
          serviceMessageId={serviceMessageId}
        />
      }
    >
      {visibleFields.has("alias") && (
        <TextField
          className={styles.textField}
          label={renderToString("SingleSignOnConfigurationScreen.widget.alias")}
        />
      )}
      {visibleFields.has("client_id") && (
        <TextField
          className={styles.textField}
          label={renderToString(
            "SingleSignOnConfigurationScreen.widget.client-id"
          )}
        />
      )}
      {visibleFields.has("clientSecret") && (
        <TextField
          className={styles.textField}
          label={renderToString(
            "SingleSignOnConfigurationScreen.widget.client-secret"
          )}
          multiline={isSecretFieldTextArea[serviceProviderType]}
        />
      )}
      {visibleFields.has("tenant") && (
        <TextField
          className={styles.textField}
          label={renderToString(
            "SingleSignOnConfigurationScreen.widget.tenant"
          )}
        />
      )}
      {visibleFields.has("key_id") && (
        <TextField
          className={styles.textField}
          label={renderToString(
            "SingleSignOnConfigurationScreen.widget.key-id"
          )}
        />
      )}
      {visibleFields.has("team_id") && (
        <TextField
          className={styles.textField}
          label={renderToString(
            "SingleSignOnConfigurationScreen.widget.team-id"
          )}
        />
      )}
    </ExtendableWidget>
  );
};

export default SingleSignOnConfigurationWidget;
