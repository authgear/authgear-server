import React, { useCallback } from "react";
import cn from "classnames";
import produce from "immer";
import { DirectionalHint } from "@fluentui/react";

import LabelWithTooltip from "../../LabelWithTooltip";
import FormTextField from "../../FormTextField";
import FormTextFieldList from "../../FormTextFieldList";
import { useIntegerTextField, useTextField } from "../../hook/useInput";
import { OAuthClientConfig } from "../../types";
import { ensureNonEmptyString } from "../../util/misc";

import styles from "./ModifyOAuthClientForm.module.scss";

interface ModifyOAuthClientFormProps {
  className?: string;
  clientConfig: OAuthClientConfig;
  onClientConfigChange: (newClientConfig: OAuthClientConfig) => void;
}

const jsonPointerRegExp: Pick<
  Record<
    keyof OAuthClientConfig,
    {
      field: RegExp;
      parent: RegExp;
      getItemJSONPointer?: (index: number) => RegExp;
    }
  >,
  | "name"
  | "access_token_lifetime_seconds"
  | "refresh_token_lifetime_seconds"
  | "redirect_uris"
  | "post_logout_redirect_uris"
> = {
  name: {
    field: /^\/oauth\/clients\/[0-9]+\/name$/,
    parent: /^\/oauth\/clients\/[0-9]+$/,
  },
  access_token_lifetime_seconds: {
    field: /^\/oauth\/clients\/[0-9]+\/access_token_lifetime_seconds$/,
    parent: /^\/oauth\/clients\/[0-9]+$/,
  },
  refresh_token_lifetime_seconds: {
    field: /^\/oauth\/clients\/[0-9]+\/refresh_token_lifetime_seconds$/,
    parent: /^\/oauth\/clients\/[0-9]+$/,
  },
  redirect_uris: {
    field: /^\/oauth\/clients\/[0-9]+\/redirect_uris$/,
    parent: /^\/oauth\/clients\/[0-9]+$/,
    getItemJSONPointer: (index) =>
      new RegExp(`^/oauth/clients/[0-9]+/redirect_uris/${index}$`),
  },
  post_logout_redirect_uris: {
    field: /^\/oauth\/clients\/[0-9]+\/post_logout_redirect_uris$/,
    parent: /^\/oauth\/clients\/[0-9]+$/,
    getItemJSONPointer: (index) =>
      new RegExp(`^/oauth/clients/[0-9]+/post_logout_redirect_uris/${index}$`),
  },
};

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

const ModifyOAuthClientForm: React.FC<ModifyOAuthClientFormProps> = function ModifyOAuthClientForm(
  props: ModifyOAuthClientFormProps
) {
  const { className, clientConfig, onClientConfigChange } = props;

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

  return (
    <section className={cn(styles.root, className)}>
      <FormTextField
        jsonPointer={jsonPointerRegExp.name.field}
        parentJSONPointer={jsonPointerRegExp.name.parent}
        fieldName="name"
        fieldNameMessageID="ModifyOAuthClientForm.name-label"
        className={styles.inputField}
        value={clientConfig.name ?? ""}
        onChange={onClientNameChange}
        required={true}
      />
      <FormTextField
        jsonPointer={jsonPointerRegExp.access_token_lifetime_seconds.field}
        parentJSONPointer={
          jsonPointerRegExp.access_token_lifetime_seconds.parent
        }
        fieldName="access_token_lifetime_seconds"
        fieldNameMessageID="ModifyOAuthClientForm.acces-token-lifetime-label"
        className={styles.inputField}
        value={clientConfig.access_token_lifetime_seconds?.toString() ?? ""}
        onChange={onAccessTokenLifetimeChange}
      />
      <FormTextField
        jsonPointer={jsonPointerRegExp.refresh_token_lifetime_seconds.field}
        parentJSONPointer={
          jsonPointerRegExp.refresh_token_lifetime_seconds.parent
        }
        fieldName="refresh_token_lifetime_seconds"
        fieldNameMessageID="ModifyOAuthClientForm.refresh-token-lifetime-label"
        className={styles.inputField}
        value={clientConfig.refresh_token_lifetime_seconds?.toString() ?? ""}
        onChange={onRefreshTokenLifetimeChange}
      />
      <FormTextFieldList
        className={styles.inputFieldList}
        jsonPointer={jsonPointerRegExp.redirect_uris.field}
        parentJSONPointer={jsonPointerRegExp.redirect_uris.parent}
        getItemJSONPointer={jsonPointerRegExp.redirect_uris.getItemJSONPointer!}
        fieldName="redirect_uris"
        fieldNameMessageID="ModifyOAuthClientForm.redirect-uris-label"
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
      <FormTextFieldList
        className={styles.inputFieldList}
        jsonPointer={jsonPointerRegExp.post_logout_redirect_uris.field}
        parentJSONPointer={jsonPointerRegExp.post_logout_redirect_uris.parent}
        getItemJSONPointer={
          jsonPointerRegExp.post_logout_redirect_uris.getItemJSONPointer!
        }
        fieldName="post_logout_redirect_uris"
        fieldNameMessageID="ModifyOAuthClientForm.post-logout-redirect-uris-label"
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
    </section>
  );
};

export default ModifyOAuthClientForm;
