import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  ActionButton,
  Callout,
  DetailsList,
  DirectionalHint,
  IButtonStyles,
  ICalloutContentStyleProps,
  ICalloutContentStyles,
  IColumn,
  ICommandBarItemProps,
  IconButton,
  IIconProps,
  IStyleFunctionOrObject,
  MessageBar,
  SelectionMode,
  Text,
  VerticalDivider,
} from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { useNavigate, useParams } from "react-router-dom";
import produce from "immer";
import cn from "classnames";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { OAuthClientConfig, PortalAPIAppConfig } from "../../types";
import { copyToClipboard } from "../../util/clipboard";
import { clearEmptyObject } from "../../util/misc";
import { useSystemConfig } from "../../context/SystemConfigContext";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import FormContainer from "../../FormContainer";

import styles from "./ApplicationsConfigurationScreen.module.scss";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import WidgetTitle from "../../WidgetTitle";
import Widget from "../../Widget";
import FormTextFieldList from "../../FormTextFieldList";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";

const COPY_ICON_PROPS: IIconProps = { iconName: "Copy" };
const COPY_ICON_STLYES: IButtonStyles = {
  root: { margin: 4 },
  rootHovered: { backgroundColor: "#d8d6d3" },
  rootPressed: { backgroundColor: "#c2c0be" },
};
const CALLOUT_STYLES: IStyleFunctionOrObject<
  ICalloutContentStyleProps,
  ICalloutContentStyles
> = { root: { padding: 8 } };

interface FormState {
  publicOrigin: string;
  clients: OAuthClientConfig[];
  allowedOrigins: string[];
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    publicOrigin: config.http?.public_origin ?? "",
    clients: config.oauth?.clients ?? [],
    allowedOrigins: config.http?.allowed_origins ?? [],
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _initialState: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.oauth ??= {};
    config.oauth.clients = currentState.clients;
    config.http ??= {};
    config.http.allowed_origins = currentState.allowedOrigins;
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
      minWidth: 150,
      className: styles.columnHeader,
    },

    {
      key: "clientId",
      fieldName: "clientId",
      name: renderToString(
        "ApplicationsConfigurationScreen.client-list.client-id"
      ),
      minWidth: 300,
      className: styles.columnHeader,
    },
    {
      key: "action",
      name: renderToString("action"),
      className: styles.columnHeader,
      minWidth: 200,
    },
  ];
}

function useDelayedAction(delayed: () => void): (delay: number) => void {
  // tuple is used instead of number because we want to trigger the effect even when delay argument is the same
  const [delay, setDelay] = useState<[number] | null>(null);

  useEffect(() => {
    if (!delay) {
      return () => {};
    }
    const timer = setTimeout(() => {
      delayed();
    }, delay[0]);
    return () => {
      clearTimeout(timer);
    };
  }, [delay, delayed]);

  return (delay: number) => setDelay([delay]);
}

interface OAuthClientIdCellProps {
  clientId: string;
  normalDismissTimeout?: number;
  quickDismissTimeout?: number;
}

const OAuthClientIdCell: React.FC<OAuthClientIdCellProps> =
  function OAuthClientIdCell(props) {
    const { clientId, normalDismissTimeout, quickDismissTimeout } = props;
    const copyButtonId = "oauth-client-id-copy-button";
    const [isCalloutVisible, setIsCalloutVisible] = useState(false);
    const { renderToString } = useContext(Context);
    const dismissCallout = useCallback(() => setIsCalloutVisible(false), []);
    const scheduleCalloutDismiss = useDelayedAction(dismissCallout);

    const onCopyClick = useCallback(() => {
      copyToClipboard(clientId);
      setIsCalloutVisible(true);
      scheduleCalloutDismiss(normalDismissTimeout ?? 2000);
    }, [clientId, normalDismissTimeout, scheduleCalloutDismiss]);

    const onMouseLeaveCopy = useCallback(() => {
      scheduleCalloutDismiss(quickDismissTimeout ?? 500);
    }, [quickDismissTimeout, scheduleCalloutDismiss]);

    return (
      <>
        <span className={styles.cellContent}>{clientId}</span>
        <IconButton
          id={copyButtonId}
          onClick={onCopyClick}
          iconProps={COPY_ICON_PROPS}
          title={renderToString("copy")}
          ariaLabel={renderToString("copy")}
          styles={COPY_ICON_STLYES}
          onMouseLeave={onMouseLeaveCopy}
        />
        {isCalloutVisible && (
          <Callout
            target={`#${copyButtonId}`}
            directionalHint={DirectionalHint.topCenter}
            onDismiss={dismissCallout}
            styles={CALLOUT_STYLES}
          >
            <Text variant="small">
              {renderToString(
                "ApplicationsConfigurationScreen.client-id-copied"
              )}
            </Text>
          </Callout>
        )}
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
