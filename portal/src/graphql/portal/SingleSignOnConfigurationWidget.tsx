import React, {
  Dispatch,
  SetStateAction,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { Toggle, Label, TextField } from "@fluentui/react";
import { FormattedMessage, Context } from "@oursky/react-messageformat";

import ExtendableWidget from "../../ExtendableWidget";
import {
  OAuthSSOProviderConfigState,
  SingleSignOnScreenState,
} from "./SingleSignOnConfigurationScreen";
import { OAuthSSOProviderType } from "../../types";
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
  setEnabled: (event: any, enabled?: boolean) => void;
}

interface SingleSignOnConfigurationWidgetProps {
  className?: string;
  serviceProviderType: OAuthSSOProviderType;
  screenState: SingleSignOnScreenState;
  setScreenState: Dispatch<SetStateAction<SingleSignOnScreenState>>;
  violations?: Violation[];
}

type WidgetTextFieldKey = keyof Omit<
  Omit<OAuthSSOProviderConfigState, "enabled">,
  "type"
>;

type WidgetErrorState = Partial<Record<WidgetTextFieldKey, string>>;

const TEXT_FIELD_STYLE = { errorMessage: { whiteSpace: "pre" } };

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

function useSetWidgetState(
  setScreenState: Dispatch<SetStateAction<SingleSignOnScreenState>>,
  serviceProviderType: OAuthSSOProviderType
) {
  return useCallback(
    (widgetState: OAuthSSOProviderConfigState) => {
      setScreenState((prev) => ({
        ...prev,
        [serviceProviderType]: widgetState,
      }));
    },
    [setScreenState, serviceProviderType]
  );
}

function useTextField(
  textFieldKey: WidgetTextFieldKey,
  widgetState: OAuthSSOProviderConfigState,
  setWidgetState: (widgetState: OAuthSSOProviderConfigState) => void
) {
  const onChange = useCallback(
    (_event, value?: string) => {
      if (value == null) {
        return;
      }
      setWidgetState({ ...widgetState, [textFieldKey]: value });
    },
    [widgetState, setWidgetState, textFieldKey]
  );
  return { onChange };
}

const SingleSignOnConfigurationWidget: React.FC<SingleSignOnConfigurationWidgetProps> = function SingleSignOnConfigurationWidget(
  props: SingleSignOnConfigurationWidgetProps
) {
  const {
    className,
    screenState,
    setScreenState,
    serviceProviderType,
    violations,
  } = props;
  const { renderToString } = useContext(Context);
  const setWidgetState = useSetWidgetState(setScreenState, serviceProviderType);

  const icon = serviceProviderIcon[serviceProviderType];
  const serviceMessageId = serviceProviderMessageId[serviceProviderType];
  const visibleFields = visibleFieldsMap[serviceProviderType];
  const widgetState = screenState[serviceProviderType];

  const [errorMap, setErrorMap] = useState<WidgetErrorState>({});

  const setEnabled = useCallback(
    (_event, enabled?: boolean) => {
      if (enabled == null) {
        return;
      }
      setWidgetState({
        ...widgetState,
        enabled,
      });
    },
    [widgetState, setWidgetState]
  );

  const { onChange: onAliasChange } = useTextField(
    "alias",
    widgetState,
    setWidgetState
  );

  const { onChange: onClientIdChange } = useTextField(
    "client_id",
    widgetState,
    setWidgetState
  );

  const { onChange: onClientSecretChange } = useTextField(
    "clientSecret",
    widgetState,
    setWidgetState
  );

  const { onChange: onTenantChange } = useTextField(
    "tenant",
    widgetState,
    setWidgetState
  );

  const { onChange: onKeyIdChange } = useTextField(
    "key_id",
    widgetState,
    setWidgetState
  );

  const { onChange: onTeamIdChange } = useTextField(
    "team_id",
    widgetState,
    setWidgetState
  );

  useEffect(() => {
    if (violations == null || violations.length === 0) {
      setErrorMap({});
      return;
    }
    const widgetDataLocationPrefix = "/identity/oauth/providers";
    const violationMap = violationSelector(violations, {
      alias: makeMissingFieldSelector(widgetDataLocationPrefix, "alias"),
      key_id: makeMissingFieldSelector(widgetDataLocationPrefix, "key_id"),
      client_id: makeMissingFieldSelector(
        widgetDataLocationPrefix,
        "client_id"
      ),
      tenant: makeMissingFieldSelector(widgetDataLocationPrefix, "tenant"),
      team_id: makeMissingFieldSelector(widgetDataLocationPrefix, "team_id"),
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
  }, [renderToString, violations]);

  return (
    <ExtendableWidget
      className={className}
      extendButtonAriaLabelId={serviceMessageId}
      extendable={true}
      readOnly={!widgetState.enabled}
      initiallyExtended={true}
      HeaderComponent={
        <WidgetHeader
          icon={icon}
          enabled={widgetState.enabled}
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
          value={widgetState.alias}
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
          value={widgetState.client_id}
          onChange={onClientIdChange}
          errorMessage={errorMap["client_id"]}
        />
      )}
      {visibleFields.has("clientSecret") && (
        <TextField
          className={styles.textField}
          styles={TEXT_FIELD_STYLE}
          label={renderToString(
            "SingleSignOnConfigurationScreen.widget.client-secret"
          )}
          multiline={isSecretFieldTextArea[serviceProviderType]}
          value={widgetState.clientSecret}
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
          value={widgetState.tenant}
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
          value={widgetState.key_id}
          onChange={onKeyIdChange}
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
          value={widgetState.team_id}
          onChange={onTeamIdChange}
          errorMessage={errorMap["team_id"]}
        />
      )}
    </ExtendableWidget>
  );
};

export default SingleSignOnConfigurationWidget;
