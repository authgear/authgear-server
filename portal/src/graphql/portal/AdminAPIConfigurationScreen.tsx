import React, { useContext, useMemo, useCallback, useState } from "react";
import { useParams, useLocation, useNavigate } from "react-router-dom";
import {
  DetailsList,
  IColumn,
  SelectionMode,
  MessageBar,
  MessageBarType,
  TooltipHost,
  Dialog,
  DialogFooter,
  IconButton,
  useTheme,
} from "@fluentui/react";
import cn from "classnames";
import { FormattedMessage, Context } from "../../intl";
import ScreenContent from "../../ScreenContent";
import ScreenTitle from "../../ScreenTitle";
import ScreenDescription from "../../ScreenDescription";
import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import {
  useAppAndSecretConfigQuery,
  AppAndSecretConfigQueryResult,
} from "./query/appAndSecretConfigQuery";
import { formatDatetime } from "../../util/formatDatetime";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { downloadStringAsFile } from "../../util/download";
import { startReauthentication } from "./Authenticated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { makeGraphQLEndpoint } from "../adminapi/apollo";
import styles from "./AdminAPIConfigurationScreen.module.css";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import PrimaryButton from "../../PrimaryButton";
import ActionButton from "../../ActionButton";
import DefaultButton from "../../DefaultButton";
import { useUpdateAppAndSecretConfigMutation } from "./mutations/updateAppAndSecretMutation";
import { useIsLoading, useLoading } from "../../hook/loading";
import { useProvideError } from "../../hook/error";
import { AppSecretKey } from "./globalTypes.generated";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import HorizontalDivider from "../../HorizontalDivider";
import TextFieldWithCopyButton from "../../TextFieldWithCopyButton";
import { DEFAULT_EXTERNAL_LINK_PROPS } from "../../ExternalLink";
import ExternalLink from "../../ExternalLink";
import { TextWithCopyButton } from "../../components/common/TextWithCopyButton";
import { useGenerateShortLivedAdminAPITokenMutation } from "./mutations/generateShortLivedAdminAPITokenMutation";
import { useCopyFeedback } from "../../hook/useCopyFeedback";
import TextField from "../../TextField";
import { parseAPIErrors, parseRawError } from "../../error/parse";
import { APIError } from "../../error/error";
import ErrorRenderer from "../../ErrorRenderer";

interface AdminAPIConfigurationScreenContentProps {
  appID: string;
  secretToken: string | null;
  queryResult: AppAndSecretConfigQueryResult;
  generateKey: () => Promise<void>;
  deleteKey: (keyID: string) => Promise<void>;
}

interface Item {
  keyID: string;
  createdAt: string | null;
  publicKeyPEM: string;
  privateKeyPEM?: string | null;
}

interface LocationState {
  keyID: string;
  shouldRefreshSecretToken: boolean;
  shouldGenerateShortLivedAdminAPIToken: boolean;
}

function isLocationState(raw: unknown): raw is LocationState {
  return (
    raw != null &&
    typeof raw === "object" &&
    typeof (raw as Partial<LocationState>).keyID === "string" &&
    typeof (raw as Partial<LocationState>).shouldRefreshSecretToken ===
    "boolean" &&
    typeof (raw as Partial<LocationState>)
      .shouldGenerateShortLivedAdminAPIToken === "boolean"
  );
}

const messageBarStyles = {
  root: {
    width: "auto",
  },
};

const EXAMPLE_QUERY = `# The Authgear Admin API follows GraphQL Cursor Connections Specification to handle pagination of results.
# Read more about the Connection model to understand the types like "Edge", "Node", and "Cursor":
# https://relay.dev/graphql/connections.htm
#
# Here is an example query of fetching a list of users with page size equal to 2.
query {
  users(first: 2) {
    pageInfo {
      hasNextPage
    }
    edges {
      cursor
      node {
        id
        createdAt
      }
    }
  }
}
`;

