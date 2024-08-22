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
  Dialog,
  DialogFooter,
  IButtonStyles,
  IColumn,
  IconButton,
  IDetailsRowProps,
  IDetailsRowStyleProps,
  IDialogContentProps,
  MessageBar,
  SelectionMode,
  Text,
} from "@fluentui/react";
import { concatStyleSetsWithProps } from "@fluentui/merge-styles";
import cn from "classnames";
import { Context, FormattedMessage } from "@oursky/react-messageformat";
import { Link, useNavigate, useParams } from "react-router-dom";
import { produce } from "immer";

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
import FeatureDisabledMessageBar from "./FeatureDisabledMessageBar";
import { useSystemConfig } from "../../context/SystemConfigContext";
import Widget from "../../Widget";
import FormContainer from "../../FormContainer";
import PrimaryButton from "../../PrimaryButton";
import ActionButton from "../../ActionButton";
import ButtonWithLoading from "../../ButtonWithLoading";
import DefaultButton from "../../DefaultButton";
import { useOAuthClientForm } from "../../hook/useOAuthClientForm";

const COPY_ICON_STLYES: IButtonStyles = {
  root: { margin: "0 4px" },
  rootHovered: { backgroundColor: "#d8d6d3" },
  rootPressed: { backgroundColor: "#c2c0be" },
};

const cellActionButtonStyles: IButtonStyles = {
  root: { padding: 0 },
  label: { margin: 0 },
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
    {
      key: "action",
      fieldName: "action",
      name: renderToString(
        "ApplicationsConfigurationScreen.client-list.action"
      ),
      minWidth: 100,
      className: styles.columnHeader,
    },
  ];
}

interface OAuthClientIdCopyButtonProps {
  clientId: string;
}

const OAuthClientIdCopyButton: React.VFC<OAuthClientIdCopyButtonProps> =
  function OAuthClientIdCopyButton(props) {
    const { clientId } = props;
    const { copyButtonProps, Feedback } = useCopyFeedback({
      textToCopy: clientId,
    });

    return (
      <div>
        <IconButton {...copyButtonProps} styles={COPY_ICON_STLYES} />
        <Feedback />
      </div>
    );
  };

interface OAuthClientIdCellProps {
  clientId: string;
}

const OAuthClientIdCell: React.VFC<OAuthClientIdCellProps> =
  function OAuthClientIdCell({ clientId }: OAuthClientIdCellProps) {
    return (
      <>
        <span className={styles.cellContent}>{clientId}</span>
        <OAuthClientIdCopyButton clientId={clientId} />
      </>
    );
  };

interface ClientCardProps {
  name?: string;
  clientId: string;
  applicationType?: string;
}

