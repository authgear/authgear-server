import React, { useCallback, useContext, useMemo } from "react";
import produce from "immer";
import { Label, Text, useTheme } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";
import FormTextField from "../../FormTextField";
import FormTextFieldList from "../../FormTextFieldList";
import { useTextField } from "../../hook/useInput";
import { ApplicationType, OAuthClientConfig } from "../../types";
import { ensureNonEmptyString } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import Toggle from "../../Toggle";
import TextFieldWithCopyButton from "../../TextFieldWithCopyButton";
import { useParams } from "react-router-dom";
import TextField from "../../TextField";
import TextFieldWithButton from "../../TextFieldWithButton";
import { startReauthentication } from "./Authenticated";

const MASKED_SECRET = "***************";

interface EditOAuthClientFormProps {
  publicOrigin: string;
  className?: string;
  clientConfig: OAuthClientConfig;
  clientSecret?: string;
  onClientConfigChange: (newClientConfig: OAuthClientConfig) => void;
}

export function getApplicationTypeMessageID(key?: string): string {
  const messageIDMap: Record<string, string> = {
    spa: "oauth-client.application-type.spa",
    traditional_webapp: "oauth-client.application-type.traditional-webapp",
    native: "oauth-client.application-type.native",
    third_party_app: "oauth-client.application-type.third-party-app",
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
  // eslint-disable-next-line complexity
  function EditOAuthClientForm(props: EditOAuthClientFormProps) {
    const {
      className,
      clientConfig,
      clientSecret,
      publicOrigin,
      onClientConfigChange,
    } = props;

    const { renderToString } = useContext(Context);
    const theme = useTheme();

    const { appID } = useParams() as { appID: string };

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

    const onClickReveal = useCallback(() => {
      startReauthentication().catch((e) => {
        // Normally there should not be any error.
        console.error(e);
      });
    }, []);

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

    const applicationTypeLabel = useMemo(() => {
      const messageID = getApplicationTypeMessageID(
        clientConfig.x_application_type
      );
      return renderToString(messageID);
    }, [clientConfig.x_application_type, renderToString]);

    const redirectURIsDescription = useMemo(() => {
      const messageIdMap: Record<ApplicationType, string> = {
        spa: "EditOAuthClientForm.redirect-uris.description.spa",
        traditional_webapp:
          "EditOAuthClientForm.redirect-uris.description.traditional-webapp",
        native: "EditOAuthClientForm.redirect-uris.description.native",
        third_party_app:
          "EditOAuthClientForm.redirect-uris.description.third-party-app",
      };
      const messageID = clientConfig.x_application_type
        ? messageIdMap[clientConfig.x_application_type]
        : "EditOAuthClientForm.redirect-uris.description.unspecified";
      return renderToString(messageID);
    }, [clientConfig.x_application_type, renderToString]);

    const showPostLogoutRedirectURIsSettings = useMemo(
      () =>
        !clientConfig.x_application_type ||
        clientConfig.x_application_type === "spa" ||
        clientConfig.x_application_type === "traditional_webapp",
      [clientConfig.x_application_type]
    );

    const showCookieSettings = useMemo(
      () =>
        !clientConfig.x_application_type ||
        clientConfig.x_application_type === "traditional_webapp" ||
        clientConfig.x_application_type === "third_party_app",
      [clientConfig.x_application_type]
    );

    const showTokenSettings = useMemo(
      () =>
        !clientConfig.x_application_type ||
        clientConfig.x_application_type === "spa" ||
        clientConfig.x_application_type === "native" ||
        clientConfig.x_application_type === "third_party_app",
      [clientConfig.x_application_type]
    );

    const showConsentScreenSettings = useMemo(
      () => clientConfig.x_application_type === "third_party_app",
      [clientConfig.x_application_type]
    );

    // fixme(custom-ui): allow setting custom ui bases on features.yaml
    const showCustomUISettings = true;

    const showClientSecret = useMemo(
      () => clientConfig.x_application_type === "third_party_app",
      [clientConfig.x_application_type]
    );

    const showEndpoint = useMemo(
      () => clientConfig.x_application_type !== "third_party_app",
      [clientConfig.x_application_type]
    );

    const refreshTokenHelpText = useMemo(() => {
      if (clientConfig.refresh_token_idle_timeout_enabled) {
        return renderToString(
          "EditOAuthClientForm.refresh-token.help-text.idle-timeout-enabled",
          {
            refreshTokenLifetime:
              clientConfig.refresh_token_lifetime_seconds?.toFixed(0) ?? "",
            refreshTokenIdleTimeout:
              clientConfig.refresh_token_idle_timeout_seconds?.toFixed(0) ?? "",
          }
        );
      }
      return renderToString(
        "EditOAuthClientForm.refresh-token.help-text.idle-timeout-disabled",
        {
          refreshTokenLifetime:
            clientConfig.refresh_token_lifetime_seconds?.toFixed(0) ?? "",
        }
      );
    }, [
      clientConfig.refresh_token_lifetime_seconds,
      clientConfig.refresh_token_idle_timeout_enabled,
      clientConfig.refresh_token_idle_timeout_seconds,
      renderToString,
    ]);

    const showEndpointsSection = useMemo(
      () => clientConfig.x_application_type === "third_party_app",
      [clientConfig.x_application_type]
    );

    const endpoints = useMemo(() => {
      const list: {
        labelMessageID: string;
        endpoint: string;
      }[] = showEndpointsSection
        ? [
            {
              labelMessageID:
                "EditOAuthClientForm.openid-configuration-endpoint.label",
              endpoint: `${publicOrigin}/.well-known/openid-configuration`,
            },
            {
              labelMessageID:
                "EditOAuthClientForm.authorization-endpoint.label",
              endpoint: `${publicOrigin}/oauth2/authorize`,
            },
            {
              labelMessageID: "EditOAuthClientForm.token-endpoint.label",
              endpoint: `${publicOrigin}/oauth2/token`,
            },
            {
              labelMessageID: "EditOAuthClientForm.userinfo-endpoint.label",
              endpoint: `${publicOrigin}/oauth2/userinfo`,
            },
            {
              labelMessageID: "EditOAuthClientForm.jwks-uri.label",
              endpoint: `${publicOrigin}/oauth2/jwks`,
            },
          ]
        : [];
      return list;
    }, [showEndpointsSection, publicOrigin]);

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
          {showClientSecret ? (
            clientSecret ? (
              <TextFieldWithCopyButton
                label={renderToString(
                  "EditOAuthClientForm.client-secret.label"
                )}
                value={clientSecret}
                readOnly={true}
              />
            ) : (
              <TextFieldWithButton
                label={renderToString(
                  "EditOAuthClientForm.client-secret.label"
                )}
                value={MASKED_SECRET}
                readOnly={true}
                buttonText={renderToString("reveal")}
                onButtonClick={onClickReveal}
              />
            )
          ) : null}
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

        <Widget className={className}>
          <WidgetTitle>
            <FormattedMessage id="EditOAuthClientForm.uris.title" />
          </WidgetTitle>
          <FormTextFieldList
            parentJSONPointer={parentJSONPointer}
            fieldName="redirect_uris"
            list={clientConfig.redirect_uris}
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
          {showPostLogoutRedirectURIsSettings ? (
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
          ) : null}
        </Widget>

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
            {endpoints.map((item, i) => (
              <TextFieldWithCopyButton
                key={i}
                label={renderToString(item.labelMessageID)}
                value={item.endpoint}
                readOnly={true}
              />
            ))}
          </Widget>
        ) : null}

        {showTokenSettings ? (
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
              {refreshTokenHelpText}
            </Text>
          </Widget>
        ) : null}
        {showTokenSettings ? (
          <Widget className={className}>
            <WidgetTitle>
              <FormattedMessage id="EditOAuthClientForm.access-token.title" />
            </WidgetTitle>
            <FormTextField
              parentJSONPointer={parentJSONPointer}
              fieldName="access_token_lifetime_seconds"
              label={renderToString("EditOAuthClientForm.access-token.label")}
              description={renderToString(
                "EditOAuthClientForm.access-token.description"
              )}
              value={
                clientConfig.access_token_lifetime_seconds?.toFixed(0) ?? ""
              }
              onChange={onAccessTokenLifetimeChange}
            />
            <Toggle
              checked={clientConfig.issue_jwt_access_token}
              onChange={onIssueJWTAccessTokenChange}
              label={renderToString(
                "EditOAuthClientForm.issue-jwt-access-token.label"
              )}
            />
          </Widget>
        ) : null}
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
      </>
    );
  };

export default EditOAuthClientForm;