const AdminAPIConfigurationScreenContent: React.VFC<AdminAPIConfigurationScreenContentProps> =
  function AdminAPIConfigurationScreenContent(props) {
    const { appID, secretToken, queryResult, generateKey, deleteKey } = props;
    const { locale, renderToString } = useContext(Context);
    const { effectiveAppConfig } = useAppAndSecretConfigQuery(appID);
    const { themes } = useSystemConfig();
    const isLoading = useIsLoading();
    const [deleteKeyID, setDeleteKeyID] = useState<string | null>(null);
    const [isDeleteDialogVisible, setIsDeleteDialogVisible] = useState(false);
    const [shortLivedAdminAPIToken, setShortLivedAdminAPIToken] = useState<
      string | null
    >(null);
    const theme = useTheme();
    const [shortLivedAdminAPITokenError, setShortLivedAdminAPITokenError] =
      useState<APIError[]>([]);

    const {
      generateShortLivedAdminAPIToken,
      loading: generatingShortLivedAdminAPITokenLoading,
    } = useGenerateShortLivedAdminAPITokenMutation(appID);
    const { copyButtonProps, Feedback } = useCopyFeedback({
      textToCopy: shortLivedAdminAPIToken ?? "",
    });

    const publicOrigin = effectiveAppConfig?.http?.public_origin;
    const adminAPIEndpoint =
      publicOrigin != null ? publicOrigin + "/_api/admin/graphql" : "";
    const rawAppID = effectiveAppConfig?.id;

    const graphqlEndpoint = useMemo(() => {
      const base = makeGraphQLEndpoint(appID);
      return base + "?query=" + encodeURIComponent(EXAMPLE_QUERY);
    }, [appID]);

    const adminAPISecrets = useMemo(() => {
      return queryResult.secretConfig?.adminAPISecrets ?? [];
    }, [queryResult.secretConfig?.adminAPISecrets]);

    const items: Item[] = useMemo(() => {
      const items: Item[] = [];
      for (const adminAPISecret of adminAPISecrets) {
        items.push({
          keyID: adminAPISecret.keyID,
          createdAt: formatDatetime(locale, adminAPISecret.createdAt ?? null),
          publicKeyPEM: adminAPISecret.publicKeyPEM,
          privateKeyPEM: adminAPISecret.privateKeyPEM,
        });
      }
      return items;
    }, [locale, adminAPISecrets]);

    const navigate = useNavigate();

    const downloadItem = useCallback(
      (keyID: string) => {
        const item = items.find((a) => a.keyID === keyID);
        if (item == null) {
          return;
        }
        if (item.privateKeyPEM != null) {
          downloadStringAsFile({
            content: item.privateKeyPEM,
            mimeType: "application/x-pem-file",
            filename: `${item.keyID}.pem`,
          });
          return;
        }

        const state: LocationState = {
          keyID,
          shouldRefreshSecretToken: true,
          shouldGenerateShortLivedAdminAPIToken: false,
        };

        startReauthentication(navigate, state).catch((e) => {
          console.error(e);
        });
      },
      [navigate, items]
    );

    const generateShortLivedAdminAPITokenHandle = useCallback(async () => {
      if (secretToken == null) {
        // generateShortLivedAdminAPITokenHandle should be called only
        // when there is a secret token
        console.error("secret token should not be null");
        return;
      }
      const token = await generateShortLivedAdminAPIToken(secretToken);
      setShortLivedAdminAPIToken(token);
    }, [generateShortLivedAdminAPIToken, secretToken]);

    const onClickGenerateShortLivedAdminAPIToken = useCallback(() => {
      setShortLivedAdminAPIToken(null);
      setShortLivedAdminAPITokenError([]);

      const reauthentication = () => {
        const state: LocationState = {
          keyID: "",
          shouldRefreshSecretToken: true,
          shouldGenerateShortLivedAdminAPIToken: true,
        };

        startReauthentication(navigate, state).catch((e) => {
          console.error(e);
        });
      };
      // There are two possible cases:
      // 1. secretToken is null
      // 2. secretToken is not null but failed to generate short-lived admin API token
      // in both cases, reauthentication is needed
      if (secretToken != null) {
        generateShortLivedAdminAPITokenHandle().catch((e) => {
          const apiErrors = parseRawError(e);
          if (apiErrors.length > 0) {
            if (apiErrors[0].reason === "Forbidden") {
              reauthentication();
            } else {
              setShortLivedAdminAPITokenError(apiErrors);
            }
          }
        });
      } else {
        reauthentication();
      }
    }, [navigate, secretToken, generateShortLivedAdminAPITokenHandle]);

    const shortLivedAdminAPITokenFieldErrors = useMemo(() => {
      const { topErrors } = parseAPIErrors(
        shortLivedAdminAPITokenError,
        [],
        []
      );
      return topErrors.length > 0 ? <ErrorRenderer errors={topErrors} /> : null;
    }, [shortLivedAdminAPITokenError]);

    useLocationEffect((state: LocationState) => {
      if (state.keyID !== "") {
        downloadItem(state.keyID);
      }
      if (state.shouldGenerateShortLivedAdminAPIToken) {
        generateShortLivedAdminAPITokenHandle();
      }
    });

    const dialogContentProps = useMemo(() => {
      return {
        title: renderToString(
          "AdminAPIConfigurationScreen.keys.delete-dialog.title"
        ),
        subText: renderToString(
          "AdminAPIConfigurationScreen.keys.delete-dialog.message"
        ),
      };
    }, [renderToString]);

    const showDialogAndSetDeleteKeyID = useCallback(
      (keyID: string) => {
        setDeleteKeyID(keyID);
        setIsDeleteDialogVisible(true);
      },
      [setIsDeleteDialogVisible]
    );

    const dismissDialogAndResetDeleteKeyID = useCallback(() => {
      setIsDeleteDialogVisible(false);
      setDeleteKeyID(null);
    }, [setIsDeleteDialogVisible]);

    const onConfirmDelete = useCallback(() => {
      if (deleteKeyID == null) {
        return;
      }
      deleteKey(deleteKeyID)
        .catch((e) => {
          console.error(e);
        })
        .finally(dismissDialogAndResetDeleteKeyID);
    }, [deleteKey, deleteKeyID, dismissDialogAndResetDeleteKeyID]);

    const keyIDColumnOnRender = useCallback((item?: Item) => {
      return (
        <span className={cn("flex", "items-center", "h-full")}>
          <TextWithCopyButton text={item?.keyID ?? ""} />
        </span>
      );
    }, []);

    const createdAtColumnOnRender = useCallback((item?: Item) => {
      return (
        <span className={cn("flex", "items-center", "h-full")}>
          {item?.createdAt ?? ""}
        </span>
      );
    }, []);

    const actionColumnOnRender = useCallback(
      (item?: Item, index?: number) => {
        const deleteButtonID = `delete-button-${index}`;
        const calloutProps = {
          target: `#${deleteButtonID}`,
        };
        return (
          <section className={cn("flex", "items-center", "h-full")}>
            <ActionButton
              className={styles.actionButton}
              theme={themes.actionButton}
              onClick={(e: React.MouseEvent<unknown>) => {
                e.preventDefault();
                e.stopPropagation();
                if (item != null) {
                  downloadItem(item.keyID);
                }
              }}
              text={<FormattedMessage id="download" />}
            />
            {items.length > 1 ? (
              <ActionButton
                id={deleteButtonID}
                className={styles.actionButton}
                theme={themes.destructive}
                onClick={(e: React.MouseEvent<unknown>) => {
                  e.preventDefault();
                  e.stopPropagation();
                  if (item != null) {
                    showDialogAndSetDeleteKeyID(item.keyID);
                  }
                }}
                text={<FormattedMessage id="delete" />}
              />
            ) : (
              <TooltipHost
                content={
                  <FormattedMessage id="AdminAPIConfigurationScreen.keys.delete.tooltip" />
                }
                calloutProps={calloutProps}
              >
                <ActionButton
                  id={deleteButtonID}
                  className={styles.actionButton}
                  theme={themes.destructive}
                  disabled={true}
                  text={<FormattedMessage id="delete" />}
                />
              </TooltipHost>
            )}
          </section>
        );
      },
      [
        downloadItem,
        items.length,
        showDialogAndSetDeleteKeyID,
        themes.actionButton,
        themes.destructive,
      ]
    );

    const columns: IColumn[] = useMemo(() => {
      return [
        {
          key: "keyID",
          fieldName: "keyID",
          name: renderToString("AdminAPIConfigurationScreen.column.key-id"),
          minWidth: 150,
          onRender: keyIDColumnOnRender,
        },
        {
          key: "createdAt",
          fieldName: "createdAt",
          name: renderToString("AdminAPIConfigurationScreen.column.created-at"),
          minWidth: 220,
          onRender: createdAtColumnOnRender,
        },
        {
          key: "action",
          name: renderToString("action"),
          minWidth: 150,
          onRender: actionColumnOnRender,
        },
      ];
    }, [
      renderToString,
      keyIDColumnOnRender,
      createdAtColumnOnRender,
      actionColumnOnRender,
    ]);

    return (
      <>
        <ScreenLayoutScrollView>
          <ScreenContent>
            <ScreenTitle className={styles.widget}>
              <FormattedMessage id="AdminAPIConfigurationScreen.title" />
            </ScreenTitle>
            <ScreenDescription className={styles.widget}>
              <FormattedMessage
                id="AdminAPIConfigurationScreen.description"
                values={{
                  b: (chunks: React.ReactNode) => <b>{chunks}</b>,
                }}
              />
            </ScreenDescription>
            <Widget className={styles.widget}>
              <WidgetTitle>
                <FormattedMessage id="AdminAPIConfigurationScreen.details.title" />
              </WidgetTitle>
              <TextFieldWithCopyButton
                label={renderToString(
                  "AdminAPIConfigurationScreen.api-endpoint.title"
                )}
                value={adminAPIEndpoint}
                readOnly={true}
              />
              <TextFieldWithCopyButton
                label={renderToString(
                  "AdminAPIConfigurationScreen.project-id.title"
                )}
                value={rawAppID}
                readOnly={true}
              />
              <WidgetDescription>
                <FormattedMessage
                  id="AdminAPIConfigurationScreen.details.description"
                  values={{
                    docLink: (chunks: React.ReactNode) => (
                      <ExternalLink href="https://docs.authgear.com/customization/admin-api">
                        {chunks}
                      </ExternalLink>
                    ),
                    code: (chunks: React.ReactNode) => <code>{chunks}</code>,
                  }}
                />
              </WidgetDescription>
            </Widget>
            <HorizontalDivider className={styles.separator} />
            <Widget className={styles.widget}>
              <WidgetTitle>
                <FormattedMessage id="AdminAPIConfigurationScreen.keys.title" />
              </WidgetTitle>
              <DetailsList
                items={items}
                columns={columns}
                selectionMode={SelectionMode.none}
              />
              {items.length >= 2 ? (
                <MessageBar
                  messageBarType={MessageBarType.warning}
                  styles={messageBarStyles}
                >
                  <FormattedMessage id="AdminAPIConfigurationScreen.keys.generate.warning" />
                </MessageBar>
              ) : (
                <ActionButton
                  className={styles.actionButton}
                  theme={themes.actionButton}
                  iconProps={{
                    iconName: "CirclePlus",
                  }}
                  onClick={generateKey}
                  disabled={isLoading}
                  text={
                    <FormattedMessage
                      id={"AdminAPIConfigurationScreen.keys.generate.label"}
                    />
                  }
                />
              )}
              <div className={cn("flex", "flex-col", "gap-2")}>
                <div className={cn("flex", "flex-row", "gap-2")}>
                  <div className={cn("flex", "flex-row", "flex-1")}>
                    <TextField
                      className={cn("flex-1")}
                      type="text"
                      label={renderToString(
                        "AdminAPIConfigurationScreen.short-lived-admin-api-token.label"
                      )}
                      value={shortLivedAdminAPIToken ?? ""}
                      placeholder={renderToString(
                        "AdminAPIConfigurationScreen.short-lived-admin-api-token.generate.placeholder"
                      )}
                      readOnly={true}
                    />
                    {shortLivedAdminAPIToken ? (
                      <>
                        <IconButton
                          {...copyButtonProps}
                          styles={{
                            root: {
                              backgroundColor: theme.palette.neutralLight,
                            },
                          }}
                          className={cn("self-end")}
                          theme={themes.actionButton}
                        />
                        <Feedback />
                      </>
                    ) : null}
                  </div>
                  <PrimaryButton
                    className={cn("self-end")}
                    onClick={onClickGenerateShortLivedAdminAPIToken}
                    text={
                      <FormattedMessage id="AdminAPIConfigurationScreen.short-lived-admin-api-token.generate" />
                    }
                    disabled={generatingShortLivedAdminAPITokenLoading}
                  />
                </div>
                {shortLivedAdminAPITokenFieldErrors ? (
                  <MessageBar messageBarType={MessageBarType.error}>
                    {shortLivedAdminAPITokenFieldErrors}
                  </MessageBar>
                ) : null}
                <WidgetDescription>
                  <FormattedMessage id="AdminAPIConfigurationScreen.short-lived-admin-api-token.description" />
                </WidgetDescription>
              </div>
            </Widget>
            <HorizontalDivider className={styles.separator} />
            <Widget className={styles.widget}>
              <WidgetTitle>
                <FormattedMessage id="AdminAPIConfigurationScreen.graphiql.title" />
              </WidgetTitle>
              <MessageBar
                messageBarType={MessageBarType.warning}
                styles={messageBarStyles}
              >
                <FormattedMessage
                  id="AdminAPIConfigurationScreen.graphiql.warning"
                  values={{
                    b: (chunks: React.ReactNode) => <b>{chunks}</b>,
                  }}
                />
              </MessageBar>
              <WidgetDescription>
                <FormattedMessage
                  id="AdminAPIConfigurationScreen.graphiql.description"
                  values={{
                    graphqlEndpoint,
                    docLink: (chunks: React.ReactNode) => (
                      <ExternalLink href="https://docs.authgear.com/customization/admin-api">
                        {chunks}
                      </ExternalLink>
                    ),
                  }}
                />
              </WidgetDescription>
              <div>
                <DefaultButton
                  {...DEFAULT_EXTERNAL_LINK_PROPS}
                  href={graphqlEndpoint}
                  text={
                    <FormattedMessage id="AdminAPIConfigurationScreen.graphiql.open" />
                  }
                />
              </div>
            </Widget>
          </ScreenContent>
        </ScreenLayoutScrollView>
        <Dialog
          hidden={!isDeleteDialogVisible}
          dialogContentProps={dialogContentProps}
          onDismiss={dismissDialogAndResetDeleteKeyID}
        >
          <DialogFooter>
            <PrimaryButton
              theme={themes.destructive}
              onClick={onConfirmDelete}
              disabled={isLoading || !isDeleteDialogVisible}
              text={
                <FormattedMessage id="AdminAPIConfigurationScreen.keys.delete-dialog.confirm" />
              }
            />
            <DefaultButton
              onClick={dismissDialogAndResetDeleteKeyID}
              disabled={isLoading || !isDeleteDialogVisible}
              text={<FormattedMessage id="cancel" />}
            />
          </DialogFooter>
        </Dialog>
      </>
    );
  };

