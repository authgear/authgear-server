import React, { useCallback, useContext, useMemo, useState } from "react";
import { produce } from "immer";
import { Label, Text, useTheme } from "@fluentui/react";
import { DateTime } from "luxon";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import { useEndpoints } from "../../hook/useEndpoints";

import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";
import FormTextField from "../../FormTextField";
import FormTextFieldList from "../../FormTextFieldList";
import { useTextField } from "../../hook/useInput";
import {
  ApplicationType,
  OAuthClientConfig,
  OAuthClientSecretKey,
} from "../../types";
import { ensureNonEmptyString } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import Toggle from "../../Toggle";
import TextFieldWithCopyButton from "../../TextFieldWithCopyButton";
import { useParams, useNavigate } from "react-router-dom";
import TextField from "../../TextField";
import { Accordion } from "../../components/common/Accordion";
import DefaultButton from "../../DefaultButton";
import ButtonWithLoading from "../../ButtonWithLoading";
import { ClientSecretsHook } from "../../hook/useClientSecrets";
import { useSystemConfig } from "../../context/SystemConfigContext";
import { useStartReauthentication } from "../../graphql/portal/Authenticated";
import {
  DeleteClientSecretConfirmationDialog,
  DeleteClientSecretConfirmationDialogData,
} from "../../components/applications/DeleteClientSecretConfirmationDialog";
import Tooltip from "../../Tooltip";
import { LocationState } from "./EditOAuthClientScreen";
import { makeValidationErrorCustomMessageIDRule } from "../../error/parse";
import { formatSeconds } from "../../util/formatDuration";

