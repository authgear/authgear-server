import React, {
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  DetailsList,
  DetailsRow,
  IButtonStyles,
  IColumn,
  ICommandBarItemProps,
  IconButton,
  IDetailsRowProps,
  MessageBar,
  SelectionMode,
  Text,
} from "@fluentui/react";
import cn from "classnames";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Link, useNavigate, useParams } from "react-router-dom";
import produce from "immer";

import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { OAuthClientConfig, PortalAPIAppConfig } from "../../types";
import { clearEmptyObject } from "../../util/misc";
import {
  AppConfigFormModel,
  useAppConfigForm,
} from "../../hook/useAppConfigForm";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import styles from "./ApplicationsConfigurationScreen.module.css";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import { useAppFeatureConfigQuery } from "./query/appFeatureConfigQuery";
import ScreenDescription from "../../ScreenDescription";
import { getApplicationTypeMessageID } from "./EditOAuthClientForm";
import CommandBarContainer from "../../CommandBarContainer";
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import { onRenderCommandBarPrimaryButton } from "../../CommandBarPrimaryButton";
import { useSystemConfig } from "../../context/SystemConfigContext";
import Widget from "../../Widget";

const COPY_ICON_STLYES: IButtonStyles = {
  root: { margin: "0 4px" },
  rootHovered: { backgroundColor: "#d8d6d3" },
  rootPressed: { backgroundColor: "#c2c0be" },
};

interface FormState {
  clients: OAuthClientConfig[];
}

function constructFormState(config: PortalAPIAppConfig): FormState {
  return {
    clients: config.oauth?.clients ?? [],
  };
}

