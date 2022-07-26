import React, { useCallback, useContext } from "react";
import produce from "immer";
import { Label } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";
import FormTextField from "../../FormTextField";
import FormTextFieldList from "../../FormTextFieldList";
import { useTextField } from "../../hook/useInput";
import { OAuthClientConfig } from "../../types";
import { ensureNonEmptyString } from "../../util/misc";
import { parseIntegerAllowLeadingZeros } from "../../util/input";
import Toggle from "../../Toggle";
import TextFieldWithCopyButton from "../../TextFieldWithCopyButton";

interface EditOAuthClientFormProps {
  publicOrigin: string;
  className?: string;
  clientConfig: OAuthClientConfig;
  onClientConfigChange: (newClientConfig: OAuthClientConfig) => void;
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

function updateClientConfig<K extends keyof OAuthClientConfig>(
  clientConfig: OAuthClientConfig,
  field: K,
  newValue: OAuthClientConfig[K]
): OAuthClientConfig {
  return produce(clientConfig, (draftConfig) => {
    draftConfig[field] = newValue;
  });
}

const EditOAuthClientForm: React.FC<EditOAuthClientFormProps> =
  // eslint-disable-next-line complexity
  function EditOAuthClientForm(props: EditOAuthClientFormProps) {
    const { className, clientConfig, publicOrigin, onClientConfigChange } =
      props;

    const { renderToString } = useContext(Context);

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

    const parentJSONPointer = /\/oauth\/clients\/\d+/;

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
            description={renderToString(
              "EditOAuthClientForm.redirect-uris.description"
            )}
          />
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
        </Widget>
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
              clientConfig.refresh_token_idle_timeout_seconds?.toFixed(0) ?? ""
            }
            onChange={onIdleTimeoutChange}
            disabled={
              !(clientConfig.refresh_token_idle_timeout_enabled ?? true)
            }
          />
        </Widget>
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
            value={clientConfig.access_token_lifetime_seconds?.toFixed(0) ?? ""}
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
        <Widget className={className}>
          <WidgetTitle>
            <FormattedMessage id="EditOAuthClientForm.cookie-settings.title" />
          </WidgetTitle>
          <WidgetDescription>
            <FormattedMessage
              id="EditOAuthClientForm.cookie-settings.description"
              values={{
                to: "./../../advanced/session",
                hostname: publicOrigin,
              }}
            />
          </WidgetDescription>
        </Widget>
      </>
    );
  };

export default EditOAuthClientForm;
