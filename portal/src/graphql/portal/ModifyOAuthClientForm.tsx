import React, { useCallback, useContext } from "react";
import cn from "classnames";
import produce from "immer";
import { Checkbox, DirectionalHint, TextField } from "@fluentui/react";
import { Context, FormattedMessage } from "@oursky/react-messageformat";

import Widget from "../../Widget";
import WidgetTitle from "../../WidgetTitle";
import WidgetDescription from "../../WidgetDescription";
import LabelWithTooltip from "../../LabelWithTooltip";
import FormTextField from "../../FormTextField";
import FormTextFieldList from "../../FormTextFieldList";
import { useIntegerTextField, useTextField } from "../../hook/useInput";
import { OAuthClientConfig } from "../../types";
import { ensureNonEmptyString } from "../../util/misc";

import styles from "./ModifyOAuthClientForm.module.scss";

const CHECKBOX_STYLES = {
  label: {
    fontWeight: "600",
  },
};

interface ModifyOAuthClientFormProps {
  className?: string;
  clientConfig: OAuthClientConfig;
  onClientConfigChange: (newClientConfig: OAuthClientConfig) => void;
  isCreation: boolean;
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

function convertIntegerStringToNumber(value: string): number | undefined {
  // Number("") = 0
  const numericValue = Number(value);
  return numericValue === 0 ? undefined : numericValue;
}

const ModifyOAuthClientForm: React.FC<ModifyOAuthClientFormProps> =
  // eslint-disable-next-line complexity
  function ModifyOAuthClientForm(props: ModifyOAuthClientFormProps) {
    const { className, clientConfig, onClientConfigChange, isCreation } = props;

    const { renderToString } = useContext(Context);

    const { onChange: onClientNameChange } = useTextField((value) => {
      onClientConfigChange(
        updateClientConfig(clientConfig, "name", ensureNonEmptyString(value))
      );
    });

    const { onChange: onAccessTokenLifetimeChange } = useIntegerTextField(
      (value) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "access_token_lifetime_seconds",
            convertIntegerStringToNumber(value)
          )
        );
      }
    );
    const { onChange: onRefreshTokenLifetimeChange } = useIntegerTextField(
      (value) => {
        onClientConfigChange(
          updateClientConfig(
            clientConfig,
            "refresh_token_lifetime_seconds",
            convertIntegerStringToNumber(value)
          )
        );
      }
    );
    const { onChange: onIdleTimeoutChange } = useIntegerTextField((value) => {
      onClientConfigChange(
        updateClientConfig(
          clientConfig,
          "refresh_token_idle_timeout_seconds",
          convertIntegerStringToNumber(value)
        )
      );
    });

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

    return (
      <div className={cn(styles.root, className)}>
        <Widget className={styles.widget}>
          {!isCreation && (
            <TextField
              label={renderToString("EditOAuthClientScreen.client-id")}
              value={clientConfig.client_id}
              readOnly={true}
              className={styles.control}
            />
          )}
          <FormTextField
            parentJSONPointer="/oauth/clients/\d+"
            fieldName="name"
            fieldNameMessageID="ModifyOAuthClientForm.name-label"
            className={styles.control}
            value={clientConfig.name ?? ""}
            onChange={onClientNameChange}
            required={true}
          />
          <FormTextFieldList
            className={styles.control}
            parentJSONPointer="/oauth/clients/\d+"
            fieldName="redirect_uris"
            list={clientConfig.redirect_uris}
            onListChange={onRedirectUrisChange}
            addButtonLabelMessageID="ModifyOAuthClientForm.add-uri"
            label={
              <LabelWithTooltip
                labelId="ModifyOAuthClientForm.redirect-uris-label"
                tooltipHeaderId="ModifyOAuthClientForm.redirect-uris-label"
                tooltipMessageId="ModifyOAuthClientForm.redirect-uris-tooltip-message"
                directionalHint={DirectionalHint.bottomLeftEdge}
                required={true}
              />
            }
          />
        </Widget>
        {!isCreation && (
          <Widget className={styles.widget}>
            <WidgetTitle>
              <FormattedMessage id="ModifyOAuthClientForm.token-settings.title" />
            </WidgetTitle>
            <WidgetDescription>
              <FormattedMessage id="ModifyOAuthClientForm.token-settings.description" />
            </WidgetDescription>
            <FormTextField
              parentJSONPointer="/oauth/clients/\d+"
              fieldName="access_token_lifetime_seconds"
              fieldNameMessageID="ModifyOAuthClientForm.acces-token-lifetime-label"
              className={styles.control}
              value={
                clientConfig.access_token_lifetime_seconds?.toString() ?? ""
              }
              onChange={onAccessTokenLifetimeChange}
            />
            <FormTextField
              parentJSONPointer="/oauth/clients/\d+"
              fieldName="refresh_token_lifetime_seconds"
              fieldNameMessageID="ModifyOAuthClientForm.refresh-token-lifetime-label"
              className={styles.control}
              value={
                clientConfig.refresh_token_lifetime_seconds?.toString() ?? ""
              }
              onChange={onRefreshTokenLifetimeChange}
            />
            <Checkbox
              className={styles.control}
              checked={clientConfig.refresh_token_idle_timeout_enabled ?? true}
              onChange={onChangeRefreshTokenIdleTimeoutEnabled}
              label={renderToString(
                "ModifyOAuthClientForm.refresh-token-idle-timeout-enabled.label"
              )}
              styles={CHECKBOX_STYLES}
            />
            <FormTextField
              parentJSONPointer="/oauth/clients/\d+"
              fieldName="refresh_token_idle_timeout_seconds"
              fieldNameMessageID="ModifyOAuthClientForm.refresh-token-idle-timeout-label"
              className={styles.control}
              value={
                clientConfig.refresh_token_idle_timeout_seconds?.toString() ??
                ""
              }
              onChange={onIdleTimeoutChange}
              disabled={
                !(clientConfig.refresh_token_idle_timeout_enabled ?? true)
              }
            />
            <div className={cn(styles.control, styles.checkboxContainer)}>
              <Checkbox
                checked={clientConfig.issue_jwt_access_token}
                onChange={onIssueJWTAccessTokenChange}
              />
              <LabelWithTooltip
                labelId="ModifyOAuthClientForm.issue-jwt-access-token-label"
                tooltipHeaderId=""
                tooltipMessageId="ModifyOAuthClientForm.issue-jwt-access-token-tooltip-message"
                directionalHint={DirectionalHint.bottomLeftEdge}
              />
            </div>
          </Widget>
        )}
        {!isCreation && (
          <Widget className={styles.widget}>
            <WidgetTitle>
              <FormattedMessage id="ModifyOAuthClientForm.cookie-settings.title" />
            </WidgetTitle>
            <WidgetDescription>
              <FormattedMessage id="ModifyOAuthClientForm.cookie-settings.description" />
            </WidgetDescription>
            <FormTextFieldList
              className={styles.control}
              parentJSONPointer="/oauth/clients/\d+"
              fieldName="post_logout_redirect_uris"
              list={clientConfig.post_logout_redirect_uris ?? []}
              onListChange={onPostLogoutRedirectUrisChange}
              addButtonLabelMessageID="ModifyOAuthClientForm.add-uri"
              label={
                <LabelWithTooltip
                  labelId="ModifyOAuthClientForm.post-logout-redirect-uris-label"
                  tooltipHeaderId="ModifyOAuthClientForm.post-logout-redirect-uris-label"
                  tooltipMessageId="ModifyOAuthClientForm.post-logout-redirect-uris-tooltip-message"
                  directionalHint={DirectionalHint.bottomLeftEdge}
                />
              }
            />
          </Widget>
        )}
      </div>
    );
  };

export default ModifyOAuthClientForm;