const AdminAPIConfigurationScreen1: React.VFC<{
  appID: string;
  secretToken: string | null;
}> = function AdminAPIConfigurationScreen1({ appID, secretToken }) {
  const queryResult = useAppAndSecretConfigQuery(appID, secretToken);
  const { refetch: refetchAppAndSecret } = queryResult;

  const { updateAppAndSecretConfig, loading, error } =
    useUpdateAppAndSecretConfigMutation(appID);
  useLoading(loading);
  useProvideError(error);

  const generateKey = useCallback(async () => {
    const appConfig = queryResult.rawAppConfig;
    if (appConfig == null) {
      return;
    }
    const generateKeyInstruction = {
      adminAPIAuthKey: {
        action: "generate",
      },
    };
    await updateAppAndSecretConfig({
      appConfig,
      secretConfigUpdateInstructions: generateKeyInstruction,
    });
    await refetchAppAndSecret();
  }, [queryResult.rawAppConfig, updateAppAndSecretConfig, refetchAppAndSecret]);

  const deleteKey = useCallback(
    async (deleteKeyID: string) => {
      const appConfig = queryResult.rawAppConfig;
      if (appConfig == null) {
        return;
      }
      const deleteKeyInstruction = {
        adminAPIAuthKey: {
          action: "delete",
          deleteData: {
            keyID: deleteKeyID,
          },
        },
      };
      await updateAppAndSecretConfig({
        appConfig,
        secretConfigUpdateInstructions: deleteKeyInstruction,
      });
      await refetchAppAndSecret();
    },
    [queryResult.rawAppConfig, updateAppAndSecretConfig, refetchAppAndSecret]
  );

  if (queryResult.isLoading) {
    return <ShowLoading />;
  }

  if (queryResult.loadError) {
    return (
      <ShowError error={queryResult.loadError} onRetry={queryResult.refetch} />
    );
  }

  return (
    <AdminAPIConfigurationScreenContent
      appID={appID}
      secretToken={secretToken}
      queryResult={queryResult}
      generateKey={generateKey}
      deleteKey={deleteKey}
    />
  );
};

const SECRETS = [AppSecretKey.AdminApiSecrets];

const AdminAPIConfigurationScreen: React.VFC =
  function AdminAPIConfigurationScreen() {
    const { appID } = useParams() as { appID: string };
    const location = useLocation();
    const [shouldRefreshToken] = useState<boolean>(() => {
      const { state } = location;
      if (isLocationState(state)) {
        return state.shouldRefreshSecretToken;
      }
      return false;
    });
    const { token, loading, error, retry } = useAppSecretVisitToken(
      appID,
      SECRETS,
      shouldRefreshToken
    );

    if (error) {
      return <ShowError error={error} onRetry={retry} />;
    }

    if (loading || token === undefined) {
      return <ShowLoading />;
    }

    return <AdminAPIConfigurationScreen1 appID={appID} secretToken={token} />;
  };

export default AdminAPIConfigurationScreen;
