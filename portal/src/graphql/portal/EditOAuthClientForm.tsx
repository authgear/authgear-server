import React, { useCallback, useContext, useMemo } from "react";
import produce from "immer";
import { Dropdown, Label, Text, useTheme } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";
import FormTextField from "../../FormTextField";
import FormTextFieldList from "../../FormTextFieldList";
import { useDropdown, useTextField } from "../../hook/useInput";
import {
  ApplicationType,
  applicationTypes,
  OAuthClientConfig,
} from "../../types";
import { ensureNonEmptyString } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import Toggle from "../../Toggle";
import TextFieldWithCopyButton from "../../TextFieldWithCopyButton";
import { useParams } from "react-router-dom";

const ALL_APPLICATION_TYPES: ApplicationType[] = [...applicationTypes];
interface EditOAuthClientFormProps {
  publicOrigin: string;
  className?: string;
  clientConfig: OAuthClientConfig;
  onClientConfigChange: (newClientConfig: OAuthClientConfig) => void;
}

export function getApplicationTypeMessageID(key?: string): string {
  const messageIDMap: Record<string, string> = {
    spa: "oauth-client.application-type.spa",
    traditional_webapp: "oauth-client.application-type.traditional-webapp",
    native: "oauth-client.application-type.native",
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

const EditOAuthClientForm: React.VFC<EditOAuthClientFormProps> =
  // eslint-disable-next-line complexity
  function EditOAuthClientForm(props: EditOAuthClientFormProps) {
    const { className, clientConfig, publicOrigin, onClientConfigChange } =
      props;

    const { renderToString } = useContext(Context);
    const theme = useTheme();

    const { appID } = useParams() as { appID: string };

    const { onChange: onClientNameChange } = useTextField((value) => {
      onClientConfigChange(
        updateClientConfig(clientConfig, "name", ensureNonEmptyString(value))
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

    const renderApplicationType = useCallback(
      (key: ApplicationType) => {
        const messageID = getApplicationTypeMessageID(key);
        return renderToString(messageID);
      },
      [renderToString]
    );

    const {
      options: applicationTypeOptions,
      onChange: onApplicationTypeChange,
    } = useDropdown(
      ALL_APPLICATION_TYPES,
      (option) => {
        onClientConfigChange(
          updateClientConfig(clientConfig, "x_application_type", option)
        );
      },
      clientConfig.x_application_type,
      renderApplicationType
    );

    const redirectURIsDescription = useMemo(() => {
      const messageIdMap: Record<ApplicationType, string> = {
        spa: "EditOAuthClientForm.redirect-uris.description.spa",
        traditional_webapp:
          "EditOAuthClientForm.redirect-uris.description.traditional-webapp",
        native: "EditOAuthClientForm.redirect-uris.description.native",
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
        clientConfig.x_application_type === "traditional_webapp",
      [clientConfig.x_application_type]
    );

    const showTokenSettings = useMemo(
      () =>
        !clientConfig.x_application_type ||
        clientConfig.x_application_type === "spa" ||
        clientConfig.x_application_type === "native",
      [clientConfig.x_application_type]
    );

    const parentJSONPointer = /\/oauth\/clients\/\d+/;

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
            onChange={onClientNameChange}
            required={true}
          />
          <TextFieldWithCopyButton
            label={renderToString("EditOAuthClientForm.client-id.label")}
            value={clientConfig.client_id}
            readOnly={true}
          />
          <TextFieldWithCopyButton
            label={renderToString("EditOAuthClientForm.endpoint.label")}
            value={publicOrigin}
            readOnly={true}
          />
          <Dropdown
            placeholder={renderToString(
              "oauth-client.application-type.unspecified"
            )}
            label={renderToString("EditOAuthClientForm.application-type.label")}
            options={applicationTypeOptions}
            selectedKey={clientConfig.x_application_type}
            onChange={onApplicationTypeChange}
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
            onListChange={onRedirectUrisChange}
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
              onListChange={onPostLogoutRedirectUrisChange}
              addButtonLabelMessageID="EditOAuthClientForm.add-uri"
              label={
                <Label>
                  <FormattedMessage id="EditOAuthClientForm.post-logout-redirect-uris.label" />
                </Label>
              }
              description={renderToString(
                "EditOAuthClientForm.post-logout-redirect-uris.description"
              )}
            />
          ) : null}
        </Widget>
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
