import React, { useCallback, useContext, useMemo, useState } from "react";
import { produce } from "immer";
import { IconButton as RadixIconButton, Text } from "@radix-ui/themes";
import { TrashIcon } from "@radix-ui/react-icons";
import { DateTime } from "luxon";
import { Context, FormattedMessage } from "../../intl";
import { useParams, useNavigate, Link } from "react-router-dom";
import ExternalLink from "../../ExternalLink";
import PortalLink from "../../Link";

import { useEndpoints } from "../../hook/useEndpoints";

import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { TextField } from "../../components/v2/TextField/TextField";
import { TextFieldList } from "../../components/v2/TextFieldList/TextFieldList";
import { Toggle } from "../../components/v2/Toggle/Toggle";
import { Tooltip } from "../../components/v2/Tooltip/Tooltip";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import { PrimaryButton } from "../../components/v2/Button/PrimaryButton/PrimaryButton";
import { SecondaryButton } from "../../components/v2/Button/SecondaryButton/SecondaryButton";
import {
  ApplicationType,
  OAuthClientConfig,
  OAuthClientSecretKey,
} from "../../types";
import { ensureNonEmptyString } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import { Accordion } from "../../components/common/Accordion";
import { ClientSecretsHook } from "../../hook/useClientSecrets";
import { useStartReauthentication } from "../../graphql/portal/Authenticated";
import {
  DeleteClientSecretConfirmationDialog,
  DeleteClientSecretConfirmationDialogData,
} from "../../components/applications/DeleteClientSecretConfirmationDialog";
import { LocationState } from "./EditOAuthClientScreen";
import { makeValidationErrorCustomMessageIDRule } from "../../error/parse";
import { formatSeconds } from "../../util/formatDuration";
import styles from "./EditOAuthClientForm.module.css";

const MASKED_SECRET = "***************";

const CONTENT_CLASSNAME = "flex flex-col gap-4";

interface EditOAuthClientFormProps {
  publicOrigin: string;
  className?: string;
  clientConfig: OAuthClientConfig;
  customUIEnabled: boolean;
  app2appEnabled: boolean;
  onClientConfigChange: (newClientConfig: OAuthClientConfig) => void;
  clientSecretHook: ClientSecretsHook;
}

export function getApplicationTypeMessageID(key?: string): string {
  const messageIDMap: Record<string, string> = {
    spa: "oauth-client.application-type.spa",
    traditional_webapp: "oauth-client.application-type.traditional-webapp",
    native: "oauth-client.application-type.native",
    confidential: "oauth-client.application-type.confidential",
    third_party_app: "oauth-client.application-type.third-party-app",
    m2m: "oauth-client.application-type.m2m",
  };
  return key && messageIDMap[key]
    ? messageIDMap[key]
    : "oauth-client.application-type.unspecified";
}

export function getReducedClientConfig(
  clientConfig: OAuthClientConfig
): Omit<OAuthClientConfig, "grant_types" | "response_types"> {
  const {
    grant_types: _grant_types,
    response_types: _response_types,
    ...rest
  } = clientConfig;

  return {
    ...rest,
    post_logout_redirect_uris: rest.post_logout_redirect_uris ?? [],
    issue_jwt_access_token: rest.issue_jwt_access_token ?? false,
  };
}

export function updateClientConfig<K extends keyof OAuthClientConfig>(
  clientConfig: OAuthClientConfig,
  field: K,
  newValue: OAuthClientConfig[K]
): OAuthClientConfig {
  return produce(clientConfig, (draftConfig) => {
    draftConfig[field] = newValue;
  });
}

const parentJSONPointer = /\/oauth\/clients\/\d+/;

interface CopyFieldProps {
  label: string;
  value: string;
  suffix?: React.ReactNode;
}

// A read-only v2 TextField with a copy button (or a custom suffix),
// replacing the FluentUI TextFieldWithCopyButton.
function CopyField({ label, value, suffix }: CopyFieldProps) {
  return (
    <TextField
      size="2"
      label={label}
      value={value}
      readOnly={true}
      suffixPlain={true}
      suffix={suffix ?? <CopyIconButton textToCopy={value} />}
    />
  );
}

interface ToggleWithDescriptionProps {
  checked?: boolean;
  disabled?: boolean;
  onCheckedChange: (checked: boolean) => void;
  text?: React.ReactNode;
  description?: React.ReactNode;
}