function constructConfig(
  config: PortalAPIAppConfig,
  _: FormState,
  currentState: FormState
): PortalAPIAppConfig {
  return produce(config, (config) => {
    config.oauth ??= {};
    config.oauth.clients = currentState.clients;
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
      key: "applicationType",
      fieldName: "applicationType",
      name: renderToString(
        "ApplicationsConfigurationScreen.client-list.application-type"
      ),
      minWidth: 250,
      className: styles.columnHeader,
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
        <div>
          <IconButton {...copyButtonProps} styles={COPY_ICON_STLYES} />
          <Feedback />
        </div>
      </>
    );
  };

interface ClientCardProps {
  name?: string;
  clientId: string;
  applicationType?: string;
}

const ClientCard: React.FC<ClientCardProps> = (props) => {
  const { name, clientId, applicationType } = props;
  const targetPath = `./${clientId}/edit`;

  const {
    themes: {
      main: {
        palette: { neutralSecondary },
      },
    },
  } = useSystemConfig();

  return (
    <Link to={targetPath}>
      <Widget>
        <Text className={styles.clientCardTitle}>{name}</Text>
        <div
          className={cn(styles.clientCardContent, styles.clientCardIdContainer)}
          style={{ color: neutralSecondary }}
        >
          <OAuthClientIdCell clientId={clientId} />
        </div>
        <Text
          className={styles.clientCardContent}
          style={{ color: neutralSecondary }}
        >
          <FormattedMessage id={getApplicationTypeMessageID(applicationType)} />
        </Text>
      </Widget>
    </Link>
  );
};

interface ClientCardListProps {
  className: string;
  items: OAuthClientConfig[];
}

const ClientCardList: React.FC<ClientCardListProps> = (props) => {
  const { className, items } = props;

  return (
    <div className={cn(styles.clientCardList, className)}>
      {items.map((card) => {
        return (
          <ClientCard
            key={card.client_id}
            name={card.name}
            clientId={card.client_id}
            applicationType={card.x_application_type}
          />
        );
      })}
    </div>
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
      form: { state },
      oauthClientsMaximum,
    } = props;
    const { renderToString } = useContext(Context);

    const oauthClientListColumns = useMemo(() => {
      return makeOAuthClientListColumns(renderToString);
    }, [renderToString]);

    const onRenderOAuthClientRow = useCallback((props?: IDetailsRowProps) => {
      if (!props) {
        return null;
      }

      const clientID = "client_id" in props.item && props.item.client_id;
      const targetPath =
        typeof clientID === "string" ? `./${clientID}/edit` : ".";
      props.styles = {
        ...props.styles,
        // Reduce the cell height after adding copy button to the list
        cell: {
          paddingTop: 4,
          paddingBottom: 4,
        },
      };
      return (
        <Link to={targetPath}>
          <DetailsRow {...props} />
        </Link>
      );
    }, []);

    const onRenderOAuthClientColumns = useCallback(
      (item?: OAuthClientConfig, _index?: number, column?: IColumn) => {
        if (item == null || column == null) {
          return null;
        }
        switch (column.key) {
          case "name":
            return (
              <span className={styles.cellContent}>{item.name ?? ""}</span>
            );
          case "clientId":
            return <OAuthClientIdCell clientId={item.client_id} />;
          case "applicationType":
            return (
              <span className={styles.cellContent}>
                <FormattedMessage
                  id={getApplicationTypeMessageID(item.x_application_type)}
                />
              </span>
            );
          default:
            return null;
        }
      },
      []
    );

    return (
      <ScreenContent>
        <ScreenTitle className={styles.widget}>
          <FormattedMessage id="ApplicationsConfigurationScreen.title" />
        </ScreenTitle>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="ApplicationsConfigurationScreen.description" />
        </ScreenDescription>
        <div className={styles.widget}>
          {oauthClientsMaximum < 99 && (
            <FeatureDisabledMessageBar>
              <FormattedMessage
                id="FeatureConfig.oauth-clients.maximum"
                values={{
                  planPagePath: "./../../billing",
                  maximum: oauthClientsMaximum,
                }}
              />
            </FeatureDisabledMessageBar>
          )}
          <div className={styles.desktopView}>
            <DetailsList
              onRenderRow={onRenderOAuthClientRow}
              className={styles.clientList}
              columns={oauthClientListColumns}
              items={state.clients}
              selectionMode={SelectionMode.none}
              onRenderItemColumn={onRenderOAuthClientColumns}
            />
          </div>
          <div className={styles.mobileView}>
            <ClientCardList
              className={styles.clientList}
              items={state.clients}
            />
          </div>
        </div>
      </ScreenContent>
    );
  };

const ApplicationsConfigurationScreen: React.FC =
  function ApplicationsConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const { renderToString } = useContext(Context);
    const navigate = useNavigate();

    const form = useAppConfigForm({
      appID,
      constructFormState,
      constructConfig,
    });
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

    const primaryItems: ICommandBarItemProps[] = useMemo(
      () => [
        {
          key: "add",
          text: renderToString(
            "ApplicationsConfigurationScreen.add-client-button"
          ),
          iconProps: { iconName: "CirclePlus" },
          onClick: () => navigate("./add"),
          disabled: limitReached,
          onRender: onRenderCommandBarPrimaryButton,
        },
      ],
      [navigate, renderToString, limitReached]
    );

    const isLoading = useMemo(
      () => form.isLoading || featureConfig.loading,
      [form.isLoading, featureConfig.loading]
    );

    const error = useMemo(
      () => form.loadError ?? featureConfig.error,
      [form.loadError, featureConfig.error]
    );

    const onRetry = useCallback(() => {
      if (form.loadError) {
        form.reload();
      }

      if (featureConfig.error) {
        featureConfig.refetch().finally(() => {});
      }
    }, [form, featureConfig]);

    useEffect(() => {
      if (!isLoading && !error && form.state.clients.length === 0) {
        navigate("./add", { replace: true });
      }
    }, [isLoading, error, form.state.clients.length, navigate]);

    if (isLoading) {
      return <ShowLoading />;
    }

    if (error) {
      return <ShowError error={error} onRetry={onRetry} />;
    }

    return (
      <CommandBarContainer messageBar={messageBar} primaryItems={primaryItems}>
        <OAuthClientConfigurationContent
          form={form}
          oauthClientsMaximum={oauthClientsMaximum}
          showNotification={showNotification}
        />
      </CommandBarContainer>
    );
  };

export default ApplicationsConfigurationScreen;
