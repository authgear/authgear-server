import React, { useContext, useMemo, useCallback, useState } from "react";
import cn from "classnames";
import { useParams, useLocation, useNavigate } from "react-router-dom";
import {
  DotsVerticalIcon,
  DownloadIcon,
  PlusIcon,
  TrashIcon,
} from "@radix-ui/react-icons";
import {
  DropdownMenu,
  Heading,
  IconButton as RadixIconButton,
  Text,
} from "@radix-ui/themes";
import { FormattedMessage, Context } from "../../intl";
import ScreenContent from "../../ScreenContent";
import ShowLoading from "../../ShowLoading";
import ShowError from "../../ShowError";
import {
  useAppAndSecretConfigQuery,
  AppAndSecretConfigQueryResult,
} from "./query/appAndSecretConfigQuery";
import { formatDatetime } from "../../util/formatDatetime";
import { downloadStringAsFile } from "../../util/download";
import { startReauthentication } from "./Authenticated";
import { useLocationEffect } from "../../hook/useLocationEffect";
import { makeGraphQLEndpoint } from "../adminapi/apollo";
import styles from "./AdminAPIConfigurationScreen.module.css";
import ScreenLayoutScrollView from "../../ScreenLayoutScrollView";
import { useUpdateAppAndSecretConfigMutation } from "./mutations/updateAppAndSecretMutation";
import { useIsLoading, useLoading } from "../../hook/loading";
import { useProvideError } from "../../hook/error";
import { AppSecretKey } from "./globalTypes.generated";
import { useAppSecretVisitToken } from "./mutations/generateAppSecretVisitTokenMutation";
import ExternalLink, { DEFAULT_EXTERNAL_LINK_PROPS } from "../../ExternalLink";
import { useGenerateShortLivedAdminAPITokenMutation } from "./mutations/generateShortLivedAdminAPITokenMutation";
import { parseAPIErrors, parseRawError } from "../../error/parse";
import { APIError } from "../../error/error";
import ErrorRenderer from "../../ErrorRenderer";
import { TextField } from "../../components/v2/TextField/TextField";
import { Callout } from "../../components/v2/Callout/Callout";
import { PrimaryButton as RadixPrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import { ConfirmationDialog } from "../../components/v2/ConfirmationDialog/ConfirmationDialog";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";

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

interface SettingsSectionProps {
  title: React.ReactNode;
  children: React.ReactNode;
}

function SettingsSection({
  title,
  children,
}: SettingsSectionProps): React.ReactElement {
  return (
    <section className={styles.section}>
      <div className={styles.sectionInner}>
        <Heading as="h2" size="3" weight="medium" className={styles.sectionHeading}>
          {title}
        </Heading>
        <div className={styles.sectionContent}>{children}</div>
      </div>
    </section>
  );
}

interface ReadOnlyCopyFieldProps {
  label: React.ReactNode;
  value: string;
  placeholder?: string;
  hint?: React.ReactNode;
  truncate?: boolean;
}

function ReadOnlyCopyField({
  label,
  value,
  placeholder,
  hint,
  truncate,
}: ReadOnlyCopyFieldProps): React.ReactElement {
  const showCopy = value.length > 0;

  return (
    <TextField
      size="2"
      label={label}
      value={value}
      placeholder={placeholder}
      readOnly={true}
      hint={hint}
      inputClassName={
        truncate ? styles.readOnlyCopyFieldTruncate : undefined
      }
      suffixPlain={true}
      suffix={showCopy ? <CopyIconButton textToCopy={value} /> : undefined}
    />
  );
}

interface AdminAPIKeysTableProps {
  items: Item[];
  onDownload: (keyID: string) => void;
  onDelete: (keyID: string) => void;
  canDelete: boolean;
}

function AdminAPIKeysTable({
  items,
  onDownload,
  onDelete,
  canDelete,
}: AdminAPIKeysTableProps): React.ReactElement {
  return (
    <div className={styles.keysTableWrapper}>
      <div className={styles.keysTable}>
        <div className={styles.keysTableHeader}>
          <div className={styles.keysTableHeaderCellKeyId}>
            <FormattedMessage id="AdminAPIConfigurationScreen.column.key-id" />
          </div>
          <div className={styles.keysTableHeaderCellCreatedAt}>
            <FormattedMessage id="AdminAPIConfigurationScreen.column.created-at" />
          </div>
          <div className={styles.keysTableHeaderCellActions} aria-hidden={true} />
        </div>
        {items.map((item) => (
          <div key={item.keyID} className={styles.keysTableRow}>
            <div className={styles.keysTableCellKeyId}>
              <div className={styles.keysTableCellKeyIdInner}>
                <Text size="2" className={styles.keysTableCellKeyIdText}>
                  {item.keyID}
                </Text>
                <CopyIconButton textToCopy={item.keyID} />
              </div>
            </div>
            <div className={styles.keysTableCellCreatedAt}>
              <Text size="2" className="truncate">
                {item.createdAt ?? ""}
              </Text>
            </div>
            <div className={styles.keysTableCellActions}>
              <DropdownMenu.Root>
                <DropdownMenu.Trigger>
                  <RadixIconButton variant="soft" color="gray" size="2">
                    <DotsVerticalIcon width="1rem" height="1rem" />
                  </RadixIconButton>
                </DropdownMenu.Trigger>
                <DropdownMenu.Content align="end">
                  <DropdownMenu.Item
                    onSelect={() => {
                      onDownload(item.keyID);
                    }}
                  >
                    <DownloadIcon />
                    <FormattedMessage id="download" />
                  </DropdownMenu.Item>
                  <DropdownMenu.Item
                    color="red"
                    disabled={!canDelete}
                    onSelect={() => {
                      if (canDelete) {
                        onDelete(item.keyID);
                      }
                    }}
                  >
                    <TrashIcon />
                    <FormattedMessage id="delete" />
                  </DropdownMenu.Item>
                  {!canDelete ? (
                    <DropdownMenu.Label className={styles.deleteDisabledHint}>
                      <FormattedMessage id="AdminAPIConfigurationScreen.keys.delete.tooltip" />
                    </DropdownMenu.Label>
                  ) : null}
                </DropdownMenu.Content>
              </DropdownMenu.Root>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

const AdminAPIConfigurationScreenContent: React.VFC<AdminAPIConfigurationScreenContentProps> =
  function AdminAPIConfigurationScreenContent(props) {
    const { appID, secretToken, queryResult, generateKey, deleteKey } = props;
    const { locale, renderToString } = useContext(Context);
    const { effectiveAppConfig } = useAppAndSecretConfigQuery(appID);
    const isLoading = useIsLoading();
    const [deleteKeyID, setDeleteKeyID] = useState<string | null>(null);
    const [isDeleteDialogVisible, setIsDeleteDialogVisible] = useState(false);
    const [shortLivedAdminAPIToken, setShortLivedAdminAPIToken] = useState<
      string | null
    >(null);
    const [shortLivedAdminAPITokenError, setShortLivedAdminAPITokenError] =
      useState<APIError[]>([]);

    const {
      generateShortLivedAdminAPIToken,
      loading: generatingShortLivedAdminAPITokenLoading,
    } = useGenerateShortLivedAdminAPITokenMutation(appID);

    const publicOrigin = effectiveAppConfig?.http?.public_origin;
    const adminAPIEndpoint =
      publicOrigin != null ? publicOrigin + "/_api/admin/graphql" : "";
    const rawAppID = effectiveAppConfig?.id ?? "";

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

    const dismissDialogAndResetDeleteKeyID = useCallback(() => {
      setIsDeleteDialogVisible(false);
      setDeleteKeyID(null);
    }, []);

    const onDeleteDialogOpenChange = useCallback(
      (open: boolean) => {
        if (!open && !isLoading) {
          dismissDialogAndResetDeleteKeyID();
        }
      },
      [dismissDialogAndResetDeleteKeyID, isLoading]
    );

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

    const showDialogAndSetDeleteKeyID = useCallback((keyID: string) => {
      setDeleteKeyID(keyID);
      setIsDeleteDialogVisible(true);
    }, []);

    const apiEndpointDocHint = (
      <FormattedMessage
        id="AdminAPIConfigurationScreen.details.description"
        values={{
          // eslint-disable-next-line react/no-unstable-nested-components
          docLink: (chunks: React.ReactNode) => (
            <ExternalLink href="https://docs.authgear.com/reference/apis/admin-api">
              {chunks}
            </ExternalLink>
          ),
          // eslint-disable-next-line react/no-unstable-nested-components
          code: (chunks: React.ReactNode) => <code>{chunks}</code>,
        }}
      />
    );

    const graphiqlWarning = (
      <Callout
        type="warning"
        showCloseButton={false}
        text={
          <FormattedMessage
            id="AdminAPIConfigurationScreen.graphiql.warning"
            values={{
              // eslint-disable-next-line react/no-unstable-nested-components
              b: (chunks: React.ReactNode) => <strong>{chunks}</strong>,
            }}
          />
        }
      />
    );

    return (
      <>
        <ScreenLayoutScrollView>
          <ScreenContent>
            <div className={cn(styles.widget, styles.pageHeader)}>
              <h1 className={styles.pageTitle}>
                <FormattedMessage id="AdminAPIConfigurationScreen.title" />
              </h1>
              <Text
                as="p"
                size="2"
                color="gray"
                className={styles.pageDescription}
              >
                <FormattedMessage
                  id="AdminAPIConfigurationScreen.description"
                  values={{
                    // eslint-disable-next-line react/no-unstable-nested-components
                    b: (chunks: React.ReactNode) => <strong>{chunks}</strong>,
                  }}
                />
              </Text>
            </div>

            <div className={styles.sections}>
              <SettingsSection
                title={
                  <FormattedMessage id="AdminAPIConfigurationScreen.api-endpoint.section-title" />
                }
              >
                <ReadOnlyCopyField
                  label={
                    <FormattedMessage id="AdminAPIConfigurationScreen.graphql-url.label" />
                  }
                  value={adminAPIEndpoint}
                />
                <ReadOnlyCopyField
                  label={
                    <FormattedMessage id="AdminAPIConfigurationScreen.project-id.title" />
                  }
                  value={rawAppID}
                />
                <Text
                  as="p"
                  size="1"
                  color="gray"
                  className={styles.sectionDescription}
                >
                  {apiEndpointDocHint}
                </Text>
              </SettingsSection>

              <SettingsSection
                title={
                  <FormattedMessage id="AdminAPIConfigurationScreen.keys.title" />
                }
              >
                <AdminAPIKeysTable
                  items={items}
                  onDownload={downloadItem}
                  onDelete={showDialogAndSetDeleteKeyID}
                  canDelete={items.length > 1}
                />
                {items.length >= 2 ? (
                  <Callout
                    type="warning"
                    showCloseButton={false}
                    text={
                      <FormattedMessage id="AdminAPIConfigurationScreen.keys.generate.warning" />
                    }
                  />
                ) : (
                  <button
                    type="button"
                    className={styles.generateKeyButton}
                    onClick={() => {
                      generateKey().catch((e) => {
                        console.error(e);
                      });
                    }}
                    disabled={isLoading}
                  >
                    <PlusIcon width="1rem" height="1rem" />
                    <FormattedMessage id="AdminAPIConfigurationScreen.keys.generate.label" />
                  </button>
                )}

                <div className={styles.tokenRow}>
                  <div className={styles.tokenFieldGroup}>
                    <div className={styles.tokenInputRow}>
                      <div className={styles.tokenInputRowField}>
                        <ReadOnlyCopyField
                          label={
                            <FormattedMessage id="AdminAPIConfigurationScreen.short-lived-admin-api-token.label" />
                          }
                          value={shortLivedAdminAPIToken ?? ""}
                          truncate={true}
                          placeholder={renderToString(
                            "AdminAPIConfigurationScreen.short-lived-admin-api-token.generate.placeholder"
                          )}
                        />
                      </div>
                      <RadixPrimaryButton
                        size="2"
                        text={
                          <FormattedMessage id="AdminAPIConfigurationScreen.short-lived-admin-api-token.generate" />
                        }
                        onClick={onClickGenerateShortLivedAdminAPIToken}
                        disabled={generatingShortLivedAdminAPITokenLoading}
                        loading={generatingShortLivedAdminAPITokenLoading}
                      />
                    </div>
                    <Text as="p" size="1" color="gray">
                      <FormattedMessage id="AdminAPIConfigurationScreen.short-lived-admin-api-token.description" />
                    </Text>
                  </div>
                  {shortLivedAdminAPITokenFieldErrors != null ? (
                    <Callout
                      type="error"
                      showCloseButton={false}
                      text={shortLivedAdminAPITokenFieldErrors}
                    />
                  ) : null}
                </div>
              </SettingsSection>

              <SettingsSection
                title={
                  <FormattedMessage id="AdminAPIConfigurationScreen.graphiql.title" />
                }
              >
                {graphiqlWarning}
                <Text as="p" size="2" color="gray">
                  <FormattedMessage
                    id="AdminAPIConfigurationScreen.graphiql.description"
                    values={{
                      // eslint-disable-next-line react/no-unstable-nested-components
                      br: () => <br />,
                    }}
                  />
                </Text>
                <div>
                  <SecondaryButton
                    size="2"
                    text={
                      <FormattedMessage id="AdminAPIConfigurationScreen.graphiql.open" />
                    }
                    onClick={() => {
                      window.open(
                        graphqlEndpoint,
                        DEFAULT_EXTERNAL_LINK_PROPS.target,
                        DEFAULT_EXTERNAL_LINK_PROPS.rel
                      );
                    }}
                  />
                </div>
              </SettingsSection>
            </div>
          </ScreenContent>
        </ScreenLayoutScrollView>
        <ConfirmationDialog
          open={isDeleteDialogVisible}
          onOpenChange={onDeleteDialogOpenChange}
          title={
            <FormattedMessage id="AdminAPIConfigurationScreen.keys.delete-dialog.title" />
          }
          description={
            <FormattedMessage id="AdminAPIConfigurationScreen.keys.delete-dialog.message" />
          }
          confirmText={
            <FormattedMessage id="AdminAPIConfigurationScreen.keys.delete-dialog.confirm" />
          }
          cancelText={<FormattedMessage id="cancel" />}
          onConfirm={onConfirmDelete}
          onCancel={dismissDialogAndResetDeleteKeyID}
          loading={isLoading}
        />
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