const ClientCard: React.VFC<ClientCardProps> = (props) => {
  const { name, clientId, applicationType } = props;
  const { appID } = useParams() as { appID: string };
  const targetPath = `/project/${appID}/configuration/apps/${clientId}/edit`;

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
        <div className={styles.clientCardIdContainer}>
          <Text
            className={styles.clientCardContent}
            style={{ color: neutralSecondary }}
          >
            {clientId}
          </Text>
          <OAuthClientIdCopyButton clientId={clientId} />
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

const ClientCardList: React.VFC<ClientCardListProps> = (props) => {
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

const OAuthClientConfigurationContent: React.VFC<OAuthClientConfigurationContentProps> =
  function OAuthClientConfigurationContent(props) {
    const {
      form: { state, reload },
      oauthClientsMaximum,
    } = props;
    const navigate = useNavigate();
    const { themes } = useSystemConfig();
    const { renderToString } = useContext(Context);
    const { appID } = useParams() as { appID: string };

    const deleteForm = useOAuthClientForm(appID, null);

    const [isRemoveDialogVisible, setIsRemoveDialogVisible] = useState(false);

    const dialogContentProps: IDialogContentProps = useMemo(() => {
      return {
        title: renderToString(
          "ApplicationsConfigurationScreen.delete-client-dialog.title"
        ),
        subText: renderToString(
          "ApplicationsConfigurationScreen.delete-client-dialog.description"
        ),
      };
    }, [renderToString]);

    const limitReached = useMemo(() => {
      return state.clients.length >= oauthClientsMaximum;
    }, [oauthClientsMaximum, state.clients.length]);

    const oauthClientListColumns = useMemo(() => {
      return makeOAuthClientListColumns(renderToString);
    }, [renderToString]);

    const onAddClientButtonClick = useCallback(
      () => navigate("./add"),
      [navigate]
    );

    const showDialogAndSetRemoveClientByID = useCallback(
      (clientID) => {
        deleteForm.setState((state) => ({
          ...state,
          removeClientByID: clientID,
        }));
        setIsRemoveDialogVisible(true);
      },
      [deleteForm, setIsRemoveDialogVisible]
    );

    const dismissDialogAndResetRemoveClientByID = useCallback(() => {
      setIsRemoveDialogVisible(false);
      deleteForm.setState((state) => ({
        ...state,
        removeClientByID: undefined,
      }));
    }, [deleteForm, setIsRemoveDialogVisible]);

    const onConfirmRemove = useCallback(() => {
      deleteForm.save().then(
        () => {
          dismissDialogAndResetRemoveClientByID();
          reload();
        },
        () => {
          dismissDialogAndResetRemoveClientByID();
        }
      );
    }, [deleteForm, reload, dismissDialogAndResetRemoveClientByID]);

    const onRenderOAuthClientRow = useCallback(
      (props?: IDetailsRowProps) => {
        if (!props) {
          return null;
        }

        const { styles: stylesProp, ...rest } = props;

        const clientID = "client_id" in props.item && props.item.client_id;
        const targetPath =
          typeof clientID === "string"
            ? `/project/${appID}/configuration/apps/${clientID}/edit`
            : ".";

        const styles = (props: IDetailsRowStyleProps) =>
          concatStyleSetsWithProps(
            props,
            {
              // Reduce the cell height after adding copy button to the list
              cell: {
                paddingTop: 4,
                paddingBottom: 4,
              },
            },
            stylesProp
          );

        return (
          <Link to={targetPath}>
            <DetailsRow styles={styles} {...rest} />
          </Link>
        );
      },
      [appID]
    );

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
          case "action":
            return (
              <span className={styles.cellContent}>
                <ActionButton
                  id={getApplicationTypeMessageID(item.x_application_type)}
                  text={renderToString(
                    "ApplicationsConfigurationScreen.delete-client.label"
                  )}
                  theme={themes.destructive}
                  styles={cellActionButtonStyles}
                  onClick={(e) => {
                    e.preventDefault();
                    showDialogAndSetRemoveClientByID(item.client_id);
                  }}
                />
              </span>
            );
          default:
            return null;
        }
      },
      [renderToString, showDialogAndSetRemoveClientByID, themes.destructive]
    );

    return (
      <ScreenContent layout="list">
        <div className={styles.screenTitle}>
          <ScreenTitle className={styles.widget}>
            <FormattedMessage id="ApplicationsConfigurationScreen.title" />
          </ScreenTitle>
          <PrimaryButton
            text={renderToString(
              "ApplicationsConfigurationScreen.add-client-button"
            )}
            iconProps={{ iconName: "Add" }}
            onClick={onAddClientButtonClick}
            disabled={limitReached}
          />
        </div>
        <ScreenDescription className={styles.widget}>
          <FormattedMessage id="ApplicationsConfigurationScreen.description" />
        </ScreenDescription>
        <div className={styles.widget}>
          {oauthClientsMaximum < 99 ? (
            <FeatureDisabledMessageBar
              messageID="FeatureConfig.oauth-clients.maximum"
              messageValues={{ maximum: oauthClientsMaximum }}
            />
          ) : null}
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
        <Dialog
          hidden={!isRemoveDialogVisible}
          dialogContentProps={dialogContentProps}
          modalProps={{ isBlocking: deleteForm.isUpdating }}
          onDismiss={dismissDialogAndResetRemoveClientByID}
        >
          <DialogFooter>
            <ButtonWithLoading
              theme={themes.actionButton}
              loading={deleteForm.isUpdating}
              onClick={onConfirmRemove}
              disabled={!isRemoveDialogVisible}
              labelId="confirm"
            />
            <DefaultButton
              onClick={dismissDialogAndResetRemoveClientByID}
              disabled={deleteForm.isUpdating || !isRemoveDialogVisible}
              text={<FormattedMessage id="cancel" />}
            />
          </DialogFooter>
        </Dialog>
      </ScreenContent>
    );
  };

const ApplicationsConfigurationScreen: React.VFC =
  function ApplicationsConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
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
      <FormContainer
        form={form}
        messageBar={messageBar}
        hideFooterComponent={true}
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
