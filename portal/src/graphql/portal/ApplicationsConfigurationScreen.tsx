import React, { useCallback, useContext, useMemo, useState } from "react";
import {
  ActionButton,
  DetailsList,
  IButtonStyles,
  IColumn,
  ICommandBarItemProps,
  IconButton,
  MessageBar,
  SelectionMode,
  Text,
  VerticalDivider,
  Toggle,
  TextField,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useNavigate, useParams } from "react-router-dom";
import produce from "immer";
import cn from "classnames";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { OAuthClientConfig, PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import { useSystemConfig } from "../../context/SystemConfigContext";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import FormContainer from "../../FormContainer";

import styles from "./ApplicationsConfigurationScreen.module.scss";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import WidgetDescription from "../../WidgetDescription";
import FormTextFieldList from "../../FormTextFieldList";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";

const COPY_ICON_STLYES: IButtonStyles = {
  root: { margin: 4 },
  rootHovered: { backgroundColor: "#d8d6d3" },
  rootPressed: { backgroundColor: "#c2c0be" },
};

interface FormState {
  publicOrigin: string;
  cookieDomain?: string;
  clients: OAuthClientConfig[];
  allowedOrigins: string[];
  persistentCookie: boolean;
  sessionLifetimeSeconds: number | undefined;
  idleTimeoutEnabled: boolean;
  idleTimeoutSeconds: number | undefined;
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    publicOrigin: config.http?.public_origin ?? "",
    cookieDomain: config.http?.cookie_domain,
    clients: config.oauth?.clients ?? [],
    allowedOrigins: config.http?.allowed_origins ?? [],
    persistentCookie: !(config.session?.cookie_non_persistent ?? false),
    sessionLifetimeSeconds: config.session?.lifetime_seconds,
    idleTimeoutEnabled: config.session?.idle_timeout_enabled ?? false,
    idleTimeoutSeconds: config.session?.idle_timeout_seconds,
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  // eslint-disable-next-line complexity
  return produce(config, (config) => {
    config.oauth ??= {};
    config.oauth.clients = currentState.clients;
    config.http ??= {};
    config.http.allowed_origins = currentState.allowedOrigins;
    config.session = config.session ?? {};
    if (initialState.persistentCookie !== currentState.persistentCookie) {
      config.session.cookie_non_persistent = !currentState.persistentCookie;
    }
    if (
      initialState.sessionLifetimeSeconds !==
      currentState.sessionLifetimeSeconds
    ) {
      config.session.lifetime_seconds = currentState.sessionLifetimeSeconds;
    }

    if (initialState.idleTimeoutEnabled !== currentState.idleTimeoutEnabled) {
      config.session.idle_timeout_enabled = currentState.idleTimeoutEnabled;
    }

    if (
      currentState.idleTimeoutEnabled &&
      initialState.idleTimeoutSeconds !== currentState.idleTimeoutSeconds
    ) {
      config.session.idle_timeout_seconds = currentState.idleTimeoutSeconds;
    }
    clearEmptyObject(config);
  });
}

function makeOAuthClientListColumns(
  renderToString: (messageId: string) => string
): IColumn[] {
  return [
    {
      key: "name",
      fieldName: "name",
      name: renderToString("ApplicationsConfigurationScreen.client-list.name"),
      minWidth: 100,
      className: styles.columnHeader,
    },

    {
      key: "clientId",
      fieldName: "clientId",
      name: renderToString(
        "ApplicationsConfigurationScreen.client-list.client-id"
      ),
      minWidth: 250,
      className: styles.columnHeader,
    },
    {
      key: "action",
      name: renderToString("action"),
      className: styles.columnHeader,
      minWidth: 150,
    },
  ];
}

interface OAuthClientIdCellProps {
  clientId: string;
}

const OAuthClientIdCell: React.FC<OAuthClientIdCellProps> =
  function OAuthClientIdCell(props) {
    const { clientId } = props;
    const { copyButtonProps, Feedback } = useCopyFeedback({
      textToCopy: clientId,
    });

    return (
      <>
        <span className={styles.cellContent}>{clientId}</span>
        <IconButton {...copyButtonProps} styles={COPY_ICON_STLYES} />
        <Feedback />
      </>
    );
  };

interface OAuthClientListActionCellProps {
  clientId: string;
  onRemoveClientClick: (clientId: string) => void;
}

const OAuthClientListActionCell: React.FC<OAuthClientListActionCellProps> =
  function OAuthClientListActionCell(props: OAuthClientListActionCellProps) {
    const { clientId, onRemoveClientClick } = props;
    const navigate = useNavigate();
    const { themes } = useSystemConfig();

    const onEditClick = useCallback(() => {
      navigate(`./${clientId}/edit`);
    }, [navigate, clientId]);

    const onRemoveClick = useCallback(() => {
      onRemoveClientClick(clientId);
    }, [clientId, onRemoveClientClick]);

    return (
      <div className={styles.cellContent}>
        <ActionButton
          className={styles.cellAction}
          theme={themes.actionButton}
          onClick={onEditClick}
        >
          <FormattedMessage id="edit" />
        </ActionButton>
        <VerticalDivider className={styles.cellActionDivider} />
        <ActionButton
          className={styles.cellAction}
          theme={themes.actionButton}
          onClick={onRemoveClick}
        >
          <FormattedMessage id="remove" />
        </ActionButton>
      </div>
    );
  };

interface CORSConfigurationWidgetProps {
  form: AppConfigFormModel<FormState>;
}

const CORSConfigurationWidget: React.FC<CORSConfigurationWidgetProps> =
  function CORSConfigurationWidget(props) {
    const { state, setState } = props.form;

    const onAllowedOriginsChange = useCallback(
      (allowedOrigins: string[]) => {
        setState((state) => ({ ...state, allowedOrigins }));
      },
      [setState]
    );

    return (
      <Widget className={cn(styles.widget, styles.controlGroup)}>
        <WidgetTitle>
          <FormattedMessage id="ApplicationsConfigurationScreen.cors.title" />
        </WidgetTitle>
        <Text className={styles.description}>
          <FormattedMessage id="ApplicationsConfigurationScreen.cors.desc" />
        </Text>
        <FormTextFieldList
          className={styles.control}
          parentJSONPointer="/http"
          fieldName="allowed_origins"
          list={state.allowedOrigins}
          onListChange={onAllowedOriginsChange}
          addButtonLabelMessageID="add"
        />
      </Widget>
    );
  };

interface SessionConfigurationWidgetProps {
  form: AppConfigFormModel<FormState>;
}

const SessionConfigurationWidget: React.FC<SessionConfigurationWidgetProps> =
  function SessionConfigurationWidget(props: SessionConfigurationWidgetProps) {
    const { state, setState } = props.form;

    const { renderToString } = useContext(Context);

    const onPersistentCookieChange = useCallback(
      (_, value?: boolean) => {
        setState((state) => ({
          ...state,
          persistentCookie: value ?? false,
        }));
      },
      [setState]
    );

    const onSessionLifetimeSecondsChange = useCallback(
      (_, value?: string) => {
        setState((prev) => ({
          ...prev,
          sessionLifetimeSeconds: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    const onIdleTimeoutEnabledChange = useCallback(
      (_, value?: boolean) => {
        setState((state) => ({
          ...state,
          idleTimeoutEnabled: value ?? false,
        }));
      },
      [setState]
    );

    const onIdleTimeoutSecondsChange = useCallback(
      (_, value?: string) => {
        setState((prev) => ({
          ...prev,
          idleTimeoutSeconds: parseIntegerAllowLeadingZeros(value),
        }));
      },
      [setState]
    );

    return (
      <Widget className={cn(styles.widget, styles.controlGroup)}>
        <WidgetTitle id="cookie-session">
          <FormattedMessage id="SessionConfigurationWidget.title" />
        </WidgetTitle>
        <WidgetDescription>
          <FormattedMessage
            id="SessionConfigurationWidget.description"
            values={{
              // cookieDomain wil be empty only if authgear.yaml is updated manually
              domain: state.cookieDomain ?? state.publicOrigin,
            }}
          />
        </WidgetDescription>
        <Toggle
          className={styles.control}
          inlineLabel={true}
          label={renderToString(
            "SessionConfigurationWidget.persistent-cookie.label"
          )}
          checked={state.persistentCookie}
          onChange={onPersistentCookieChange}
        />
        <TextField
          className={styles.control}
          type="text"
          label={renderToString(
            "SessionConfigurationWidget.session-lifetime.label"
          )}
          value={state.sessionLifetimeSeconds?.toFixed(0) ?? ""}
          onChange={onSessionLifetimeSecondsChange}
        />
        <Toggle
          className={styles.control}
          inlineLabel={true}
          label={renderToString(
            "SessionConfigurationWidget.invalidate-session-after-idling.label"
          )}
          checked={state.idleTimeoutEnabled}
          onChange={onIdleTimeoutEnabledChange}
        />
        <TextField
          className={styles.control}
          type="text"
          disabled={!state.idleTimeoutEnabled}
          label={renderToString(
            "SessionConfigurationWidget.idle-timeout.label"
          )}
          value={state.idleTimeoutSeconds?.toFixed(0) ?? ""}
          onChange={onIdleTimeoutSecondsChange}
        />
      </Widget>
    );
  };

interface OAuthClientConfigurationContentProps {
  form: AppConfigFormModel<FormState>;
  oauthClientsMaximum: number;
  showNotification: (msg: string) => void;
}

const OAuthClientConfigurationContent: React.FC<OAuthClientConfigurationContentProps> =
  function OAuthClientConfigurationContent(props) {
    const {
      form,
      form: { state, setState },
      oauthClientsMaximum,
    } = props;
    const { renderToString } = useContext(Context);

    const oauthClientListColumns = useMemo(() => {
      return makeOAuthClientListColumns(renderToString);
    }, [renderToString]);

    const onRemoveClientClick = useCallback(
      (clientId: string) => {
        setState((state) => ({
          ...state,
          clients: state.clients.filter((c) => c.client_id !== clientId),
        }));
      },
      [setState]
    );

    const onRenderOAuthClientColumns = useCallback(
      (item?: OAuthClientConfig, _index?: number, column?: IColumn) => {
        if (item == null || column == null) {
          return null;
        }
        switch (column.key) {
          case "action":
            return (
              <OAuthClientListActionCell
                clientId={item.client_id}
                onRemoveClientClick={onRemoveClientClick}
              />
            );
          case "name":
            return (
              <span className={styles.cellContent}>{item.name ?? ""}</span>
            );
          case "clientId":
            return <OAuthClientIdCell clientId={item.client_id} />;
          default:
            return null;
        }
      },
      [onRemoveClientClick]
    );

    return (
      <ScreenContent className={styles.root}>
        <ScreenTitle>
          <FormattedMessage id="ApplicationsConfigurationScreen.title" />
        </ScreenTitle>
        <Widget className={cn(styles.widget, styles.controlGroup)}>
          <WidgetTitle>
            <FormattedMessage id="ApplicationsConfigurationScreen.title" />
          </WidgetTitle>
          <Text className={styles.description}>
            <FormattedMessage
              id="ApplicationsConfigurationScreen.client-endpoint.desc"
              values={{
                clientEndpoint: state.publicOrigin,
                dnsUrl: "../../custom-domains",
              }}
            />
          </Text>
          {oauthClientsMaximum < 99 && (
            <MessageBar>
              <FormattedMessage
                id="FeatureConfig.oauth-clients.maximum"
                values={{
                  planPagePath: "../../billing",
                  maximum: oauthClientsMaximum,
                }}
              />
            </MessageBar>
          )}
          <DetailsList
            className={styles.clientList}
            columns={oauthClientListColumns}
            items={state.clients}
            selectionMode={SelectionMode.none}
            onRenderItemColumn={onRenderOAuthClientColumns}
          />
        </Widget>
        <CORSConfigurationWidget form={form} />
        <SessionConfigurationWidget form={form} />
      </ScreenContent>
    );
  };

const ApplicationsConfigurationScreen: React.FC =
  function ApplicationsConfigurationScreen() {
    const { appID } = useParams();
    const { renderToString } = useContext(Context);
    const navigate = useNavigate();

    const form = useAppConfigForm(appID, constructFormState, constructConfig);
    const featureConfig = useAppFeatureConfigQuery(appID);

    const [messageBar, setMessageBar] = useState<React.ReactNode>(null);
    const showNotification = useCallback((msg: string) => {
      setMessageBar(
        <MessageBar onDismiss={() => setMessageBar(null)}>
          <p>{msg}</p>
        </MessageBar>
      );
    }, []);

    const oauthClientsMaximum = useMemo(() => {
      return featureConfig.effectiveFeatureConfig?.oauth?.client?.maximum ?? 99;
    }, [featureConfig.effectiveFeatureConfig?.oauth?.client?.maximum]);

    const limitReached = useMemo(() => {
      return form.state.clients.length >= oauthClientsMaximum;
    }, [oauthClientsMaximum, form.state.clients.length]);

    const commandBarFarItems: ICommandBarItemProps[] = useMemo(
      () => [
        {
          key: "save",
          text: renderToString(
            "ApplicationsConfigurationScreen.add-client-button"
          ),
          iconProps: { iconName: "CirclePlus" },
          onClick: () => navigate("./add"),
          className: limitReached ? styles.readOnly : undefined,
        },
      ],
      [navigate, renderToString, limitReached]
    );

    if (form.isLoading || featureConfig.loading) {
      return <ShowLoading />;
    }

    if (form.loadError) {
      return <ShowError error={form.loadError} onRetry={form.reload} />;
    }

    if (featureConfig.error) {
      return (
        <ShowError
          error={form.loadError}
          onRetry={() => {
            featureConfig.refetch().finally(() => {});
          }}
        />
      );
    }

    return (
      <FormContainer
        form={form}
        messageBar={messageBar}
        farItems={commandBarFarItems}
      >
        <OAuthClientConfigurationContent
          form={form}
          oauthClientsMaximum={oauthClientsMaximum}
          showNotification={showNotification}
        />
      </FormContainer>
    );
  };

export default ApplicationsConfigurationScreen;