const MASKED_SECRET = "***************";

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
    const { themes } = useSystemConfig();
    const theme = useTheme();

    const { appID } = useParams() as { appID: string };

    const { startReauthentication, isRevealing } =
      useStartReauthentication<LocationState>();

    const [deleteClientSecretDialogData, setDeleteClientSecretDialogData] =
      useState<DeleteClientSecretConfirmationDialogData | null>(null);

    const { onChange: onNameChange } = useTextField((value) => {
      onClientConfigChange(
        updateClientConfig(clientConfig, "name", ensureNonEmptyString(value))
      );
    });

    const { onChange: onClientNameChange } = useTextField((value) => {
      onClientConfigChange(
        updateClientConfig(
          clientConfig,
          "client_name",
          ensureNonEmptyString(value)
        )
      );
    });

    const onAccessTokenLifetimeChange = useCallback(
      (_, value?: string) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "access_token_lifetime_seconds",
            parseIntegerAllowLeadingZeros(value)
          )
        );
      },
      [clientConfig, onClientConfigChange]
    );

    const onRefreshTokenLifetimeChange = useCallback(
      (_, value?: string) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "refresh_token_lifetime_seconds",
            parseIntegerAllowLeadingZeros(value)
          )
        );
      },
      [clientConfig, onClientConfigChange]
    );

    const onIdleTimeoutChange = useCallback(
      (_, value?: string) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "refresh_token_idle_timeout_seconds",
            parseIntegerAllowLeadingZeros(value)
          )
        );
      },
      [clientConfig, onClientConfigChange]
    );

    const onRedirectUrisChange = useCallback(
      (list: string[]) => {
        onClientConfigChange(
          updateClientConfig(clientConfig, "redirect_uris", list)
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onPostLogoutRedirectUrisChange = useCallback(
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

    const onChangeRefreshTokenIdleTimeoutEnabled = useCallback(
      (_, value?: boolean) => {
        if (value == null) {
          return;
        }
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "refresh_token_idle_timeout_enabled",
            value
          )
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onChangeExpireWhenLoginOnOtherDevice = useCallback(
      (_, value?: boolean) => {
        if (value == null) {
          return;
        }
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "x_max_concurrent_session",
            value === true ? 1 : undefined
          )
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onChangeRefreshTokenRotationEnabled = useCallback(
      (_, value?: boolean) => {
        if (value == null) {
          return;
        }
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "refresh_token_rotation_enabled",
            value
          )
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onIssueJWTAccessTokenChange = useCallback(
      (_, value?: boolean) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "issue_jwt_access_token",
            value ?? false
          )
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onChangeSenderConstraining = useCallback(
      (_, value?: boolean) => {
        if (value == null) {
          return;
        }
        onClientConfigChange(
          updateClientConfig(clientConfig, "x_dpop_disabled", !value)
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onApp2AppEnabledChange = useCallback(
      (_, value?: boolean) => {
        onClientConfigChange(
          updateClientConfig(clientConfig, "x_app2app_enabled", value ?? false)
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const onApp2AppMigrationChange = useCallback(
      (_, value?: boolean) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "x_app2app_insecure_device_key_binding_enabled",
            value ?? false
          )
        );
      },
      [onClientConfigChange, clientConfig]
    );

    const { onChange: onPolicyURIChange } = useTextField((value) => {
      onClientConfigChange(
        updateClientConfig(
          clientConfig,
          "policy_uri",
          ensureNonEmptyString(value)
        )
      );
    });

    const { onChange: onTOSURIChange } = useTextField((value) => {
      onClientConfigChange(
        updateClientConfig(clientConfig, "tos_uri", ensureNonEmptyString(value))
      );
    });

    const { onChange: onCustomUIURI } = useTextField((value) => {
      onClientConfigChange(
        updateClientConfig(
          clientConfig,
          "x_custom_ui_uri",
          ensureNonEmptyString(value)
        )
      );
    });

    const onGenerateClientSecretClick = useCallback(async () => {
      await clientSecretHook.generate(clientConfig.client_id);
    }, [clientSecretHook, clientConfig.client_id]);

    const onDeleteClientSecretClick = useCallback(
      async (keyItem: OAuthClientSecretKey) => {
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
        !clientConfig.x_application_type ||
        clientConfig.x_application_type === "traditional_webapp" ||
        clientConfig.x_application_type === "confidential" ||
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

    const showEndpoint = useMemo(
      () =>
        !clientConfig.x_application_type ||
        clientConfig.x_application_type === "spa" ||
        clientConfig.x_application_type === "traditional_webapp" ||
        clientConfig.x_application_type === "native",
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

    return (
      <>
        <Widget className={className}>
          <WidgetTitle>
            <FormattedMessage id="EditOAuthClientForm.basic-info.title" />
          </WidgetTitle>
          <FormTextField
            parentJSONPointer={parentJSONPointer}
            fieldName="name"
            label={renderToString("EditOAuthClientForm.name.label")}
            value={clientConfig.name ?? ""}
            onChange={onNameChange}
            required={true}
          />
          <TextFieldWithCopyButton
            label={renderToString("EditOAuthClientForm.client-id.label")}
            value={clientConfig.client_id}
            readOnly={true}
          />
          {showEndpoint ? (
            <TextFieldWithCopyButton
              label={renderToString("EditOAuthClientForm.endpoint.label")}
              value={publicOrigin}
              readOnly={true}
            />
          ) : null}
          <TextField
            label={renderToString("EditOAuthClientForm.application-type.label")}
            value={applicationTypeLabel}
            readOnly={true}
          />
        </Widget>

        {showClientSecret && clientSecrets && clientSecrets.length > 0 ? (
          <>
            <WidgetTitle>
              <FormattedMessage id="EditOAuthClientForm.client-secrets.title" />
            </WidgetTitle>
            {clientSecrets.map((keyItem) => (
              <div key={keyItem.keyID}>
                <TextFieldWithCopyButton
                  label={renderToString(
                    "EditOAuthClientForm.client-secret.label"
                  )}
                  value={keyItem.key ? keyItem.key : MASKED_SECRET}
                  readOnly={true}
                  hideCopyButton={!keyItem.key}
                  additionalIconButtons={
                    clientSecrets.length < 2
                      ? undefined
                      : [
                          {
                            iconProps: { iconName: "Delete" },
                            disabled:
                              clientSecretHook.isLoading ||
                              clientSecretHook.isUpdating,
                            onClick: () => {
                              onDeleteClientSecretClick(keyItem);
                            },
                            theme: themes.destructive,
                          },
                        ]
                  }
                />
                <Text
                  styles={{
                    root: {
                      color: theme.palette.neutralTertiary,
                    },
                  }}
                >
                  {keyItem.createdAt != null ? (
                    <FormattedMessage
                      id="EditOAuthClientForm.client-secret.created-at"
                      values={{
                        datetime: DateTime.fromISO(
                          keyItem.createdAt
                        ).toLocaleString(DateTime.DATETIME_MED_WITH_SECONDS),
                      }}
                    />
                  ) : null}
                </Text>
              </div>
            ))}
            <div className="flex flex-row space-x-4">
              <ButtonWithLoading
                labelId="reveal"
                onClick={onRevealSecretClick}
                disabled={clientSecrets.every((item) => !!item.key)}
                loading={isRevealing}
              />
              {clientSecrets.length < 2 ? (
                <DefaultButton
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
          </>
        ) : null}

        {showURIsSection ? (
          <Widget className={className}>
            {redirectURIsDescription != null ? (
              <>
                <WidgetTitle id="uris">
                  <FormattedMessage id="EditOAuthClientForm.uris.title" />
                </WidgetTitle>
                <FormTextFieldList
                  parentJSONPointer={parentJSONPointer}
                  fieldName="redirect_uris"
                  list={clientConfig.redirect_uris ?? []}
                  onListItemAdd={onRedirectUrisChange}
                  onListItemChange={onRedirectUrisChange}
                  onListItemDelete={onRedirectUrisChange}
                  addButtonLabelMessageID="EditOAuthClientForm.add-uri"
                  label={
                    <Label>
                      <FormattedMessage id="EditOAuthClientForm.redirect-uris.label" />
                    </Label>
                  }
                  description={redirectURIsDescription}
                />
              </>
            ) : null}
            {showPostLogoutRedirectURIsSettings ? (
              <Accordion
                text={
                  <FormattedMessage id="EditOAuthClientForm.more-options" />
                }
              >
                <FormTextFieldList
                  parentJSONPointer={parentJSONPointer}
                  fieldName="post_logout_redirect_uris"
                  list={clientConfig.post_logout_redirect_uris ?? []}
                  onListItemAdd={onPostLogoutRedirectUrisChange}
                  onListItemChange={onPostLogoutRedirectUrisChange}
                  onListItemDelete={onPostLogoutRedirectUrisChange}
                  addButtonLabelMessageID="EditOAuthClientForm.add-uri"
                  label={
                    <Label>
                      <FormattedMessage id="EditOAuthClientForm.post-logout-redirect-uris.label" />
                    </Label>
                  }
                  description={renderToString(
                    clientConfig.x_application_type === "spa"
                      ? "EditOAuthClientForm.post-logout-redirect-uris.spa.description"
                      : "EditOAuthClientForm.post-logout-redirect-uris.description"
                  )}
                />
              </Accordion>
            ) : null}
          </Widget>
        ) : null}

        {showConsentScreenSettings ? (
          <Widget className={className}>
            <WidgetTitle>
              <FormattedMessage id="EditOAuthClientForm.consent-screen.title" />
            </WidgetTitle>
            <FormTextField
              parentJSONPointer={parentJSONPointer}
              fieldName="client_name"
              label={renderToString("EditOAuthClientForm.client-name.label")}
              description={renderToString(
                "EditOAuthClientForm.client-name.description"
              )}
              value={clientConfig.client_name ?? ""}
              onChange={onClientNameChange}
              required={true}
            />
            <FormTextField
              parentJSONPointer={parentJSONPointer}
              fieldName="policy_uri"
              label={renderToString("EditOAuthClientForm.policy-uri.label")}
              description={renderToString(
                "EditOAuthClientForm.policy-uri.description"
              )}
              value={clientConfig.policy_uri ?? ""}
              onChange={onPolicyURIChange}
            />
            <FormTextField
              parentJSONPointer={parentJSONPointer}
              fieldName="tos_uri"
              label={renderToString("EditOAuthClientForm.tos-uri.label")}
              description={renderToString(
                "EditOAuthClientForm.tos-uri.description"
              )}
              value={clientConfig.tos_uri ?? ""}
              onChange={onTOSURIChange}
            />
          </Widget>
        ) : null}

        {showCustomUISettings ? (
          <Widget className={className}>
            <WidgetTitle>
              <FormattedMessage id="EditOAuthClientForm.custom-ui.title" />
            </WidgetTitle>
            <FormTextField
              parentJSONPointer={parentJSONPointer}
              fieldName="x_custom_ui_uri"
              label={renderToString("EditOAuthClientForm.custom-ui-uri.label")}
              description={renderToString(
                "EditOAuthClientForm.custom-ui-uri.description"
              )}
              value={clientConfig.x_custom_ui_uri ?? ""}
              onChange={onCustomUIURI}
            />
          </Widget>
        ) : null}

        {showEndpointsSection ? (
          <Widget className={className}>
            <WidgetTitle>
              <FormattedMessage id="EditOAuthClientForm.endpoints.title" />
            </WidgetTitle>
            {endpointsWithLabelIDs.map((e) => {
              return e.endpoint ? (
                <TextFieldWithCopyButton
                  key={e.labelMessageID}
                  label={renderToString(e.labelMessageID)}
                  value={e.endpoint}
                  readOnly={true}
                />
              ) : null;
            })}
          </Widget>
        ) : null}

        {showRefreshTokenSettings ? (
          <Widget className={className}>
            <WidgetTitle>
              <FormattedMessage id="EditOAuthClientForm.refresh-token.title" />
            </WidgetTitle>
            <FormTextField
              parentJSONPointer={parentJSONPointer}
              fieldName="refresh_token_lifetime_seconds"
              label={renderToString("EditOAuthClientForm.refresh-token.label")}
              description={renderToString(
                "EditOAuthClientForm.refresh-token.description"
              )}
              value={
                clientConfig.refresh_token_lifetime_seconds?.toFixed(0) ?? ""
              }
              onChange={onRefreshTokenLifetimeChange}
            />
            <Toggle
              checked={clientConfig.refresh_token_idle_timeout_enabled ?? true}
              onChange={onChangeRefreshTokenIdleTimeoutEnabled}
              label={renderToString(
                "EditOAuthClientForm.refresh-token-idle-timeout-enabled.label"
              )}
              description={renderToString(
                "EditOAuthClientForm.refresh-token-idle-timeout-enabled.description"
              )}
            />
            <FormTextField
              parentJSONPointer={parentJSONPointer}
              fieldName="refresh_token_idle_timeout_seconds"
              label={renderToString(
                "EditOAuthClientForm.refresh-token-idle-timeout.label"
              )}
              description={renderToString(
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
            <Toggle
              checked={clientConfig.x_max_concurrent_session === 1}
              onChange={onChangeExpireWhenLoginOnOtherDevice}
              label={renderToString(
                "EditOAuthClientForm.expire-when-login-on-other-device.label"
              )}
              description={renderToString(
                "EditOAuthClientForm.expire-when-login-on-other-device.description"
              )}
            />
            <Toggle
              checked={clientConfig.refresh_token_rotation_enabled ?? false}
              onChange={onChangeRefreshTokenRotationEnabled}
              label={renderToString(
                "EditOAuthClientForm.refresh-token-rotation-enabled.label"
              )}
              description={renderToString(
                "EditOAuthClientForm.refresh-token-rotation-enabled.description"
              )}
            />
          </Widget>
        ) : null}
        {showAccessTokenSettings ? (
          <Widget className={className}>
            <WidgetTitle>
              <FormattedMessage id="EditOAuthClientForm.access-token.title" />
            </WidgetTitle>
            <FormTextField
              parentJSONPointer={parentJSONPointer}
              fieldName="access_token_lifetime_seconds"
              label={renderToString("EditOAuthClientForm.access-token.label")}
              description={renderToString(
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
              <Label
                htmlFor="issue-jwt-access-token-toggle"
                disabled={isIssueJWTAccessTokenToggleDisabled}
              >
                {renderToString(
                  "EditOAuthClientForm.issue-jwt-access-token.label"
                )}
              </Label>
              <Tooltip
                tooltipMessageId={
                  alwaysIssueJWTAccessTokenTooltipMessageID ?? ""
                }
                isHidden={alwaysIssueJWTAccessTokenTooltipMessageID == null}
              >
                <Toggle
                  id="issue-jwt-access-token-toggle"
                  checked={clientConfig.issue_jwt_access_token}
                  disabled={isIssueJWTAccessTokenToggleDisabled}
                  onChange={onIssueJWTAccessTokenChange}
                />
              </Tooltip>
            </div>
          </Widget>
        ) : null}
        <Widget className={className}>
          <WidgetTitle>
            <FormattedMessage id="EditOAuthClientForm.sender-constraining.title" />
          </WidgetTitle>
          <Toggle
            checked={!(clientConfig.x_dpop_disabled ?? false)}
            onChange={onChangeSenderConstraining}
            label={renderToString(
              "EditOAuthClientForm.sender-constraining.require.label"
            )}
            description={
              <FormattedMessage id="EditOAuthClientForm.sender-constraining.description" />
            }
          />
        </Widget>
        {showCookieSettings ? (
          <Widget className={className}>
            <WidgetTitle>
              <FormattedMessage id="EditOAuthClientForm.cookie-settings.title" />
            </WidgetTitle>
            <WidgetDescription>
              <FormattedMessage
                id="EditOAuthClientForm.cookie-settings.description"
                values={{
                  to: `/project/${appID}/advanced/session`,
                  hostname: publicOrigin,
                }}
              />
            </WidgetDescription>
          </Widget>
        ) : null}
        {showApp2AppSettings ? (
          <Widget className={className}>
            <WidgetTitle id="app2app">
              <FormattedMessage id="EditOAuthClientForm.app2app.title" />
            </WidgetTitle>
            <Toggle
              checked={clientConfig.x_app2app_enabled}
              onChange={onApp2AppEnabledChange}
              label={renderToString("EditOAuthClientForm.app2app.enable.label")}
              description={renderToString(
                "EditOAuthClientForm.app2app.enable.description"
              )}
            />
            <Toggle
              checked={
                clientConfig.x_app2app_insecure_device_key_binding_enabled
              }
              onChange={onApp2AppMigrationChange}
              label={renderToString(
                "EditOAuthClientForm.app2app.migration.label"
              )}
              description={renderToString(
                "EditOAuthClientForm.app2app.migration.description"
              )}
            />
            <HelpText>
              <FormattedMessage id="EditOAuthClientForm.app2app.uris.description" />
            </HelpText>
          </Widget>
        ) : null}
        <DeleteClientSecretConfirmationDialog
          data={deleteClientSecretDialogData}
          onConfirm={onConfirmDeleteClientSecret}
          onDismiss={onDismissDeleteClientSecret}
          isLoading={clientSecretHook.isLoading || clientSecretHook.isUpdating}
        />
      </>
    );
  };

export default EditOAuthClientForm;

function HelpText(props: { children: React.ReactNode }) {
  const { children } = props;
  const theme = useTheme();
  return (
    <Text
      block={true}
      styles={{
        root: {
          background: theme.palette.neutralLighter,
          lineHeight: "20px",
          padding: "8px 12px",
        },
      }}
    >
      {children}
    </Text>
  );
}