function ToggleWithDescription({
  checked,
  disabled,
  onCheckedChange,
  text,
  description,
}: ToggleWithDescriptionProps) {
  return (
    <div className="flex flex-col gap-1">
      <Toggle
        checked={checked}
        disabled={disabled}
        onCheckedChange={onCheckedChange}
        text={text}
      />
      {description != null ? (
        <Text as="p" size="1" color="gray">
          {description}
        </Text>
      ) : null}
    </div>
  );
}

function HelpText(props: { children: React.ReactNode }) {
  const { children } = props;
  return (
    <Text as="p" size="2" className={styles.helpText}>
      {children}
    </Text>
  );
}

const EditOAuthClientForm: React.VFC<EditOAuthClientFormProps> =
  function EditOAuthClientForm(props: EditOAuthClientFormProps) {
    const {
      className,
      clientConfig,
      publicOrigin,
      customUIEnabled,
      app2appEnabled,
      onClientConfigChange,
      clientSecretHook,
    } = props;

    const { renderToString, locale } = useContext(Context);

    const { appID } = useParams() as { appID: string };

    const { startReauthentication, isRevealing } =
      useStartReauthentication<LocationState>();

    const [deleteClientSecretDialogData, setDeleteClientSecretDialogData] =
      useState<DeleteClientSecretConfirmationDialogData | null>(null);

    const onNameChange: React.ChangeEventHandler<HTMLInputElement> =
      useCallback(
        (e) => {
          onClientConfigChange(
            updateClientConfig(
              clientConfig,
              "name",
              ensureNonEmptyString(e.currentTarget.value)
            )
          );
        },
        [clientConfig, onClientConfigChange]
      );

    const onClientNameChange: React.ChangeEventHandler<HTMLInputElement> =
      useCallback(
        (e) => {
          onClientConfigChange(
            updateClientConfig(
              clientConfig,
              "client_name",
              ensureNonEmptyString(e.currentTarget.value)
            )
          );
        },
        [clientConfig, onClientConfigChange]
      );

    const onAccessTokenLifetimeChange: React.ChangeEventHandler<HTMLInputElement> =
      useCallback(
        (e) => {
          onClientConfigChange(
            updateClientConfig(
              clientConfig,
              "access_token_lifetime_seconds",
              parseIntegerAllowLeadingZeros(e.currentTarget.value)
            )
          );
        },
        [clientConfig, onClientConfigChange]
      );

    const onRefreshTokenLifetimeChange: React.ChangeEventHandler<HTMLInputElement> =
      useCallback(
        (e) => {
          onClientConfigChange(
            updateClientConfig(
              clientConfig,
              "refresh_token_lifetime_seconds",
              parseIntegerAllowLeadingZeros(e.currentTarget.value)
            )
          );
        },
        [clientConfig, onClientConfigChange]
      );

    const onIdleTimeoutChange: React.ChangeEventHandler<HTMLInputElement> =
      useCallback(
        (e) => {
          onClientConfigChange(
            updateClientConfig(
              clientConfig,
              "refresh_token_idle_timeout_seconds",
              parseIntegerAllowLeadingZeros(e.currentTarget.value)
            )
          );
        },
        [clientConfig, onClientConfigChange]
      );

    const setRedirectUris = useCallback(
      (list: string[]) => {
        onClientConfigChange(
          updateClientConfig(clientConfig, "redirect_uris", list)
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onRedirectUriAdd = useCallback(
      (list: string[], item: string) => {
        setRedirectUris([...list, item]);
      },
      [setRedirectUris]
    );

    const onRedirectUriChange = useCallback(
      (list: string[], index: number, item: string) => {
        setRedirectUris(list.map((value, i) => (i === index ? item : value)));
      },
      [setRedirectUris]
    );

    const onRedirectUriDelete = useCallback(
      (list: string[], index: number) => {
        setRedirectUris(list.filter((_, i) => i !== index));
      },
      [setRedirectUris]
    );

    const setPostLogoutRedirectUris = useCallback(
      (list: string[]) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "post_logout_redirect_uris",
            list.length > 0 ? list : undefined
          )
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onPostLogoutRedirectUriAdd = useCallback(
      (list: string[], item: string) => {
        setPostLogoutRedirectUris([...list, item]);
      },
      [setPostLogoutRedirectUris]
    );

    const onPostLogoutRedirectUriChange = useCallback(
      (list: string[], index: number, item: string) => {
        setPostLogoutRedirectUris(
          list.map((value, i) => (i === index ? item : value))
        );
      },
      [setPostLogoutRedirectUris]
    );

    const onPostLogoutRedirectUriDelete = useCallback(
      (list: string[], index: number) => {
        setPostLogoutRedirectUris(list.filter((_, i) => i !== index));
      },
      [setPostLogoutRedirectUris]
    );

    const onChangeRefreshTokenIdleTimeoutEnabled = useCallback(
      (checked: boolean) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "refresh_token_idle_timeout_enabled",
            checked
          )
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onChangeExpireWhenLoginOnOtherDevice = useCallback(
      (checked: boolean) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "x_max_concurrent_session",
            checked ? 1 : undefined
          )
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onChangeRefreshTokenRotationEnabled = useCallback(
      (checked: boolean) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "refresh_token_rotation_enabled",
            checked
          )
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onIssueJWTAccessTokenChange = useCallback(
      (checked: boolean) => {
        onClientConfigChange(
          updateClientConfig(clientConfig, "issue_jwt_access_token", checked)
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onChangeSenderConstraining = useCallback(
      (checked: boolean) => {
        onClientConfigChange(
          updateClientConfig(clientConfig, "x_dpop_disabled", !checked)
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onApp2AppEnabledChange = useCallback(
      (checked: boolean) => {
        onClientConfigChange(
          updateClientConfig(clientConfig, "x_app2app_enabled", checked)
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onApp2AppMigrationChange = useCallback(
      (checked: boolean) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "x_app2app_insecure_device_key_binding_enabled",
            checked
          )
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onPolicyURIChange: React.ChangeEventHandler<HTMLInputElement> =
      useCallback(
        (e) => {
          onClientConfigChange(
            updateClientConfig(
              clientConfig,
              "policy_uri",
              ensureNonEmptyString(e.currentTarget.value)
            )
          );
        },
        [clientConfig, onClientConfigChange]
      );

    const onTOSURIChange: React.ChangeEventHandler<HTMLInputElement> =
      useCallback(
        (e) => {
          onClientConfigChange(
            updateClientConfig(
              clientConfig,
              "tos_uri",
              ensureNonEmptyString(e.currentTarget.value)
            )
          );
        },
        [clientConfig, onClientConfigChange]
      );

    const onCustomUIURI: React.ChangeEventHandler<HTMLInputElement> =
      useCallback(
        (e) => {
          onClientConfigChange(
            updateClientConfig(
              clientConfig,
              "x_custom_ui_uri",
              ensureNonEmptyString(e.currentTarget.value)
            )
          );
        },
        [clientConfig, onClientConfigChange]
      );

    const onGenerateClientSecretClick = useCallback(() => {
      void clientSecretHook.generate(clientConfig.client_id);
    }, [clientSecretHook, clientConfig.client_id]);

    const onDeleteClientSecretClick = useCallback(
      (keyItem: OAuthClientSecretKey) => {
        setDeleteClientSecretDialogData({ clientSecret: keyItem });
      },
      []
    );

    const navigate = useNavigate();
    const onRevealSecretClick = useCallback(() => {
      startReauthentication(navigate, { isClientSecretRevealed: true });
    }, [startReauthentication, navigate]);

    const onConfirmDeleteClientSecret = useCallback(async () => {
      if (deleteClientSecretDialogData == null) {
        return;
      }
      await clientSecretHook.delete(
        clientConfig.client_id,
        deleteClientSecretDialogData.clientSecret.keyID
      );
      setDeleteClientSecretDialogData(null);
    }, [
      clientSecretHook,
      clientConfig.client_id,
      deleteClientSecretDialogData,
    ]);

    const onDismissDeleteClientSecret = useCallback(() => {
      setDeleteClientSecretDialogData(null);
    }, []);

    const applicationTypeLabel = useMemo(() => {
      const messageID = getApplicationTypeMessageID(
        clientConfig.x_application_type
      );
      return renderToString(messageID);
    }, [clientConfig.x_application_type, renderToString]);

    const redirectURIsDescription = useMemo(() => {
      const messageIdMap: Record<ApplicationType, string | undefined> = {
        spa: "EditOAuthClientForm.redirect-uris.description.spa",
        traditional_webapp:
          "EditOAuthClientForm.redirect-uris.description.traditional-webapp",
        native: "EditOAuthClientForm.redirect-uris.description.native",
        confidential:
          "EditOAuthClientForm.redirect-uris.description.confidential",
        third_party_app:
          "EditOAuthClientForm.redirect-uris.description.third-party-app",
        m2m: undefined,
      };
      const messageID = clientConfig.x_application_type
        ? messageIdMap[clientConfig.x_application_type]
        : "EditOAuthClientForm.redirect-uris.description.unspecified";
      return messageID ? renderToString(messageID) : undefined;
    }, [clientConfig.x_application_type, renderToString]);

    const showPostLogoutRedirectURIsSettings = useMemo(
      () =>
        !clientConfig.x_application_type ||
        clientConfig.x_application_type === "spa" ||
        clientConfig.x_application_type === "traditional_webapp" ||
        clientConfig.x_application_type === "confidential",
      [clientConfig.x_application_type]
    );

    const showCookieSettings = useMemo(
      () =>
        clientConfig.x_application_type === "traditional_webapp" ||
        clientConfig.x_application_type === "third_party_app",
      [clientConfig.x_application_type]
    );

    const showRefreshTokenSettings = useMemo(
      () =>
        !clientConfig.x_application_type ||
        clientConfig.x_application_type === "spa" ||
        clientConfig.x_application_type === "native" ||
        clientConfig.x_application_type === "confidential" ||
        clientConfig.x_application_type === "third_party_app",
      [clientConfig.x_application_type]
    );

    const showDPoPSettings = useMemo(() => {
      if (!showRefreshTokenSettings) {
        return false;
      }
      return clientConfig.x_application_type === "native";
    }, [clientConfig.x_application_type, showRefreshTokenSettings]);

    const showAccessTokenSettings = useMemo(() => {
      if (showRefreshTokenSettings) {
        return true;
      }
      return (["m2m"] as OAuthClientConfig["x_application_type"][]).includes(
        clientConfig.x_application_type
      );
    }, [clientConfig.x_application_type, showRefreshTokenSettings]);

    const alwaysIssueJWTAccessTokenTooltipMessageID = useMemo(() => {
      const map: Map<OAuthClientConfig["x_application_type"], string | null> =
        new Map([
          ["m2m", "EditOAuthClientForm.issue-jwt-access-token.tooltip-m2m"],
        ]);
      return map.get(clientConfig.x_application_type) ?? null;
    }, [clientConfig.x_application_type]);

    const isIssueJWTAccessTokenToggleDisabled = useMemo(() => {
      return alwaysIssueJWTAccessTokenTooltipMessageID != null;
    }, [alwaysIssueJWTAccessTokenTooltipMessageID]);

    const showApp2AppSettings =
      clientConfig.x_application_type === "native" && app2appEnabled;

    const showConsentScreenSettings = useMemo(
      () => clientConfig.x_application_type === "third_party_app",
      [clientConfig.x_application_type]
    );

    const customUISupported = useMemo(
      () =>
        (
          [
            "spa",
            "native",
            "confidential",
            "third_party_app",
            "traditional_webapp",
            undefined,
          ] as OAuthClientConfig["x_application_type"][]
        ).includes(clientConfig.x_application_type),
      [clientConfig.x_application_type]
    );

    const showCustomUISettings = useMemo(
      () => customUIEnabled && customUISupported,
      [customUIEnabled, customUISupported]
    );

    const showClientSecret = useMemo(
      () =>
        (
          [
            "confidential",
            "third_party_app",
            "m2m",
          ] as OAuthClientConfig["x_application_type"][]
        ).includes(clientConfig.x_application_type),
      [clientConfig.x_application_type]
    );

    const refreshTokenHelpText = useMemo(() => {
      if (clientConfig.refresh_token_idle_timeout_enabled) {
        return renderToString(
          "EditOAuthClientForm.refresh-token.help-text.idle-timeout-enabled",
          {
            refreshTokenLifetimeFormattedDuration:
              clientConfig.refresh_token_lifetime_seconds != null
                ? formatSeconds(
                    locale,
                    clientConfig.refresh_token_lifetime_seconds
                  ) ?? ""
                : "",
            refreshTokenIdleTimeoutFormattedDuration:
              clientConfig.refresh_token_idle_timeout_seconds != null
                ? formatSeconds(
                    locale,
                    clientConfig.refresh_token_idle_timeout_seconds
                  ) ?? ""
                : "",
          }
        );
      }
      return renderToString(
        "EditOAuthClientForm.refresh-token.help-text.idle-timeout-disabled",
        {
          refreshTokenLifetimeFormattedDuration:
            clientConfig.refresh_token_lifetime_seconds != null
              ? formatSeconds(
                  locale,
                  clientConfig.refresh_token_lifetime_seconds
                ) ?? ""
              : "",
        }
      );
    }, [
      locale,
      clientConfig.refresh_token_lifetime_seconds,
      clientConfig.refresh_token_idle_timeout_enabled,
      clientConfig.refresh_token_idle_timeout_seconds,
      renderToString,
    ]);

    const showEndpointsSection = useMemo(
      () =>
        (
          [
            "confidential",
            "third_party_app",
            "m2m",
          ] as OAuthClientConfig["x_application_type"][]
        ).includes(clientConfig.x_application_type),
      [clientConfig.x_application_type]
    );

    const endpoints = useEndpoints(
      publicOrigin,
      clientConfig.x_application_type
    );

    const endpointsWithLabelIDs = useMemo(
      () => [
        {
          endpoint: endpoints.openidConfiguration,
          labelMessageID:
            "EditOAuthClientForm.openid-configuration-endpoint.label",
        },
        {
          endpoint: endpoints.authorize,
          labelMessageID: "EditOAuthClientForm.authorization-endpoint.label",
        },
        {
          endpoint: endpoints.token,
          labelMessageID: "EditOAuthClientForm.token-endpoint.label",
        },
        {
          endpoint: endpoints.userinfo,
          labelMessageID: "EditOAuthClientForm.userinfo-endpoint.label",
        },
        {
          endpoint: endpoints.endSession,
          labelMessageID: "EditOAuthClientForm.end-session-endpoint.label",
        },
        {
          endpoint: endpoints.jwksUri,
          labelMessageID: "EditOAuthClientForm.jwks-uri.label",
        },
      ],
      [endpoints]
    );

    const showURIsSection =
      redirectURIsDescription != null || showPostLogoutRedirectURIsSettings;

    const clientSecrets = useMemo(() => {
      return clientSecretHook.oauthClientSecrets.find(
        (item) => item.clientID === clientConfig.client_id
      )?.keys;
    }, [clientConfig.client_id, clientSecretHook.oauthClientSecrets]);

    const issueJWTAccessTokenToggle = (
      <Toggle
        checked={clientConfig.issue_jwt_access_token}
        disabled={isIssueJWTAccessTokenToggleDisabled}
        onCheckedChange={onIssueJWTAccessTokenChange}
        text={renderToString(
          "EditOAuthClientForm.issue-jwt-access-token.label"
        )}
      />
    );

    return (
      <>
        <SettingsSectionCard
          className={className}
          contentClassName={CONTENT_CLASSNAME}
          title={<FormattedMessage id="EditOAuthClientForm.basic-info.title" />}
        >
          <TextField
            size="2"
            parentJSONPointer={parentJSONPointer}
            fieldName="name"
            label={renderToString("EditOAuthClientForm.name.label")}
            value={clientConfig.name ?? ""}
            onChange={onNameChange}
            required={true}
          />
          <CopyField
            label={renderToString("EditOAuthClientForm.client-id.label")}
            value={clientConfig.client_id}
          />
          <CopyField
            label={renderToString("EditOAuthClientForm.endpoint.label")}
            value={publicOrigin}
          />
          <TextField
            size="2"
            label={renderToString("EditOAuthClientForm.application-type.label")}
            value={applicationTypeLabel}
            readOnly={true}
          />
        </SettingsSectionCard>

        {showClientSecret && clientSecrets && clientSecrets.length > 0 ? (
          <SettingsSectionCard
            className={className}
            contentClassName={CONTENT_CLASSNAME}
            title={
              <FormattedMessage id="EditOAuthClientForm.client-secrets.title" />
            }
          >
            {clientSecrets.map((keyItem) => {
              const showCopyButton = !!keyItem.key;
              const showDeleteButton = clientSecrets.length >= 2;
              return (
                <div key={keyItem.keyID} className="flex flex-col gap-1">
                  <TextField
                    size="2"
                    label={renderToString(
                      "EditOAuthClientForm.client-secret.label"
                    )}
                    value={keyItem.key ? keyItem.key : MASKED_SECRET}
                    readOnly={true}
                    suffixPlain={true}
                    suffix={
                      showCopyButton || showDeleteButton ? (
                        <div className="flex flex-row items-center gap-2">
                          {showCopyButton ? (
                            <CopyIconButton textToCopy={keyItem.key} />
                          ) : null}
                          {showDeleteButton ? (
                            <RadixIconButton
                              type="button"
                              variant="ghost"
                              color="red"
                              size="1"
                              aria-label={renderToString("delete")}
                              disabled={
                                clientSecretHook.isLoading ||
                                clientSecretHook.isUpdating
                              }
                              onClick={() => {
                                onDeleteClientSecretClick(keyItem);
                              }}
                            >
                              <TrashIcon width="1rem" height="1rem" />
                            </RadixIconButton>
                          ) : null}
                        </div>
                      ) : undefined
                    }
                  />
                  {keyItem.createdAt != null ? (
                    <Text as="p" size="1" color="gray">
                      <FormattedMessage
                        id="EditOAuthClientForm.client-secret.created-at"
                        values={{
                          datetime: DateTime.fromISO(
                            keyItem.createdAt
                          ).toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS),
                        }}
                      />
                    </Text>
                  ) : null}
                </div>
              );
            })}
            <div className="flex flex-row space-x-4">
              <PrimaryButton
                size="2"
                onClick={onRevealSecretClick}
                disabled={clientSecrets.every((item) => !!item.key)}
                loading={isRevealing}
                text={<FormattedMessage id="reveal" />}
              />
              {clientSecrets.length < 2 ? (
                <SecondaryButton
                  size="2"
                  text={renderToString(
                    "EditOAuthClientForm.client-secrets.create-new-secret"
                  )}
                  onClick={onGenerateClientSecretClick}
                  disabled={
                    clientSecretHook.isLoading || clientSecretHook.isUpdating
                  }
                />
              ) : null}
            </div>
          </SettingsSectionCard>
        ) : null}

        {showURIsSection ? (
          <SettingsSectionCard
            className={className}
            contentClassName={CONTENT_CLASSNAME}
            title={
              <span id="uris">
                <FormattedMessage id="EditOAuthClientForm.uris.title" />
              </span>
            }
          >
            {redirectURIsDescription != null ? (
              <TextFieldList
                parentJSONPointer={parentJSONPointer}
                fieldName="redirect_uris"
                list={clientConfig.redirect_uris ?? []}
                onListItemAdd={onRedirectUriAdd}
                onListItemChange={onRedirectUriChange}
                onListItemDelete={onRedirectUriDelete}
                addButtonLabelMessageID="EditOAuthClientForm.add-uri"
                label={
                  <FormattedMessage id="EditOAuthClientForm.redirect-uris.label" />
                }
                description={redirectURIsDescription}
              />
            ) : null}
            {showPostLogoutRedirectURIsSettings ? (
              <Accordion
                text={
                  <FormattedMessage id="EditOAuthClientForm.more-options" />
                }
              >
                <TextFieldList
                  parentJSONPointer={parentJSONPointer}
                  fieldName="post_logout_redirect_uris"
                  list={clientConfig.post_logout_redirect_uris ?? []}
                  onListItemAdd={onPostLogoutRedirectUriAdd}
                  onListItemChange={onPostLogoutRedirectUriChange}
                  onListItemDelete={onPostLogoutRedirectUriDelete}
                  addButtonLabelMessageID="EditOAuthClientForm.add-uri"
                  label={
                    <FormattedMessage id="EditOAuthClientForm.post-logout-redirect-uris.label" />
                  }
                  description={renderToString(
                    clientConfig.x_application_type === "spa"
                      ? "EditOAuthClientForm.post-logout-redirect-uris.spa.description"
                      : "EditOAuthClientForm.post-logout-redirect-uris.description"
                  )}
                />
              </Accordion>
            ) : null}
          </SettingsSectionCard>
        ) : null}

        {showConsentScreenSettings ? (
          <SettingsSectionCard
            className={className}
            contentClassName={CONTENT_CLASSNAME}
            title={
              <FormattedMessage id="EditOAuthClientForm.consent-screen.title" />
            }
          >
            <TextField
              size="2"
              parentJSONPointer={parentJSONPointer}
              fieldName="client_name"
              label={renderToString("EditOAuthClientForm.client-name.label")}
              hint={renderToString(
                "EditOAuthClientForm.client-name.description"
              )}
              value={clientConfig.client_name ?? ""}
              onChange={onClientNameChange}
              required={true}
            />
            <TextField
              size="2"
              parentJSONPointer={parentJSONPointer}
              fieldName="policy_uri"
              label={renderToString("EditOAuthClientForm.policy-uri.label")}
              hint={renderToString(
                "EditOAuthClientForm.policy-uri.description"
              )}
              value={clientConfig.policy_uri ?? ""}
              onChange={onPolicyURIChange}
            />
            <TextField
              size="2"
              parentJSONPointer={parentJSONPointer}
              fieldName="tos_uri"
              label={renderToString("EditOAuthClientForm.tos-uri.label")}
              hint={renderToString("EditOAuthClientForm.tos-uri.description")}
              value={clientConfig.tos_uri ?? ""}
              onChange={onTOSURIChange}
            />
          </SettingsSectionCard>
        ) : null}

        {showCustomUISettings ? (
          <SettingsSectionCard
            className={className}
            contentClassName={CONTENT_CLASSNAME}
            title={
              <FormattedMessage id="EditOAuthClientForm.custom-ui.title" />
            }
          >
            <TextField
              size="2"
              parentJSONPointer={parentJSONPointer}
              fieldName="x_custom_ui_uri"
              label={renderToString("EditOAuthClientForm.custom-ui-uri.label")}
              hint={renderToString(
                "EditOAuthClientForm.custom-ui-uri.description"
              )}
              value={clientConfig.x_custom_ui_uri ?? ""}
              onChange={onCustomUIURI}
            />
          </SettingsSectionCard>
        ) : null}

        {showEndpointsSection ? (
          <SettingsSectionCard
            className={className}
            contentClassName={CONTENT_CLASSNAME}
            title={
              <FormattedMessage id="EditOAuthClientForm.endpoints.title" />
            }
          >
            {endpointsWithLabelIDs.map((e) => {
              return e.endpoint ? (
                <CopyField
                  key={e.labelMessageID}
                  label={renderToString(e.labelMessageID)}
                  value={e.endpoint}
                />
              ) : null;
            })}
          </SettingsSectionCard>
        ) : null}

        {showRefreshTokenSettings ? (
          <SettingsSectionCard
            className={className}
            contentClassName={CONTENT_CLASSNAME}
            title={
              <FormattedMessage id="EditOAuthClientForm.refresh-token.title" />
            }
          >
            <TextField
              size="2"
              parentJSONPointer={parentJSONPointer}
              fieldName="refresh_token_lifetime_seconds"
              label={renderToString("EditOAuthClientForm.refresh-token.label")}
              hint={renderToString(
                "EditOAuthClientForm.refresh-token.description"
              )}
              value={
                clientConfig.refresh_token_lifetime_seconds?.toFixed(0) ?? ""
              }
              onChange={onRefreshTokenLifetimeChange}
            />
            <ToggleWithDescription
              checked={clientConfig.refresh_token_idle_timeout_enabled ?? true}
              onCheckedChange={onChangeRefreshTokenIdleTimeoutEnabled}
              text={renderToString(
                "EditOAuthClientForm.refresh-token-idle-timeout-enabled.label"
              )}
              description={renderToString(
                "EditOAuthClientForm.refresh-token-idle-timeout-enabled.description"
              )}
            />
            <TextField
              size="2"
              parentJSONPointer={parentJSONPointer}
              fieldName="refresh_token_idle_timeout_seconds"
              label={renderToString(
                "EditOAuthClientForm.refresh-token-idle-timeout.label"
              )}
              hint={renderToString(
                "EditOAuthClientForm.refresh-token-idle-timeout.description"
              )}
              value={
                clientConfig.refresh_token_idle_timeout_seconds?.toFixed(0) ??
                ""
              }
              onChange={onIdleTimeoutChange}
              disabled={
                !(clientConfig.refresh_token_idle_timeout_enabled ?? true)
              }
            />
            <HelpText>{refreshTokenHelpText}</HelpText>
            <ToggleWithDescription
              checked={clientConfig.x_max_concurrent_session === 1}
              onCheckedChange={onChangeExpireWhenLoginOnOtherDevice}
              text={renderToString(
                "EditOAuthClientForm.expire-when-login-on-other-device.label"
              )}
              description={renderToString(
                "EditOAuthClientForm.expire-when-login-on-other-device.description"
              )}
            />
            <ToggleWithDescription
              checked={clientConfig.refresh_token_rotation_enabled ?? false}
              onCheckedChange={onChangeRefreshTokenRotationEnabled}
              text={renderToString(
                "EditOAuthClientForm.refresh-token-rotation-enabled.label"
              )}
              description={renderToString(
                "EditOAuthClientForm.refresh-token-rotation-enabled.description"
              )}
            />
          </SettingsSectionCard>
        ) : null}
        {showAccessTokenSettings ? (
          <SettingsSectionCard
            className={className}
            contentClassName={CONTENT_CLASSNAME}
            title={
              <FormattedMessage id="EditOAuthClientForm.access-token.title" />
            }
          >
            <TextField
              size="2"
              parentJSONPointer={parentJSONPointer}
              fieldName="access_token_lifetime_seconds"
              label={renderToString("EditOAuthClientForm.access-token.label")}
              hint={renderToString(
                clientConfig.x_application_type === "m2m"
                  ? "EditOAuthClientForm.access-token.description.m2m"
                  : "EditOAuthClientForm.access-token.description"
              )}
              value={
                clientConfig.access_token_lifetime_seconds?.toFixed(0) ?? ""
              }
              errorRules={[
                makeValidationErrorCustomMessageIDRule(
                  "maximum",
                  /\/access_token_lifetime_seconds$/,
                  "EditOAuthClientForm.access-token.error.maximum"
                ),
              ]}
              onChange={onAccessTokenLifetimeChange}
            />
            <div>
              {alwaysIssueJWTAccessTokenTooltipMessageID != null ? (
                <Tooltip
                  content={
                    <FormattedMessage
                      id={alwaysIssueJWTAccessTokenTooltipMessageID}
                    />
                  }
                >
                  <span className="inline-flex">
                    {issueJWTAccessTokenToggle}
                  </span>
                </Tooltip>
              ) : (
                issueJWTAccessTokenToggle
              )}
            </div>
          </SettingsSectionCard>
        ) : null}
        {showDPoPSettings ? (
          <SettingsSectionCard
            className={className}
            contentClassName={CONTENT_CLASSNAME}
            title={
              <FormattedMessage id="EditOAuthClientForm.sender-constraining.title" />
            }
          >
            <ToggleWithDescription
              checked={!(clientConfig.x_dpop_disabled ?? false)}
              onCheckedChange={onChangeSenderConstraining}
              text={renderToString(
                "EditOAuthClientForm.sender-constraining.require.label"
              )}
              description={
                <FormattedMessage
                  id="EditOAuthClientForm.sender-constraining.description"
                  values={{
                    // eslint-disable-next-line react/no-unstable-nested-components
                    externalLink: (chunks: React.ReactNode) => (
                      <ExternalLink href="https://docs.authgear.com/security/sender-constraining">
                        {chunks}
                      </ExternalLink>
                    ),
                  }}
                />
              }
            />
          </SettingsSectionCard>
        ) : null}
        {showCookieSettings ? (
          <SettingsSectionCard
            className={className}
            contentClassName={CONTENT_CLASSNAME}
            title={
              <FormattedMessage id="EditOAuthClientForm.cookie-settings.title" />
            }
            description={
              <FormattedMessage
                id="EditOAuthClientForm.cookie-settings.description"
                values={{
                  hostname: publicOrigin,
                  // eslint-disable-next-line react/no-unstable-nested-components
                  reactRouterLink: (chunks: React.ReactNode) => (
                    <PortalLink to={`/project/${appID}/advanced/session`}>
                      {chunks}
                    </PortalLink>
                  ),
                }}
              />
            }
          >
            {null}
          </SettingsSectionCard>
        ) : null}
        {showApp2AppSettings ? (
          <SettingsSectionCard
            className={className}
            contentClassName={CONTENT_CLASSNAME}
            title={
              <span id="app2app">
                <FormattedMessage id="EditOAuthClientForm.app2app.title" />
              </span>
            }
          >
            <ToggleWithDescription
              checked={clientConfig.x_app2app_enabled}
              onCheckedChange={onApp2AppEnabledChange}
              text={renderToString("EditOAuthClientForm.app2app.enable.label")}
              description={renderToString(
                "EditOAuthClientForm.app2app.enable.description"
              )}
            />
            <ToggleWithDescription
              checked={
                clientConfig.x_app2app_insecure_device_key_binding_enabled
              }
              onCheckedChange={onApp2AppMigrationChange}
              text={renderToString(
                "EditOAuthClientForm.app2app.migration.label"
              )}
              description={renderToString(
                "EditOAuthClientForm.app2app.migration.description"
              )}
            />
            <HelpText>
              <FormattedMessage
                id="EditOAuthClientForm.app2app.uris.description"
                values={{
                  // eslint-disable-next-line react/no-unstable-nested-components
                  reactRouterLink: (chunks: React.ReactNode) => (
                    <Link to="#uris">{chunks}</Link>
                  ),
                }}
              />
            </HelpText>
          </SettingsSectionCard>
        ) : null}
        <DeleteClientSecretConfirmationDialog
          data={deleteClientSecretDialogData}
          // eslint-disable-next-line @typescript-eslint/strict-void-return
          onConfirm={onConfirmDeleteClientSecret}
          onDismiss={onDismissDeleteClientSecret}
          isLoading={clientSecretHook.isLoading || clientSecretHook.isUpdating}
        />
      </>
    );
  };

export default EditOAuthClientForm;
