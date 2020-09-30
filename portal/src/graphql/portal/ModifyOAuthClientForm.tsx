import React, { useContext, useEffect } from "react";
import cn from "classnames";
import { DirectionalHint, TagPicker, TextField } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";

import LabelWithTooltip from "../../LabelWithTooltip";
import { useTagPickerWithNewTags, useTextField } from "../../hook/useInput";
import { OAuthClientConfig } from "../../types";

import styles from "./ModifyOAuthClientForm.module.scss";

interface ModifyOAuthClientFormProps {
  className?: string;
  clientConfig: OAuthClientConfig;
  onClientConfigChange: (newClientConfig: OAuthClientConfig) => void;
}

function constructClientConfigState(
  clientId: string,
  accessTokenLifetime: string,
  refreshTokenLifetime: string,
  redirectUris: string[],
  postLogoutRedirectUris: string[]
): OAuthClientConfig {
  // Number("") = 0
  const accessTokenLifetimeSec = Number(accessTokenLifetime);
  const refreshTokenLifetimeSec = Number(refreshTokenLifetime);

  return {
    client_id: clientId,
    redirect_uris: redirectUris,
    post_logout_redirect_uris:
      postLogoutRedirectUris.length > 0 ? postLogoutRedirectUris : undefined,
    grant_types: ["authorization_code", "refresh_token"],
    response_types: ["code", "none"],
    access_token_lifetime_seconds:
      accessTokenLifetimeSec > 0 ? accessTokenLifetimeSec : undefined,
    refresh_token_lifetime_seconds:
      refreshTokenLifetimeSec > 0 ? refreshTokenLifetimeSec : undefined,
  };
}

const ModifyOAuthClientForm: React.FC<ModifyOAuthClientFormProps> = function ModifyOAuthClientForm(
  props: ModifyOAuthClientFormProps
) {
  const { className, clientConfig, onClientConfigChange } = props;

  const { renderToString } = useContext(Context);

  const { value: clientName, onChange: onClientNameChange } = useTextField("");

  const {
    value: accessTokenLifetime,
    onChange: onAccessTokenLifetimeChange,
  } = useTextField(
    clientConfig.access_token_lifetime_seconds?.toString() ?? "",
    "integer"
  );
  const {
    value: refreshTokenLifetime,
    onChange: onRefreshTokenLifetimeChange,
  } = useTextField(
    clientConfig.refresh_token_lifetime_seconds?.toString() ?? "",
    "integer"
  );

  const {
    list: redirectUris,
    onChange: onRedirectUrisChange,
    onResolveSuggestions: onResolveRedirectUrisSuggestions,
    defaultSelectedItems: defaultSelectedRedirectUris,
  } = useTagPickerWithNewTags(clientConfig.redirect_uris);

  const {
    list: postLogoutRedirectUris,
    onChange: onPostLogoutRedirectUrisChange,
    onResolveSuggestions: onResolvePostLogoutUrisSuggestions,
    defaultSelectedItems: defaultSelectedPostLogoutUris,
  } = useTagPickerWithNewTags(clientConfig.post_logout_redirect_uris ?? []);

  useEffect(() => {
    onClientConfigChange(
      constructClientConfigState(
        clientConfig.client_id,
        accessTokenLifetime,
        refreshTokenLifetime,
        redirectUris,
        postLogoutRedirectUris
      )
    );
  }, [
    clientConfig.client_id,
    onClientConfigChange,

    accessTokenLifetime,
    refreshTokenLifetime,
    redirectUris,
    postLogoutRedirectUris,
  ]);

  return (
    <section className={cn(styles.root, className)}>
      <TextField
        className={styles.inputField}
        label={renderToString("ModifyOAuthClientForm.name-label")}
        value={clientName}
        onChange={onClientNameChange}
      />
      <TextField
        className={styles.inputField}
        label={renderToString(
          "ModifyOAuthClientForm.acces-token-lifetime-label"
        )}
        value={accessTokenLifetime}
        onChange={onAccessTokenLifetimeChange}
      />
      <TextField
        className={styles.inputField}
        label={renderToString(
          "ModifyOAuthClientForm.refresh-token-lifetime-label"
        )}
        value={refreshTokenLifetime}
        onChange={onRefreshTokenLifetimeChange}
      />
      <LabelWithTooltip
        labelId="ModifyOAuthClientForm.redirect-uris-label"
        tooltipHeaderId="ModifyOAuthClientForm.redirect-uris-label"
        tooltipMessageId="ModifyOAuthClientForm.redirect-uris-tooltip-message"
        directionalHint={DirectionalHint.bottomLeftEdge}
      />
      <TagPicker
        className={styles.inputField}
        onChange={onRedirectUrisChange}
        onResolveSuggestions={onResolveRedirectUrisSuggestions}
        defaultSelectedItems={defaultSelectedRedirectUris}
      />
      <LabelWithTooltip
        labelId="ModifyOAuthClientForm.post-logout-redirect-uris-label"
        tooltipHeaderId="ModifyOAuthClientForm.post-logout-redirect-uris-label"
        tooltipMessageId="ModifyOAuthClientForm.post-logout-redirect-uris-tooltip-message"
        directionalHint={DirectionalHint.bottomLeftEdge}
      />
      <TagPicker
        className={styles.inputField}
        onChange={onPostLogoutRedirectUrisChange}
        onResolveSuggestions={onResolvePostLogoutUrisSuggestions}
        defaultSelectedItems={defaultSelectedPostLogoutUris}
      />
    </section>
  );
};

export default ModifyOAuthClientForm;
