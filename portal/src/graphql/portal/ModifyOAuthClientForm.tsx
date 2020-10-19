import React, { useContext, useMemo } from "react";
import cn from "classnames";
import produce from "immer";
import { DirectionalHint, TagPicker, Text, TextField } from "@fluentui/react";
import { Context } from "@oursky/react-messageformat";

import LabelWithTooltip from "../../LabelWithTooltip";
import {
  useIntegerTextField,
  useTagPickerWithNewTags,
  useTextField,
} from "../../hook/useInput";
import { OAuthClientConfig } from "../../types";
import { parseError } from "../../util/error";
import { defaultFormatErrorMessageList } from "../../util/validation";
import { ensureNonEmptyString } from "../../util/misc";

import styles from "./ModifyOAuthClientForm.module.scss";

interface ModifyOAuthClientFormProps {
  className?: string;
  clientConfig: OAuthClientConfig;
  onClientConfigChange: (newClientConfig: OAuthClientConfig) => void;
  updateAppConfigError: unknown;
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
  const {
    className,
    clientConfig,
    onClientConfigChange,
    updateAppConfigError,
  } = props;

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

  const {
    selectedItems: redirectUriItems,
    onChange: onRedirectUrisChange,
    onResolveSuggestions: onResolveRedirectUrisSuggestions,
  } = useTagPickerWithNewTags(clientConfig.redirect_uris, (list) => {
    onClientConfigChange(
      updateClientConfig(clientConfig, "redirect_uris", list)
    );
  });

  const {
    selectedItems: postLogoutRedirectUriItems,
    onChange: onPostLogoutRedirectUrisChange,
    onResolveSuggestions: onResolvePostLogoutUrisSuggestions,
  } = useTagPickerWithNewTags(
    clientConfig.post_logout_redirect_uris ?? [],
    (list) => {
      onClientConfigChange(
        updateClientConfig(
          clientConfig,
          "post_logout_redirect_uris",
          list.length > 0 ? list : undefined
        )
      );
    }
  );

  const errorMessageMap = useMemo(() => {
    const clientNameErrorMessages: string[] = [];
    const redirectUrisErrorMessages: string[] = [];
    const postLogoutRedirectUrisErrorMessages: string[] = [];
    const violations = parseError(updateAppConfigError);
    const redirectUrisRegexp = /^\/oauth\/clients\/[^/]*\/redirect_uris(\/[0-9]*)?$/;
    const postLogoutRedirectUrisRegexp = /^\/oauth\/clients\/[^/]*\/post_logout_redirect_uris(\/[0-9]*)?$/;
    for (const violation of violations) {
      switch (violation.kind) {
        case "minItems":
          if (redirectUrisRegexp.test(violation.location)) {
            redirectUrisErrorMessages.push(
              renderToString(
                "ModifyOAuthClientForm.redirect-uris.min-items-error",
                { minItems: violation.minItems }
              )
            );
          }
          break;
        case "format":
          if (redirectUrisRegexp.test(violation.location)) {
            redirectUrisErrorMessages.push(
              renderToString(
                "ModifyOAuthClientForm.redirect-uris.invalid-error"
              )
            );
          }
          if (postLogoutRedirectUrisRegexp.test(violation.location)) {
            postLogoutRedirectUrisErrorMessages.push(
              renderToString(
                "ModifyOAuthClientForm.post-logout-redirect-uris.invalid-error"
              )
            );
          }
          break;
        case "required":
          if (violation.missingField.includes("name")) {
            clientNameErrorMessages.push(
              renderToString("ModifyOAuthClientForm.name.required-error")
            );
          }
          break;
        default:
          break;
      }
    }

    return {
      clientName: defaultFormatErrorMessageList(clientNameErrorMessages),
      redirectUris: defaultFormatErrorMessageList(redirectUrisErrorMessages),
      postLogoutRedirectUris: defaultFormatErrorMessageList(
        postLogoutRedirectUrisErrorMessages
      ),
    };
  }, [updateAppConfigError, renderToString]);

  return (
    <section className={cn(styles.root, className)}>
      <TextField
        className={styles.inputField}
        label={renderToString("ModifyOAuthClientForm.name-label")}
        value={clientConfig.name ?? ""}
        onChange={onClientNameChange}
        required={true}
        errorMessage={errorMessageMap.clientName}
      />
      <TextField
        className={styles.inputField}
        label={renderToString(
          "ModifyOAuthClientForm.acces-token-lifetime-label"
        )}
        value={clientConfig.access_token_lifetime_seconds?.toString() ?? ""}
        onChange={onAccessTokenLifetimeChange}
      />
      <TextField
        className={styles.inputField}
        label={renderToString(
          "ModifyOAuthClientForm.refresh-token-lifetime-label"
        )}
        value={clientConfig.refresh_token_lifetime_seconds?.toString() ?? ""}
        onChange={onRefreshTokenLifetimeChange}
      />
      <LabelWithTooltip
        labelId="ModifyOAuthClientForm.redirect-uris-label"
        tooltipHeaderId="ModifyOAuthClientForm.redirect-uris-label"
        tooltipMessageId="ModifyOAuthClientForm.redirect-uris-tooltip-message"
        directionalHint={DirectionalHint.bottomLeftEdge}
        required={true}
      />
      <TagPicker
        className={styles.inputField}
        onChange={onRedirectUrisChange}
        onResolveSuggestions={onResolveRedirectUrisSuggestions}
        selectedItems={redirectUriItems}
      />
      <Text className={styles.errorMessage}>
        {errorMessageMap.redirectUris}
      </Text>
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
        selectedItems={postLogoutRedirectUriItems}
      />
      <Text className={styles.errorMessage}>
        {errorMessageMap.postLogoutRedirectUris}
      </Text>
    </section>
  );
};

export default ModifyOAuthClientForm;
